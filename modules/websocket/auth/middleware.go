package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// ContextKeyClaims is the gin context key for storing validated claims.
	ContextKeyClaims = "ws_claims"
)

// Middleware returns a Gin handler that validates JWT tokens for WebSocket
// upgrade requests. It checks the Authorization header first, then falls back
// to the "token" query parameter. Verification is HS256-only, via
// ValidateToken(token, jwtSecret) — hosts whose own tokens are signed
// RS256 (or need any other verification scheme) should use
// MiddlewareWithResolver instead.
func Middleware(jwtSecret string, authRequired bool) gin.HandlerFunc {
	return MiddlewareWithResolver(func(token string) (*Claims, error) {
		return ValidateToken(token, jwtSecret)
	}, authRequired)
}

// ClaimsResolverFunc validates a raw bearer/query token and returns its
// claims — a pluggable alternative to Middleware's fixed HS256-shared-
// secret verification, for hosts (e.g. mori-backend, RS256) whose own
// TokenIssuer already knows how to validate its own tokens.
type ClaimsResolverFunc func(token string) (*Claims, error)

// MiddlewareWithResolver is Middleware, generalized to accept any
// ClaimsResolverFunc instead of assuming HS256 + a shared secret.
func MiddlewareWithResolver(resolve ClaimsResolverFunc, authRequired bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			if authRequired {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"status":  false,
					"message": "authentication required",
				})
				return
			}
			c.Next()
			return
		}

		claims, err := resolve(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "invalid token",
			})
			return
		}

		c.Set(ContextKeyClaims, claims)
		c.Next()
	}
}

// extractBearerToken extracts the token from a "Bearer <token>" header value.
func extractBearerToken(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return parts[1]
}

// GetClaims extracts the validated claims from the gin context.
func GetClaims(c *gin.Context) *Claims {
	v, ok := c.Get(ContextKeyClaims)
	if !ok {
		return nil
	}
	claims, _ := v.(*Claims)
	return claims
}
