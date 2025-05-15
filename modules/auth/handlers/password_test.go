package handlers_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	backendMocks "gitlab.com/jideobs/nebularcore/modules/auth/backends/mocks"
	"gitlab.com/jideobs/nebularcore/modules/auth/config"
	"gitlab.com/jideobs/nebularcore/modules/auth/handlers"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces/mocks"
	"gitlab.com/jideobs/nebularcore/tools/test"
)

func TestPasswordHandlerScenarios(t *testing.T) {
	userRepo := mocks.NewUserRepository(t)
	authManager := backendMocks.NewAuthenticationManager(t)
	authMiddleware := mocks.NewAuthMiddleware(t)
	config := &config.Config{}
	handler := handlers.NewPasswordHandler(userRepo, authManager, authMiddleware, config)

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
				userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(mockUser, nil)
				userRepo.On("Update", mock.Anything, mockUser, mock.Anything).Return(nil)
				mockUser.On("SetPasswordResetToken", mock.Anything).Return(nil)
				mockUser.On("SetPasswordResetAt", mock.AnythingOfType("*time.Time")).Return(nil)
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
				mockUser.On("SetPasswordResetToken", mock.AnythingOfType("*string")).Return(nil)
				mockUser.On("SetPasswordResetAt", mock.AnythingOfType("*time.Time")).Return(nil)
				now := time.Now()
				mockUser.On("GetPasswordResetAt").Return(&now)
				mockUser.On("SetPassword", mock.AnythingOfType("string")).Return(nil)
				userRepo.On("FindByResetToken", mock.Anything, "valid_token").Return(mockUser, nil)
				userRepo.On("Update", mock.Anything, mockUser, mock.Anything).Return(nil)

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
				userID := uuid.New()
				mockUser := mocks.NewUser(t)
				mockUser.On("GetID").Return(userID)
				mockUser.On("GetPasswordHash").Return("$2y$10$rdZHiX3UnEJ8LPTy5VzI9OHl6MdX5E.PYnjkSM25xZ9aFMBt5Qu.e")
				mockUser.On("SetPassword", mock.AnythingOfType("string")).Return(nil)

				authMiddleware.On("RequireAuth").Return(func() gin.HandlerFunc {
					return func(c *gin.Context) {
						c.Set("user", mockUser)
						c.Next()
					}
				})

				userRepo.On("FindByID", mock.Anything, userID).Return(mockUser, nil)
				userRepo.On("Update", mock.Anything, mockUser, mock.Anything).Return(nil)

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
			userRepo.ExpectedCalls = nil
			userRepo.Calls = nil
			authMiddleware.ExpectedCalls = nil
			authMiddleware.Calls = nil

			scenario.Test(t)
		})
	}
}
