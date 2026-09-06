package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/volvlabs/nebularcore/models/responses"
	"github.com/volvlabs/nebularcore/modules/auth/config"
	"github.com/volvlabs/nebularcore/modules/auth/interfaces"
	"github.com/volvlabs/nebularcore/modules/auth/models/requests"
	"github.com/volvlabs/nebularcore/modules/auth/password"
	"github.com/volvlabs/nebularcore/tools/handlers"
)

// PasswordHandler handles password-related HTTP requests
type PasswordHandler struct {
	passwordManager password.Manager
	authMiddleware  interfaces.AuthMiddleware
	config          *config.Config
}

// NewPasswordHandler creates a new password handler
func NewPasswordHandler(
	passwordManager password.Manager,
	authMiddleware interfaces.AuthMiddleware,
	config *config.Config,
) *PasswordHandler {
	return &PasswordHandler{
		passwordManager: passwordManager,
		authMiddleware:  authMiddleware,
		config:          config,
	}
}

// RegisterRoutes registers the password handler routes
func (h *PasswordHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/auth/password/reset-request", h.RequestPasswordReset)
	router.POST("/auth/password/reset-verify", h.VerifyPasswordReset)
	router.POST("/auth/password/change", h.authMiddleware.RequireAuth(), h.ChangePassword)
}

// @ID request-password-reset
// @Summary      Request Password Reset
// @Description  Requests a password reset
// @Tags         auth, default
// @Accept       json
// @Produce      json
// @Param 		 request body requests.PasswordResetPayload true "Password reset request"
// @Success      202  {object}  responses.ApiResponsePayload
// @Failure      400  {object}  handlers.ApiError
// @Failure      500  {object}  handlers.ApiError
// @Router       /auth/password/reset-request [post]
func (h *PasswordHandler) RequestPasswordReset(c *gin.Context) {
	var req requests.PasswordResetPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		handlers.NewBadRequestError(c, "bad request payload", nil)
		return
	}

	if err := h.passwordManager.RequestPasswordReset(c.Request.Context(), req); err != nil {
		handlers.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponsePayload{
		Message: "If the email exists, a reset link will be sent",
	})
}

// @ID verify-password-reset
// @Summary      Verify Password Reset
// @Description  Verifies a password reset
// @Tags         auth, default
// @Accept       json
// @Produce      json
// @Param 		 request body requests.PasswordResetVerifyPayload true "Password reset verify request"
// @Success      202  {object}  responses.ApiResponsePayload
// @Failure      400  {object}  handlers.ApiError
// @Failure      500  {object}  handlers.ApiError
// @Router       /auth/password/reset-verify [post]
func (h *PasswordHandler) VerifyPasswordReset(c *gin.Context) {
	var payload requests.PasswordResetVerifyPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		handlers.NewBadRequestError(c, "bad request payload", nil)
		return
	}

	if err := h.passwordManager.VerifyPasswordReset(c.Request.Context(), payload); err != nil {
		handlers.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponsePayload{
		Message: "Password has been reset successfully",
	})
}

// @ID change-password
// @Summary      Change Password
// @Description  Changes a user's password
// @Tags         auth, default
// @Accept       json
// @Produce      json
// @Param 		 request body requests.PasswordChangePayload true "Password change request"
// @Success      202  {object}  responses.ApiResponsePayload
// @Failure      401  {object}  handlers.ApiError
// @Failure      400  {object}  handlers.ApiError
// @Failure      500  {object}  handlers.ApiError
// @Security 	 BearerAuth
// @Router       /auth/password/change [post]
func (h *PasswordHandler) ChangePassword(c *gin.Context) {
	var payload requests.PasswordChangePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		handlers.NewBadRequestError(c, "bad request payload", nil)
		return
	}

	userInfo, exists := c.Get("user")
	if !exists {
		handlers.NewUnauthorizedError(c)
		return
	}

	if err := h.passwordManager.ChangePassword(c.Request.Context(), userInfo.(interfaces.User), payload); err != nil {
		handlers.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, responses.ApiResponsePayload{
		Message: "Password changed successfully",
	})
}
