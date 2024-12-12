package apis

import (
	"strings"

	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/tools/security"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
)

const (
	ContextClaimsKey           = "claims"
	ContextTenantIdKey         = "tenantId"
	ContextTenantSchemaNameKey = "tenantSchemaName"
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

		c.Set(ContextClaimsKey, claims)
		c.Next()
	}
}

func AuthorizeRequest(app core.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, _ := c.Get(ContextClaimsKey)
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
		tenantSession, err := app.Dao().WithSchemaSession(schemaName)
		if err != nil {
			log.Err(err).Msgf("error occurred creating db session for tenant %s", tenantId)
			NewInternalServerError(c)
			return
		}
		c.Set(ContextTenantIdKey, tenantId)
		c.Set(ContextTenantSchemaNameKey, schemaName)
		c.Set(core.ContextDBSessionKey, tenantSession)

		c.Next()
	}
}
