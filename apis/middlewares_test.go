package apis

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/volvlabs/nebularcore/core"
	"github.com/volvlabs/nebularcore/test"
	"github.com/volvlabs/nebularcore/tools"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticateRequestThenLoadAuthContext(t *testing.T) {
	// Arrange
	scenarios := []test.ApiScenario{
		{
			Name:   "should return 200 success response",
			Url:    "/",
			Method: http.MethodGet,
			Body:   nil,
			RequestHeaders: map[string]string{
				"Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdCIsImV4cCI6MTg5ODYzNjEzN30.gqRkHjpK5s1PxxBn9qPaWEWxTbpc1PPSD-an83TsXRY",
			},
			ExpectedStatus: http.StatusOK,
			ExpectedContent: []string{
				`"code":200`,
				`"message":"success"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				router.Use(AuthenticateRequestThenLoadAuthContext(app))
				router.GET("/", func(c *gin.Context) {
					claims, exists := c.Get("claims")
					assert.True(t, exists)
					assert.NotNil(t, claims)
					c.JSON(http.StatusOK, map[string]any{"code": http.StatusOK, "message": "success"})
				})
			},
		},
		{
			Name:   "should return 401 because of invalid authentication token",
			Url:    "/",
			Method: http.MethodGet,
			Body:   nil,
			RequestHeaders: map[string]string{
				"Authorization": "invalid_token",
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedContent: []string{
				`"code":401`,
				`"message":"unauthorized"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				router.Use(AuthenticateRequestThenLoadAuthContext(app))
				router.GET("/", func(c *gin.Context) {
					c.JSON(http.StatusOK, map[string]any{"code": http.StatusOK, "message": "success"})
				})
			},
		},
		{
			Name:   "should return 401 because token value is empty",
			Url:    "/",
			Method: http.MethodGet,
			Body:   nil,
			RequestHeaders: map[string]string{
				"Authorization": "",
			},
			ExpectedStatus: http.StatusUnauthorized,
			ExpectedContent: []string{
				`"code":401`,
				`"message":"unauthorized"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				router.Use(AuthenticateRequestThenLoadAuthContext(app))
				router.GET("/", func(c *gin.Context) {
					c.JSON(http.StatusOK, map[string]any{"code": http.StatusOK, "message": "success"})
				})
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

type tenantTestApp struct{ core.App }

func (tenantTestApp) SchemaName(tenantId string) string { return "tenant_" + tenantId }

type traceCtxKey struct{}

func TestTenantMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("preserves values and cancellation from the request context", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), traceCtxKey{}, "trace-1")
			c.Request = c.Request.WithContext(ctx)
			c.Next()
		})
		r.Use(TenantMiddleware(tenantTestApp{}))

		var ctx context.Context
		r.GET("/", func(c *gin.Context) {
			ctx = c.Request.Context()
			c.Status(http.StatusOK)
		})

		reqCtx, cancel := context.WithCancel(context.Background())
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(reqCtx)
		req.Header.Set("X-Tenant-ID", "acme")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		cancel()

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "acme", ctx.Value(tools.ContextTenantIdKey))
		assert.Equal(t, "tenant_acme", ctx.Value(tools.ContextTenantSchemaNameKey))
		assert.Equal(t, "trace-1", ctx.Value(traceCtxKey{}), "upstream context values must survive")
		assert.ErrorIs(t, ctx.Err(), context.Canceled, "request cancellation must propagate")
	})

	t.Run("rejects requests without tenant id", func(t *testing.T) {
		r := gin.New()
		r.Use(TenantMiddleware(tenantTestApp{}))
		r.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
