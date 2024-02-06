package apis

import (
	"net/http"
	"testing"

	"gitlab.com/jideobs/nebularcore/test"

	"github.com/gin-gonic/gin"
)

func TestHealthCheck(t *testing.T) {
	// Arrange:
	scenario := test.ApiScenario{
		Name:           "should return 200 success response",
		Url:            "/api",
		Method:         http.MethodGet,
		Body:           nil,
		ExpectedStatus: http.StatusOK,
		ExpectedContent: []string{
			`{"code":200,"message":"Ok"}`,
		},
		BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
			BindHealthApi(router.Group("/api"))
		},
	}
	// Act-Asset:
	scenario.Test(t)
}
