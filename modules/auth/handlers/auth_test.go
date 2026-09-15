package handlers_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	backendsMocks "github.com/volvlabs/nebularcore/modules/auth/backends/mocks"
	"github.com/volvlabs/nebularcore/modules/auth/config"
	autherrors "github.com/volvlabs/nebularcore/modules/auth/errors"
	"github.com/volvlabs/nebularcore/modules/auth/handlers"
	"github.com/volvlabs/nebularcore/modules/auth/interfaces/mocks"
	"github.com/volvlabs/nebularcore/modules/auth/models/responses"
	"github.com/volvlabs/nebularcore/tools/test"
)

func setupAuthHandler(t *testing.T) (*handlers.AuthHandler, *backendsMocks.AuthenticationManager, *mocks.TokenIssuer, *mocks.GoogleSignin) {
	authManager := backendsMocks.NewAuthenticationManager(t)
	tokenIssuer := mocks.NewTokenIssuer(t)
	googleSignin := mocks.NewGoogleSignin(t)
	cfg := &config.Config{
		Social: config.SocialConfig{
			Enabled: true,
		},
	}

	return handlers.NewAuthHandler(
		authManager,
		tokenIssuer,
		googleSignin,
		cfg,
	), authManager, tokenIssuer, googleSignin
}

func TestAuthHandler_InitiateSocialLogin(t *testing.T) {
	tests := []test.ApiScenario{
		{
			Name:            "successful social login",
			Url:             "/api/v1/auth/providers/google/initiate",
			Method:          "GET",
			ExpectedStatus:  200,
			ExpectedContent: []string{"url", "https://example.com/auth/google"},
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				handler, _, _, googleSignin := setupAuthHandler(t)
				googleSignin.On("GetAuthURL", mock.Anything).Return("https://example.com/auth/google")

				api := router.Group("/api/v1")
				handler.RegisterRoutes(api, func(c *gin.Context) {})
			},
		},
		{
			Name:            "invalid provider",
			Url:             "/api/v1/auth/providers/invalid/initiate",
			Method:          "GET",
			ExpectedStatus:  400,
			ExpectedContent: []string{"invalid provider"},
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				handler, _, _, _ := setupAuthHandler(t)

				api := router.Group("/api/v1")
				handler.RegisterRoutes(api, func(c *gin.Context) {})
			},
		},
		{
			Name:            "provider not supported",
			Url:             "/api/v1/auth/providers/apple/initiate",
			Method:          "GET",
			ExpectedStatus:  400,
			ExpectedContent: []string{"provider not supported"},
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				handler, _, _, _ := setupAuthHandler(t)

				api := router.Group("/api/v1")
				handler.RegisterRoutes(api, func(c *gin.Context) {})
			},
		},
	}

	for _, tt := range tests {
		tt.Test(t)
	}
}

func TestAuthHandler_SocialSignInOrSignUp(t *testing.T) {
	tests := []test.ApiScenario{
		{
			Name:   "should successfully redirect to signup",
			Url:    "/api/v1/auth/providers/google/login",
			Method: "POST",
			Body: strings.NewReader(`{
				"idToken": "idToken"
			}`),
			ExpectedStatus:  http.StatusPermanentRedirect,
			ExpectedContent: []string{},
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				handler, authManager, _, _ := setupAuthHandler(t)

				authManager.On("Authenticate",
					mock.Anything,
					map[string]any{"idToken": "idToken"},
				).Return(nil, autherrors.ErrSocialEmailDoesNotExist)

				api := router.Group("/api/v1")
				handler.RegisterRoutes(api, func(c *gin.Context) {
					c.JSON(http.StatusCreated, map[string]string{
						"message": "account created successfully",
					})
				})
			},
		},
		{
			Name:   "should successfully login",
			Url:    "/api/v1/auth/providers/google/login",
			Method: "POST",
			Body: strings.NewReader(`{
					"idToken": "idToken"
				}`),
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{},
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				handler, authManager, tokenIssuer, _ := setupAuthHandler(t)

				user := mocks.NewUser(t)
				authManager.On("Authenticate",
					mock.Anything,
					map[string]any{
						"idToken": "idToken",
					},
				).Return(user, nil)
				tokenIssuer.On("IssueToken", user).Return(&responses.TokenResponse{
					AccessToken:  "test-token",
					RefreshToken: "test-refresh",
				}, nil)

				api := router.Group("/api/v1")
				handler.RegisterRoutes(api, func(c *gin.Context) {
					c.JSON(http.StatusCreated, map[string]string{
						"message": "account created successfully",
					})
				})
			},
		},
		{
			Name:   "should return error since login failed with existing social account",
			Url:    "/api/v1/auth/providers/google/login",
			Method: "POST",
			Body: strings.NewReader(`{
					"idToken": "idToken"
				}`),
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedContent: []string{},
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				handler, authManager, _, _ := setupAuthHandler(t)

				authManager.On("Authenticate",
					mock.Anything,
					map[string]any{
						"idToken": "idToken",
					},
				).Return(nil, autherrors.ErrInvalidCredentials)

				api := router.Group("/api/v1")
				handler.RegisterRoutes(api, func(c *gin.Context) {
					c.JSON(http.StatusCreated, map[string]string{
						"message": "account created successfully",
					})
				})
			},
		},
	}

	for _, tt := range tests {
		tt.Test(t)
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []test.ApiScenario{
		{
			Name:   "successful login",
			Url:    "/api/v1/auth/login",
			Method: "POST",
			Body: strings.NewReader(`{
				"username": "testuser",
				"password": "testpass"
			}`),
			ExpectedStatus:  200,
			ExpectedContent: []string{"accessToken", "refreshToken"},
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				handler, authManager, tokenIssuer, _ := setupAuthHandler(t)

				user := mocks.NewUser(t)
				authManager.On("Authenticate", mock.Anything, mock.Anything).Return(user, nil)
				tokenIssuer.On("IssueToken", user).Return(&responses.TokenResponse{
					AccessToken:  "test-token",
					RefreshToken: "test-refresh",
				}, nil)

				api := router.Group("/api/v1")
				handler.RegisterRoutes(api, nil)
			},
		},
		{
			Name:   "invalid credentials",
			Url:    "/api/v1/auth/login",
			Method: "POST",
			Body: strings.NewReader(`{
				"username": "invalid",
				"password": "invalid"
			}`),
			ExpectedStatus:  400,
			ExpectedContent: []string{"invalid credentials"},
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				handler, authManager, _, _ := setupAuthHandler(t)
				authManager.On("Authenticate", mock.Anything, mock.Anything).Return(nil, nil)

				api := router.Group("/api/v1")
				handler.RegisterRoutes(api, nil)
			},
		},
	}

	for _, tt := range tests {
		tt.Test(t)
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	tests := []test.ApiScenario{
		{
			Name:   "successful token refresh",
			Url:    "/api/v1/auth/refresh",
			Method: "POST",
			Body: strings.NewReader(`{
				"refreshToken": "valid-refresh-token"
			}`),
			ExpectedStatus:  200,
			ExpectedContent: []string{"accessToken", "refreshToken"},
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				handler, _, tokenIssuer, _ := setupAuthHandler(t)

				tokenIssuer.On("RefreshToken", "valid-refresh-token").Return(&responses.TokenResponse{
					AccessToken:  "new-token",
					RefreshToken: "new-refresh",
				}, nil)

				api := router.Group("/api/v1")
				handler.RegisterRoutes(api, nil)
			},
		},
		{
			Name:   "invalid refresh token",
			Url:    "/api/v1/auth/refresh",
			Method: "POST",
			Body: strings.NewReader(`{
				"refreshToken": "invalid-token"
			}`),
			ExpectedStatus:  400,
			ExpectedContent: []string{"invalid refresh token"},
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				handler, _, tokenIssuer, _ := setupAuthHandler(t)
				tokenIssuer.On("RefreshToken", "invalid-token").Return(nil, autherrors.ErrInvalidToken)

				api := router.Group("/api/v1")
				handler.RegisterRoutes(api, nil)
			},
		},
	}

	for _, tt := range tests {
		tt.Test(t)
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	tests := []test.ApiScenario{
		{
			Name:   "successful logout",
			Url:    "/api/v1/auth/logout",
			Method: "POST",
			RequestHeaders: map[string]string{
				"Authorization": "Bearer valid-token",
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"Successfully logged out"},
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				handler, _, tokenIssuer, _ := setupAuthHandler(t)
				tokenIssuer.On("RevokeToken", "valid-token").Return(nil)

				api := router.Group("/api/v1")
				handler.RegisterRoutes(api, nil)
			},
		},
		{
			Name:            "missing token",
			Url:             "/api/v1/auth/logout",
			Method:          "POST",
			ExpectedStatus:  400,
			ExpectedContent: []string{"missing token"},
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				handler, _, _, _ := setupAuthHandler(t)

				api := router.Group("/api/v1")
				handler.RegisterRoutes(api, nil)
			},
		},
	}

	for _, tt := range tests {
		tt.Test(t)
	}
}
