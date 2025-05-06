package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gitlab.com/jideobs/nebularcore/modules/auth/backends"
	"gitlab.com/jideobs/nebularcore/modules/auth/config"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces"
	"gitlab.com/jideobs/nebularcore/modules/auth/models/requests"
	"golang.org/x/crypto/bcrypt"
)

// PasswordHandler handles password-related HTTP requests
type PasswordHandler struct {
	userRepo       interfaces.UserRepository
	authManager    backends.AuthenticationManager
	authMiddleware interfaces.AuthMiddleware
	config         *config.Config
}

// NewPasswordHandler creates a new password handler
func NewPasswordHandler(
	userRepo interfaces.UserRepository,
	authManager backends.AuthenticationManager,
	authMiddleware interfaces.AuthMiddleware,
	config *config.Config,
) *PasswordHandler {
	return &PasswordHandler{
		userRepo:       userRepo,
		authManager:    authManager,
		authMiddleware: authMiddleware,
		config:         config,
	}
}

// RegisterRoutes registers the password handler routes
func (h *PasswordHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/auth/password/reset-request", h.RequestPasswordReset)
	router.POST("/auth/password/reset-verify", h.VerifyPasswordReset)
	router.POST("/auth/password/change", h.authMiddleware.RequireAuth(), h.ChangePassword)
}

// RequestPasswordReset handles password reset requests
func (h *PasswordHandler) RequestPasswordReset(c *gin.Context) {
	var req requests.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	user, err := h.userRepo.FindByEmail(c.Request.Context(), req.Email)
	if err != nil {
		log.Error().Err(err).Msg("Error finding user by email")
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link will be sent"})
		return
	}

	if user == nil {
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link will be sent"})
		return
	}

	// Generate reset token
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		log.Error().Err(err).Msg("Error generating reset token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}
	resetToken := hex.EncodeToString(token)

	// Save reset token and expiry
	now := time.Now()
	user.SetPasswordResetToken(&resetToken)
	user.SetPasswordResetAt(&now)

	updates := map[string]any{
		"password_reset_token": resetToken,
		"password_reset_at":    now,
	}

	if err := h.userRepo.Update(c.Request.Context(), user, updates); err != nil {
		log.Error().Err(err).Msg("Error updating user with reset token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	// TODO: Send reset email with token
	// This should be implemented by the application using the framework

	c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link will be sent"})
}

// VerifyPasswordReset handles password reset verification
func (h *PasswordHandler) VerifyPasswordReset(c *gin.Context) {
	var req requests.PasswordResetVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by reset token
	user, err := h.userRepo.FindByResetToken(c.Request.Context(), req.Token)
	if err != nil {
		log.Error().Err(err).Msg("Error finding user by reset token")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
		return
	}

	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
		return
	}

	// Check if token is expired (24 hours)
	resetAt := user.GetPasswordResetAt()
	if resetAt == nil || time.Since(*resetAt) > 24*time.Hour {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reset token has expired"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Error hashing password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	// Update password and clear reset token
	user.SetPassword(string(hashedPassword))
	user.SetPasswordResetToken(nil)
	user.SetPasswordResetAt(nil)

	updates := map[string]any{
		"password":             string(hashedPassword),
		"password_reset_token": nil,
		"password_reset_at":    nil,
	}

	if err := h.userRepo.Update(c.Request.Context(), user, updates); err != nil {
		log.Error().Err(err).Msg("Error updating user password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password has been reset successfully"})
}

// ChangePassword handles password change requests for authenticated users
func (h *PasswordHandler) ChangePassword(c *gin.Context) {
	var req requests.PasswordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user from context
	userInfo, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userMap := userInfo.(map[string]any)
	userID, ok := userMap["sub"]
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Convert string ID to UUID
	uuid, err := uuid.Parse(userID.(string))
	if err != nil {
		log.Error().Err(err).Msg("Error parsing user ID")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	// Get user from database
	user, err := h.userRepo.FindByID(c.Request.Context(), uuid)
	if err != nil {
		log.Error().Err(err).Msg("Error finding user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.GetPasswordHash()), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Error hashing password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	// Update password
	user.SetPassword(string(hashedPassword))
	updates := map[string]any{
		"password": string(hashedPassword),
	}

	if err := h.userRepo.Update(c.Request.Context(), user, updates); err != nil {
		log.Error().Err(err).Msg("Error updating user password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
