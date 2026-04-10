package apis

import (
	"net/http"
	"testing"

	"github.com/volvlabs/nebularcore/test"

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
