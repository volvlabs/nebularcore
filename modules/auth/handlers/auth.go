package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/jideobs/nebularcore/modules/auth/backends"
	"gitlab.com/jideobs/nebularcore/modules/auth/config"
	autherrors "gitlab.com/jideobs/nebularcore/modules/auth/errors"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authManager backends.AuthenticationManager
	tokenIssuer interfaces.TokenIssuer
	config      *config.Config
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(
	authManager backends.AuthenticationManager,
	tokenIssuer interfaces.TokenIssuer,
	config *config.Config,
) *AuthHandler {
	return &AuthHandler{
		authManager: authManager,
		tokenIssuer: tokenIssuer,
		config:      config,
	}
}

// RegisterRoutes registers the authentication routes
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/logout", h.Logout)
	}
}

// Login handles user login with various methods
func (h *AuthHandler) Login(c *gin.Context) {
	var credentials map[string]any

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authManager.Authenticate(c.Request.Context(), credentials)
	if err != nil {
		if authErr, ok := err.(*autherrors.AuthError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": authErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication failed"})
		return
	}

	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credentials"})
		return
	}

	// Issue JWT token
	tokenResp, err := h.tokenIssuer.IssueToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}

	c.JSON(http.StatusOK, tokenResp)
}

// RefreshToken handles token refresh requests
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenResp, err := h.tokenIssuer.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, tokenResp)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing token"})
		return
	}

	// Remove "Bearer " prefix
	token = token[7:]

	if err := h.tokenIssuer.RevokeToken(token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke token"})
		return
	}

	c.Status(http.StatusNoContent)
}
