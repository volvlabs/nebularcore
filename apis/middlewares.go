package apis

import (
	"context"
	"strings"
	"time"

	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/tools"
	"gitlab.com/jideobs/nebularcore/tools/security"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
)

func AuthenticateRequestThenLoadAuthContext(app core.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			log.Info().Msg("Auth: no request authorization")
			NewUnauthorizedError(c)
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")

		claims, err := security.ParseJWT(token, app.Settings().AuthTokenSecret)
		if err != nil {
			log.Err(err).Msgf("Auth: error parsing JWT Token")
			NewUnauthorizedError(c)
			return
		}

		c.Set(tools.ContextClaimsKey, claims)
		c.Next()
	}
}

func AuthorizeRequest(app core.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, _ := c.Get(tools.ContextClaimsKey)
		claims := value.(jwt.MapClaims)
		subjectRole := cast.ToString(claims["role"])
		if !app.IsACLEnforced() {
			c.Next()
		}

		if subjectRole == "" {
			NewForbiddenError(c)
			return
		}

		method := c.Request.Method
		resource := c.Request.URL.Path

		isAuthorized, err := app.Acm().IsAuthroized(subjectRole, resource, method)
		if err != nil {
			log.Err(err).Msgf("error occurred trying to authorize")
			NewInternalServerError(c)

			return
		}

		if !isAuthorized {
			NewForbiddenError(c)
			return
		}

		c.Next()
	}
}

func TenantMiddleware(app core.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantId := c.GetHeader("X-Tenant-ID")
		if tenantId == "" {
			log.Warn().Msgf("Auth: access without tenant ID")
			NewUnauthorizedError(c)
			return
		}

		schemaName := app.SchemaName(tenantId)

		ctx := context.Background()
		ctx = context.WithValue(ctx, tools.ContextTenantIdKey, tenantId)
		ctx = context.WithValue(ctx, tools.ContextTenantSchemaNameKey, schemaName)

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func RequestLogger(app core.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Stop timer
		latency := time.Since(start)

		// Get request ID if exists
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Get client IP
		clientIP := c.ClientIP()

		// Get error if exists
		var errorMessage string
		if len(c.Errors) > 0 {
			errorMessage = c.Errors.String()
		}

		// Get request method and path with query parameters
		if raw != "" {
			path = path + "?" + raw
		}

		// Create log event
		event := log.Info()
		if c.Writer.Status() >= 400 {
			event = log.Error()
		}

		// Add common fields
		event.Str("request_id", requestID).
			Str("client_ip", clientIP).
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Dur("latency", latency).
			Int("body_size", c.Writer.Size()).
			Str("user_agent", c.Request.UserAgent())

		// Add error message if exists
		if errorMessage != "" {
			event.Str("error", errorMessage)
		}

		// Add request headers (excluding sensitive ones)
		headers := make(map[string]string)
		for k, v := range c.Request.Header {
			// Skip sensitive headers
			if !isSensitiveHeader(k) && len(v) > 0 {
				headers[k] = v[0]
			}
		}
		if len(headers) > 0 {
			event.Interface("headers", headers)
		}

		// Log the event
		event.Msg("Request processed")
	}
}

// isSensitiveHeader checks if a header key contains sensitive information
func isSensitiveHeader(header string) bool {
	sensitiveHeaders := map[string]bool{
		"Authorization":       true,
		"Cookie":              true,
		"Set-Cookie":          true,
		"X-CSRF-Token":        true,
		"X-API-Key":           true,
		"X-Access-Token":      true,
		"X-Authorization":     true,
		"X-Forwarded-For":     true,
		"X-Real-IP":           true,
		"Proxy-Authorization": true,
	}
	return sensitiveHeaders[header]
}
