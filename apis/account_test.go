package apis

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/volvlabs/nebularcore/test"
	"github.com/volvlabs/nebularcore/tools/filesystem"
	"net/http"
	"strings"
	"testing"
)

func TestCreateAdmin(t *testing.T) {
	testApp, _ := test.NewTestApp()
	tearDownMigration := test.RunMigration(
		t, filesystem.GetRootDir("../"), testApp.DataDir())
	defer tearDownMigration(t)

	admin := setupAdmin(testApp)
	token := getAuthorizationToken(testApp, admin)
	scenarios := []test.ApiScenario{
		{
			Name:   "should create admin successfully",
			Url:    "/accounts",
			Method: http.MethodPost,
			Body: strings.NewReader(`{
				"firstName": "John",
				"lastName": "Doe",
				"email": "john.doe@gmail.com",
				"role": "Admin",
				"password": "password123"
			}`),
			RunMigration: true,
			RequestHeaders: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", token),
			},
			ExpectedStatus: http.StatusOK,
			ExpectedContent: []string{
				`"token":`,
				`"user":{`,
				`"id"`,
				`"avatar":""`,
				`"email":"john.doe@gmail.com"`,
				`"firstName":"John"`,
				`"lastName":"Doe"`,
				`"role":"Admin"`,
				`"created"`,
				`"updated"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				BindAccountApi(app, router.Group(""))
			},
		},
		{
			Name:         "should return 400 bad request because of empty request body",
			Url:          "/accounts",
			Method:       http.MethodPost,
			Body:         strings.NewReader(""),
			RunMigration: true,
			RequestHeaders: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", token),
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedContent: []string{
				`"code":400`,
				`"message":"error handling submitted data"`,
				`"errors":`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				BindAccountApi(app, router.Group(""))
			},
		},
		{
			Name:         "should return 400 bad request because required vals are not set",
			Url:          "/accounts",
			Method:       http.MethodPost,
			Body:         strings.NewReader("{}"),
			RunMigration: true,
			RequestHeaders: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", token),
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedContent: []string{
				`"code":400`,
				`"message":"error validating request body"`,
				`"errors"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				BindAccountApi(app, router.Group(""))
			},
		},
		{
			Name:   "should return 500 internal server error because a server error occurred",
			Url:    "/accounts",
			Method: http.MethodPost,
			Body: strings.NewReader(`{
				"firstName": "John",
				"lastName": "Doe",
				"email": "john.doe@gmail.com",
				"role": "Admin",
				"password": "password123"
			}`),
			RunMigration: false,
			RequestHeaders: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", token),
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedContent: []string{
				`"code":500`,
				`"message":"internal server error"`,
				`"errors"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				BindAccountApi(app, router.Group(""))
			},
		},
	}
	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}
