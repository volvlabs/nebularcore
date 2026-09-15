package handlers_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/volvlabs/nebularcore/modules/auth/config"
	"github.com/volvlabs/nebularcore/modules/auth/handlers"
	"github.com/volvlabs/nebularcore/modules/auth/interfaces/mocks"
	"github.com/volvlabs/nebularcore/modules/auth/models/requests"
	passwordMocks "github.com/volvlabs/nebularcore/modules/auth/password/mocks"
	"github.com/volvlabs/nebularcore/tools/test"
)

func TestPasswordHandlerScenarios(t *testing.T) {
	passwordManager := passwordMocks.NewManager(t)
	authMiddleware := mocks.NewAuthMiddleware(t)
	config := &config.Config{}
	handler := handlers.NewPasswordHandler(passwordManager, authMiddleware, config)

	scenarios := []test.ApiScenario{
		{
			Name:   "Request Password Reset - Valid Email",
			Method: http.MethodPost,
			Url:    "/api/v1/auth/password/reset-request",
			Body: strings.NewReader(`{
				"email": "test@example.com"
			}`),
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				mockUser := mocks.NewUser(t)
				passwordManager.On("RequestPasswordReset", mock.Anything, mock.MatchedBy(func(payload requests.PasswordResetPayload) bool {
					return payload.Email == "test@example.com"
				})).Return(nil)
				authMiddleware.On("RequireAuth").Return(func() gin.HandlerFunc {
					return func(c *gin.Context) {
						c.Set("user", mockUser)
						c.Next()
					}
				})

				apiGroup := router.Group("/api/v1")
				handler.RegisterRoutes(apiGroup)
			},
			ExpectedStatus: http.StatusOK,
			ExpectedContent: []string{
				"If the email exists, a reset link will be sent",
			},
		},
		{
			Name:   "Request Password Reset - Invalid Email",
			Method: http.MethodPost,
			Url:    "/api/v1/auth/password/reset-request",
			Body: strings.NewReader(`{
				"email": "invalid"
			}`),
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				authMiddleware.On("RequireAuth").Return(func() gin.HandlerFunc {
					return func(c *gin.Context) {
						mockUser := mocks.NewUser(t)
						c.Set("user", mockUser)
						c.Next()
					}
				})

				apiGroup := router.Group("/api/v1")
				handler.RegisterRoutes(apiGroup)
			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:   "Verify Password Reset - Valid Token",
			Method: http.MethodPost,
			Url:    "/api/v1/auth/password/reset-verify",
			Body: strings.NewReader(`{
				"token": "valid_token",
				"password": "new_password123"
			}`),
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				mockUser := mocks.NewUser(t)
				passwordManager.On("VerifyPasswordReset", mock.Anything, mock.MatchedBy(func(payload requests.PasswordResetVerifyPayload) bool {
					return payload.Token == "valid_token" && payload.Password == "new_password123"
				})).Return(nil)
				authMiddleware.On("RequireAuth").Return(func() gin.HandlerFunc {
					return func(c *gin.Context) {
						c.Set("user", mockUser)
						c.Next()
					}
				})

				apiGroup := router.Group("/api/v1")
				handler.RegisterRoutes(apiGroup)
			},
			ExpectedStatus: http.StatusOK,
			ExpectedContent: []string{
				"Password has been reset successfully",
			},
		},
		{
			Name:   "Change Password - Valid Request",
			Method: http.MethodPost,
			Url:    "/api/v1/auth/password/change",
			Body: strings.NewReader(`{
				"currentPassword": "old_password",
				"newPassword": "new_password123"
			}`),
			RequestHeaders: map[string]string{
				"Authorization": "Bearer valid_token",
			},
			BeforeTestFunc: func(t *testing.T, router *gin.Engine) {
				mockUser := mocks.NewUser(t)
				passwordManager.On("ChangePassword",
					mock.Anything,
					mockUser,
					mock.MatchedBy(func(payload requests.PasswordChangePayload) bool {
						return payload.CurrentPassword == "old_password" && payload.NewPassword == "new_password123"
					})).Return(nil)

				authMiddleware.On("RequireAuth").Return(func() gin.HandlerFunc {
					return func(c *gin.Context) {
						c.Set("user", mockUser)
						c.Next()
					}
				})

				apiGroup := router.Group("/api/v1")
				handler.RegisterRoutes(apiGroup)
			},
			ExpectedStatus: http.StatusOK,
			ExpectedContent: []string{
				"Password changed successfully",
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			passwordManager.ExpectedCalls = nil
			passwordManager.Calls = nil
			authMiddleware.ExpectedCalls = nil
			authMiddleware.Calls = nil

			scenario.Test(t)
		})
	}
}
