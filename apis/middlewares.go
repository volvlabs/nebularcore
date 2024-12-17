package apis

import (
	"context"
	"strings"

	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/tools"
	"gitlab.com/jideobs/nebularcore/tools/security"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
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
		dbSession, err := app.Dao().WithSchemaSession(schemaName)
		if err != nil {
			log.Err(err).Msgf("error occurred creating db session for tenant %s", tenantId)
			NewInternalServerError(c)
			return
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, tools.ContextTenantIdKey, tenantId)
		ctx = context.WithValue(ctx, tools.ContextTenantSchemaNameKey, schemaName)
		ctx = context.WithValue(ctx, tools.ContextDBSessionKey, dbSession)

		c.Request = c.Request.WithContext(ctx)

		c.Next()

		app.Dao().ResetSchema()
	}
}
