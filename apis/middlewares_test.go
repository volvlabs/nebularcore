package apis

import (
	"net/http"
	"testing"

	"gitlab.com/volvlabs/nebularcore/test"

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
			Name:   "should return 401 because of invalid auth token",
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

func TestAuthorizeRequest(t *testing.T) {
	// Arrange
	scenarios := []test.ApiScenario{
		{
			Name:   "should return 200 because role is allowed to access resource",
			Url:    "/admin/create",
			Method: http.MethodGet,
			Body:   nil,
			RequestHeaders: map[string]string{
				"Authorization": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzU3MzI1ODUsInJvbGUiOiJhZG1pbiJ9.jkiyHZoru0EZeOfEJqA3Ug2cMk4zkpS4dhTVpaZyk70",
			},
			ExpectedStatus: http.StatusOK,
			ExpectedContent: []string{
				`"code":200`,
				`"message":"success"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				router.Use(AuthenticateRequestThenLoadAuthContext(app))
				router.Use(AuthorizeRequest(app))
				router.GET("/admin/create", func(c *gin.Context) {
					c.JSON(http.StatusOK, map[string]any{"code": http.StatusOK, "message": "success"})
				})
			},
		},
		{
			Name:   "should return 403 because role is not allowed to access resource",
			Url:    "/admin/create",
			Method: http.MethodGet,
			Body:   nil,
			RequestHeaders: map[string]string{
				"Authorization": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzU3NTY0MDIsInJvbGUiOiJ1c2VyIn0.xxmWXfHz8Knfcgqji1t7u4rh5KhHZOmOOalnEWxQ1rU",
			},
			ExpectedStatus: http.StatusForbidden,
			ExpectedContent: []string{
				`"code":403`,
				`"message":"forbidden"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				router.Use(AuthenticateRequestThenLoadAuthContext(app))
				router.Use(AuthorizeRequest(app))
				router.GET("/admin/create", func(c *gin.Context) {
					c.JSON(http.StatusOK, map[string]any{"code": http.StatusOK, "message": "success"})
				})
			},
		},
		{
			Name:   "should return 403 since role is not available within the auth claims",
			Url:    "/admin/create",
			Method: http.MethodGet,
			Body:   nil,
			RequestHeaders: map[string]string{
				"Authorization": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzU3NTY4Nzl9.bwmD8PUR1kJrupU8baQCn_u--Jv75MMFmyAEF-GVy-4",
			},
			ExpectedStatus: http.StatusForbidden,
			ExpectedContent: []string{
				`"code":403`,
				`"message":"forbidden"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				router.Use(AuthenticateRequestThenLoadAuthContext(app))
				router.Use(AuthorizeRequest(app))
				router.GET("/admin/create", func(c *gin.Context) {
					c.JSON(http.StatusOK, map[string]any{"code": http.StatusOK, "message": "success"})
				})
			},
		},
	}
	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}
