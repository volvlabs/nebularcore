package test

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"gitlab.com/jideobs/nebularcore/tools/filesystem"
)

type ApiScenario struct {
	Name            string
	Url             string
	Method          string
	Body            io.Reader
	RunMigration    bool
	RequestHeaders  map[string]string
	ExpectedStatus  int
	ExpectedContent []string
	BeforeTestFunc  func(t *testing.T, app *TestApp, router *gin.Engine)
}

func (a *ApiScenario) createRoute() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.Default()
}

func (a *ApiScenario) Test(t *testing.T) {
	t.Run(a.Name, func(t *testing.T) {
		testApp, err := NewTestApp()
		if err != nil {
			t.Errorf("error occurred creating test app, %v", err)
		}

		if a.RunMigration {
			tearDownMigration := RunMigration(t, filesystem.GetRootDir("../../"), testApp.DataDir())
			defer tearDownMigration(t)
		}

		r := httptest.NewRequest(a.Method, a.Url, a.Body)

		for k, v := range a.RequestHeaders {
			r.Header.Set(k, v)
		}

		router := a.createRoute()
		w := httptest.NewRecorder()

		if a.BeforeTestFunc != nil {
			a.BeforeTestFunc(t, testApp, router)
		}

		router.ServeHTTP(w, r)

		for _, expectedContent := range a.ExpectedContent {
			if !strings.Contains(w.Body.String(), expectedContent) {
				t.Errorf("expected content %s not found in %s", expectedContent, w.Body.String())
			}
		}

		assert.Equal(t, a.ExpectedStatus, w.Code)
	})
}
