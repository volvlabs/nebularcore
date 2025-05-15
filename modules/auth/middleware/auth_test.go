package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	backendMocks "gitlab.com/jideobs/nebularcore/modules/auth/backends/mocks"
	"gitlab.com/jideobs/nebularcore/modules/auth/config"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces/mocks"
	"gitlab.com/jideobs/nebularcore/modules/auth/middleware"
)

func setupTest(t *testing.T) (*gin.Engine, *middleware.AuthMiddleware, *backendMocks.AuthenticationManager) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	authManager := backendMocks.NewAuthenticationManager(t)

	cfg := &config.MiddlewareConfig{
		AuthorizationEnabled: true,
		PermissionModelPath:  "test-data/test-model.conf",
		PermissionPolicyPath: "test-data/test-policy.csv",
	}

	authMiddleware, err := middleware.NewAuthMiddleware(authManager, cfg)
	assert.NoError(t, err)

	return r, authMiddleware, authManager
}

func TestJWTMiddleware(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		setupMock func(*backendMocks.AuthenticationManager)
		wantCode  int
	}{
		{
			name:  "valid token",
			token: "valid.jwt.token",
			setupMock: func(m *backendMocks.AuthenticationManager) {
				mockUser := mocks.NewUser(t)
				m.On("ValidateToken", mock.Anything, "valid.jwt.token").Return(mockUser, nil)
			},
			wantCode: http.StatusOK,
		},
		{
			name:  "invalid token",
			token: "invalid.jwt.token",
			setupMock: func(m *backendMocks.AuthenticationManager) {
				m.On("ValidateToken", mock.Anything, "invalid.jwt.token").Return(nil, errors.New("invalid token"))
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:      "missing token",
			token:     "",
			setupMock: func(m *backendMocks.AuthenticationManager) {},
			wantCode:  http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r, authMiddleware, authManager := setupTest(t)
			tt.setupMock(authManager)

			r.Use(authMiddleware.JWT())
			r.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestAPIKeyMiddleware(t *testing.T) {
	tests := []struct {
		name             string
		apiKey           string
		apiSecret        string
		setupMock        func(*backendMocks.AuthenticationManager)
		expectedResponse []string
		wantCode         int
	}{
		{
			name:      "valid credentials",
			apiKey:    "valid-key",
			apiSecret: "valid-secret",
			setupMock: func(m *backendMocks.AuthenticationManager) {
				mockUser := mocks.NewUser(t)
				m.On("Authenticate", mock.Anything, mock.Anything).Return(mockUser, nil)
			},
			expectedResponse: []string{"status", "success"},
			wantCode:         http.StatusOK,
		},
		{
			name:      "invalid credentials",
			apiKey:    "invalid-key",
			apiSecret: "invalid-secret",
			setupMock: func(m *backendMocks.AuthenticationManager) {
				m.On("Authenticate", mock.Anything, mock.Anything).Return(nil, errors.New("invalid credentials"))
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:      "missing credentials",
			apiKey:    "",
			apiSecret: "",
			setupMock: func(m *backendMocks.AuthenticationManager) {},
			wantCode:  http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r, authMiddleware, authManager := setupTest(t)
			tt.setupMock(authManager)

			r.Use(authMiddleware.APIKey())
			r.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}
			if tt.apiSecret != "" {
				req.Header.Set("X-API-Secret", tt.apiSecret)
			}

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			for _, key := range tt.expectedResponse {
				assert.Contains(t, w.Body.String(), key)
			}
		})
	}
}

func TestAuthMiddleware_RequireRole(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		wantCode int
	}{
		{
			name:     "valid role",
			role:     "admin",
			wantCode: http.StatusOK,
		},
		{
			name:     "invalid role",
			role:     "user",
			wantCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, authMiddleware, authManager := setupTest(t)

			r.Use(authMiddleware.JWT())
			r.Use(authMiddleware.RequireRole(tt.role))
			r.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			mockUser := mocks.NewUser(t)
			mockUser.On("GetRole").Return(tt.role)
			authManager.On("ValidateToken", mock.Anything, "valid.jwt.token").
				Return(mockUser, nil)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer valid.jwt.token")

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestRequirePermissionMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		action   string
		wantCode int
	}{
		{
			name:     "has required permission",
			resource: "/test",
			action:   "GET",
			wantCode: http.StatusOK,
		},
		{
			name:     "missing required permission",
			resource: "/test",
			action:   "POST",
			wantCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r, authMiddleware, authManager := setupTest(t)

			r.Use(authMiddleware.JWT())
			r.Use(authMiddleware.RequirePermission(tt.resource, tt.action))
			r.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			userId := "98d99bae-7daa-430b-a0f9-a8e79b3ac1f6"
			mockUser := mocks.NewUser(t)
			mockUser.On("GetID").Return(uuid.MustParse(userId))
			authManager.On("ValidateToken", mock.Anything, "valid.jwt.token").Return(mockUser, nil)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer valid.jwt.token")

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}
