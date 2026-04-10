package apis

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/volvlabs/nebularcore/core"
	"github.com/volvlabs/nebularcore/tools"
	"github.com/volvlabs/nebularcore/tools/security"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
)

// responseBodyWriter is a custom response writer that captures the response body
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

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
		// Collect initial request data
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Copy relevant request headers
		headers := make(map[string]string)
		for k, v := range c.Request.Header {
			if !isSensitiveHeader(k) && len(v) > 0 {
				headers[k] = v[0]
			}
		}

		// Read the request body if it exists
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// Restore the request body for later use
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Setup response body capture
		responseBody := &bytes.Buffer{}
		writer := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           responseBody,
		}
		c.Writer = writer

		// Process request
		c.Next()

		// Capture final response status and size
		status := c.Writer.Status()
		bodySize := c.Writer.Size()
		latency := time.Since(start)

		// Get error if exists
		var errorMessage string
		if len(c.Errors) > 0 {
			errorMessage = c.Errors.String()
		}

		// Handle logging asynchronously
		go func(
			status int,
			bodySize int,
			latency time.Duration,
			path string,
			raw string,
			requestID string,
			method string,
			clientIP string,
			userAgent string,
			headers map[string]string,
			errorMessage string,
			requestBody []byte,
			responseBody *bytes.Buffer,
		) {
			// Get request method and path with query parameters
			if raw != "" {
				path = path + "?" + raw
			}

			// Create log event
			event := log.Info()
			isError := status >= 400
			if isError {
				event = log.Error()
			}

			// Add common fields
			event.Str("request_id", requestID).
				Str("client_ip", clientIP).
				Str("method", method).
				Str("path", path).
				Int("status", status).
				Dur("latency", latency).
				Int("body_size", bodySize).
				Str("user_agent", userAgent)

			// Add error message if exists
			if errorMessage != "" {
				event.Str("error", errorMessage)
			}

			// Add headers if any
			if len(headers) > 0 {
				event.Interface("headers", headers)
			}

			// If there's an error, include request and response bodies
			if isError {
				// Add request body if it exists and is not too large
				if len(requestBody) > 0 && len(requestBody) < 10000 { // Limit to 10KB
					if isJSON(headers["Content-Type"]) {
						var prettyJSON bytes.Buffer
						if err := json.Indent(&prettyJSON, requestBody, "", "  "); err == nil {
							event.Str("request_body", prettyJSON.String())
						} else {
							event.Str("request_body", string(requestBody))
						}
					} else {
						event.Str("request_body", string(requestBody))
					}
				}

				// Add response body if it exists and is not too large
				if responseBody.Len() > 0 && responseBody.Len() < 10000 { // Limit to 10KB
					respBody := responseBody.Bytes()
					if isJSON(c.Writer.Header().Get("Content-Type")) {
						var prettyJSON bytes.Buffer
						if err := json.Indent(&prettyJSON, respBody, "", "  "); err == nil {
							event.Str("response_body", prettyJSON.String())
						} else {
							event.Str("response_body", string(respBody))
						}
					} else {
						event.Str("response_body", string(respBody))
					}
				}
			}

			// Log the event
			event.Msg("Request processed")
		}(
			status,
			bodySize,
			latency,
			path,
			raw,
			requestID,
			method,
			clientIP,
			userAgent,
			headers,
			errorMessage,
			requestBody,
			responseBody,
		)
	}
}

// isJSON checks if the content type indicates JSON data
func isJSON(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "application/json")
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
