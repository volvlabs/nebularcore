package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/volvlabs/nebularcore/modules/auth/authorization"
	"github.com/volvlabs/nebularcore/modules/auth/backends"
	"github.com/volvlabs/nebularcore/modules/auth/config"
	"github.com/volvlabs/nebularcore/modules/auth/interfaces"
)

// AuthMiddleware provides authentication middleware for Gin
type AuthMiddleware struct {
	authManager backends.AuthenticationManager
	authMgr     *authorization.AuthorizationManager
	config      *config.AuthorizationConfig
}

// NewAuthMiddleware creates a new authentication middleware. authMgr is the
// shared AuthorizationManager (see authorization.NewAuthorizationManager) —
// this middleware no longer constructs its own casbin enforcer, it enforces
// through the same one AuthorizationManager owns.
func NewAuthMiddleware(
	authManager backends.AuthenticationManager,
	authMgr *authorization.AuthorizationManager,
	cfg *config.AuthorizationConfig,
) *AuthMiddleware {
	return &AuthMiddleware{
		authManager: authManager,
		authMgr:     authMgr,
		config:      cfg,
	}
}

// JWT returns middleware for JWT authentication
func (m *AuthMiddleware) JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization token",
			})
			return
		}

		user, err := m.authManager.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		// Store user in context
		c.Set("user", user)
		c.Next()
	}
}

// APIKey returns middleware for API key authentication
func (m *AuthMiddleware) APIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		apiSecret := c.GetHeader("X-API-Secret")

		if apiKey == "" || apiSecret == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing API credentials",
			})
			return
		}

		user, err := m.authManager.Authenticate(c.Request.Context(), map[string]any{
			"api_key":    apiKey,
			"api_secret": apiSecret,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid API credentials",
			})
			return
		}

		// Store user in context
		c.Set("user", user)
		c.Next()
	}
}

// Optional returns middleware that attempts authentication but doesn't require it
func (m *AuthMiddleware) Optional() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try JWT first
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token != "" {
			if user, err := m.authManager.ValidateToken(c.Request.Context(), token); err == nil {
				c.Set("user", user)
				c.Next()
				return
			}
		}

		// Try API key
		apiKey := c.GetHeader("X-API-Key")
		apiSecret := c.GetHeader("X-API-Secret")
		if apiKey != "" && apiSecret != "" {
			if user, err := m.authManager.Authenticate(c.Request.Context(), map[string]interface{}{
				"api_key":    apiKey,
				"api_secret": apiSecret,
			}); err == nil {
				c.Set("user", user)
				c.Next()
				return
			}
		}

		// No valid authentication found, but that's ok
		c.Next()
	}
}

// RequireAuth returns middleware that requires authentication using any method
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token != "" {
			if user, err := m.authManager.ValidateToken(c.Request.Context(), token); err == nil {
				c.Set("user", user)
				c.Next()
				return
			}
		}

		apiKey := c.GetHeader("X-API-Key")
		apiSecret := c.GetHeader("X-API-Secret")
		if apiKey != "" && apiSecret != "" {
			if user, err := m.authManager.Authenticate(c.Request.Context(), map[string]interface{}{
				"api_key":    apiKey,
				"api_secret": apiSecret,
			}); err == nil {
				c.Set("user", user)
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
	}
}

// RequireRole returns middleware that requires a specific role
func (m *AuthMiddleware) RequireRole(role string) gin.HandlerFunc {
	if !m.config.MiddlewareEnabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			return
		}

		allowed, err := m.authMgr.Enforcer().Enforce(user.(interfaces.User).GetRole(), c.Request.URL.Path, "*")
		if err != nil || !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			return
		}

		c.Next()
	}
}

// RequirePermission returns middleware that requires a specific permission
func (m *AuthMiddleware) RequirePermission(resource, action string) gin.HandlerFunc {
	if !m.config.MiddlewareEnabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			return
		}

		allowed, err := m.authMgr.Enforcer().Enforce(user.(interfaces.User).GetID().String(), resource, action)
		if err != nil || !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			return
		}

		c.Next()
	}
}

// extractBearerToken extracts the token from the Authorization header
func extractBearerToken(header string) string {
	if header == "" {
		return ""
	}

	parts := strings.Split(header, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}
