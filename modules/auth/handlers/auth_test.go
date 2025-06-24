package handlers_test

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	backendsMocks "gitlab.com/jideobs/nebularcore/modules/auth/backends/mocks"
	"gitlab.com/jideobs/nebularcore/modules/auth/config"
	autherrors "gitlab.com/jideobs/nebularcore/modules/auth/errors"
	"gitlab.com/jideobs/nebularcore/modules/auth/handlers"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces/mocks"
	"gitlab.com/jideobs/nebularcore/modules/auth/models/responses"
	"gitlab.com/jideobs/nebularcore/tools/test"
)

func setupAuthHandler(t *testing.T) (*handlers.AuthHandler, *backendsMocks.AuthenticationManager, *mocks.TokenIssuer) {
	authManager := backendsMocks.NewAuthenticationManager(t)
	tokenIssuer := mocks.NewTokenIssuer(t)
	cfg := &config.Config{}

	return handlers.NewAuthHandler(authManager, tokenIssuer, cfg), authManager, tokenIssuer
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
			ExpectedContent: []string{"accessToken", "test-token"},
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				handler, authManager, tokenIssuer := setupAuthHandler(t)

				user := mocks.NewUser(t)
				authManager.On("Authenticate", mock.Anything, mock.Anything).Return(user, nil)
				tokenIssuer.On("IssueToken", user).Return(&responses.TokenResponse{
					AccessToken:  "test-token",
					RefreshToken: "test-refresh",
				}, nil)

				api := router.Group("/api/v1")
				handler.RegisterRoutes(api)
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
				handler, authManager, _ := setupAuthHandler(t)
				authManager.On("Authenticate", mock.Anything, mock.Anything).Return(nil, nil)

				api := router.Group("/api/v1")
				handler.RegisterRoutes(api)
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
			ExpectedContent: []string{"test-token", "test-refresh"},
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				handler, _, tokenIssuer := setupAuthHandler(t)

				tokenIssuer.On("RefreshToken", "valid-refresh-token").Return(&responses.TokenResponse{
					AccessToken:  "test-token",
					RefreshToken: "test-refresh",
				}, nil)

				api := router.Group("/api/v1")
				handler.RegisterRoutes(api)
			},
		},
		{
			Name:   "invalid refresh token",
			Url:    "/api/v1/auth/refresh",
			Method: "POST",
			Body: strings.NewReader(`{
				"refreshToken": "invalid-token"
			}`),
			ExpectedStatus:  401,
			ExpectedContent: []string{"invalid refresh token"},
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				handler, _, tokenIssuer := setupAuthHandler(t)
				tokenIssuer.On("RefreshToken", "invalid-token").Return(nil, autherrors.ErrInvalidToken)

				api := router.Group("/api/v1")
				handler.RegisterRoutes(api)
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
			ExpectedStatus: 204,
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				handler, _, tokenIssuer := setupAuthHandler(t)
				tokenIssuer.On("RevokeToken", "valid-token").Return(nil)

				api := router.Group("/api/v1")
				handler.RegisterRoutes(api)
			},
		},
		{
			Name:            "missing token",
			Url:             "/api/v1/auth/logout",
			Method:          "POST",
			ExpectedStatus:  400,
			ExpectedContent: []string{"missing token"},
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				handler, _, _ := setupAuthHandler(t)

				api := router.Group("/api/v1")
				handler.RegisterRoutes(api)
			},
		},
	}

	for _, tt := range tests {
		tt.Test(t)
	}
}
