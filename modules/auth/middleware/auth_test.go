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

	authorizationMocks "gitlab.com/jideobs/nebularcore/modules/auth/authorization/mocks"
	backendMocks "gitlab.com/jideobs/nebularcore/modules/auth/backends/mocks"
	"gitlab.com/jideobs/nebularcore/modules/auth/config"
	"gitlab.com/jideobs/nebularcore/modules/auth/middleware"
	"gitlab.com/jideobs/nebularcore/modules/auth/models"
)

func setupTest(t *testing.T) (*gin.Engine, *middleware.AuthMiddleware, *backendMocks.AuthenticationManager, *authorizationMocks.Manager) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	authManager := backendMocks.NewAuthenticationManager(t)
	authzManager := authorizationMocks.NewManager(t)

	cfg := &config.MiddlewareConfig{
		AuthorizationEnabled: true,
	}

	authMiddleware, err := middleware.NewAuthMiddleware(authManager, authzManager, cfg)
	assert.NoError(t, err)

	return r, authMiddleware, authManager, authzManager
}

func TestNewAuthMiddleware(t *testing.T) {
	authManager := backendMocks.NewAuthenticationManager(t)
	authzManager := authorizationMocks.NewManager(t)

	middlewareConfig := &config.MiddlewareConfig{
		AuthorizationEnabled: true,
	}

	middleware, err := middleware.NewAuthMiddleware(authManager, authzManager, middlewareConfig)
	assert.NoError(t, err)
	assert.NotNil(t, middleware)
}

func TestJWTMiddleware(t *testing.T) {
	r, authMiddleware, authManager, _ := setupTest(t)

	r.GET("/test", authMiddleware.JWT(), func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	tests := []struct {
		name      string
		token     string
		setupMock func(*backendMocks.AuthenticationManager)
		wantCode  int
	}{
		{
			name:  "missing token",
			token: "",
			setupMock: func(m *backendMocks.AuthenticationManager) {
				// No setup needed
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:  "invalid token",
			token: "invalid.jwt.token",
			setupMock: func(m *backendMocks.AuthenticationManager) {
				m.On("ValidateToken", mock.Anything, "invalid.jwt.token").
					Return(nil, errors.New("invalid token"))
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:  "valid token",
			token: "valid.jwt.token",
			setupMock: func(m *backendMocks.AuthenticationManager) {
				m.On("ValidateToken", mock.Anything, "valid.jwt.token").
					Return(&models.User{}, nil)
			},
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt.setupMock(authManager)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		if tt.token != "" {
			req.Header.Set("Authorization", "Bearer "+tt.token)
		}

		r.ServeHTTP(w, req)
		assert.Equal(t, tt.wantCode, w.Code)
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
				m.On("Authenticate", mock.Anything, map[string]interface{}{
					"api_key":    "valid-key",
					"api_secret": "valid-secret",
				}).Return(&models.User{
					ID:       uuid.New(),
					Email:    "user123@example.com",
					Password: "password",
				}, nil)
			},
			expectedResponse: []string{"status", "success"},
			wantCode:         http.StatusOK,
		},
		{
			name:      "invalid credentials",
			apiKey:    "invalid-key",
			apiSecret: "invalid-secret",
			setupMock: func(m *backendMocks.AuthenticationManager) {
				m.On("Authenticate", mock.Anything, map[string]interface{}{
					"api_key":    "invalid-key",
					"api_secret": "invalid-secret",
				}).Return(nil, errors.New("invalid credentials"))
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
			r, authMiddleware, authManager, _ := setupTest(t)

			tt.setupMock(authManager)

			r.GET("/test", authMiddleware.APIKey(), func(c *gin.Context) {
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
			r, authMiddleware, authManager, authorizationManager := setupTest(t)

			r.GET("/test", authMiddleware.JWT(), authMiddleware.RequireRole(tt.role), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			user := &models.User{
				ID:       uuid.New(),
				Email:    "user123@example.com",
				Password: "password",
			}
			authManager.On("ValidateToken", mock.Anything, "valid.jwt.token").Return(user, nil)
			authorizationManager.On("HasRole", mock.Anything, user.ID, tt.role).Return(tt.role == "admin", nil)

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
			r, authMiddleware, authManager, authzManager := setupTest(t)

			r.GET("/test", authMiddleware.JWT(), authMiddleware.RequirePermission(tt.resource, tt.action), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "success"})
			})

			user := &models.User{
				ID:       uuid.New(),
				Email:    "user123@example.com",
				Password: "password",
			}
			authManager.On("ValidateToken", mock.Anything, "valid.jwt.token").Return(user, nil)
			authzManager.On("HasPermission", mock.Anything, user.ID, tt.resource, tt.action).Return(tt.action == "GET", nil)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer valid.jwt.token")

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}
