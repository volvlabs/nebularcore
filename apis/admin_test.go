package apis

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"gitlab.com/jideobs/nebularcore/models"
	"gitlab.com/jideobs/nebularcore/test"
	"gitlab.com/jideobs/nebularcore/tools/filesystem"
	"gitlab.com/jideobs/nebularcore/tools/security"
)

func setupAdmin(app *test.TestApp) *models.Admin {
	user := &models.Admin{
		Email:     "john.doe@gmail.com",
		FirstName: "John",
		LastName:  "Doe",
	}
	app.Dao().Save(user)
	return user
}

func getAuthorizationToken(app *test.TestApp, user *models.Admin) string {
	token, _ := security.NewJWT(
		jwt.MapClaims{"id": user.Id, "identity": user.Email, "role": "admin"},
		app.Settings().AuthTokenSecret,
		app.Settings().AuthTokenDuration,
	)

	return token
}

func TestCreateAdmin(t *testing.T) {
	testapp, _ := test.NewTestApp()
	tearDownMigration := test.RunMigration(
		t, filesystem.GetRootDir("../"), testapp.DataDir())
	defer tearDownMigration(t)

	admin := setupAdmin(testapp)
	token := getAuthorizationToken(testapp, admin)
	scenarios := []test.ApiScenario{
		{
			Name:   "should create admin successfully",
			Url:    "/admin",
			Method: http.MethodPost,
			Body: strings.NewReader(`{
				"firstName": "John",
				"lastName": "Doe",
				"email": "john.doe@gmail.com",
				"role": "operator",
				"password": "password123"
			}`),
			RunMigration: true,
			RequestHeaders: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", token),
			},
			ExpectedStatus: http.StatusOK,
			ExpectedContent: []string{
				`"token":`,
				`"admin":{`,
				`"id"`,
				`"avatar":""`,
				`"email":"john.doe@gmail.com"`,
				`"firstName":"John"`,
				`"lastName":"Doe"`,
				`"role":"operator"`,
				`"created"`,
				`"updated"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				BindAdminApi(app, router.Group(""))
			},
		},
		{
			Name:         "should return 400 bad request because of empty request body",
			Url:          "/admin",
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
				BindAdminApi(app, router.Group(""))
			},
		},
		{
			Name:         "should return 400 bad request because required vals are not set",
			Url:          "/admin",
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
				BindAdminApi(app, router.Group(""))
			},
		},
		{
			Name:   "should return 500 internal server error because a server error occurred",
			Url:    "/admin",
			Method: http.MethodPost,
			Body: strings.NewReader(`{
				"firstName": "John",
				"lastName": "Doe",
				"email": "john.doe@gmail.com",
				"role": "operator",
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
				BindAdminApi(app, router.Group(""))
			},
		},
	}
	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestLoginAdmin(t *testing.T) {
	scenarios := []test.ApiScenario{
		{
			Name:   "should login admin successfully",
			Url:    "/admin/login",
			Method: http.MethodPost,
			Body: strings.NewReader(`{
				"identity": "john.doe23@gmail.com",
				"password": "XXXXXXXXXXX"
			}`),
			RunMigration:   true,
			RequestHeaders: map[string]string{},
			ExpectedStatus: http.StatusOK,
			ExpectedContent: []string{
				`"token":`,
				`"admin":{`,
				`"id"`,
				`"avatar":""`,
				`"email":"john.doe23@gmail.com"`,
				`"firstName":"John"`,
				`"lastName":"Doe"`,
				`"role":"operator"`,
				`"created"`,
				`"updated"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				admin := &models.Admin{
					FirstName:    "John",
					LastName:     "Doe",
					Email:        "john.doe23@gmail.com",
					Role:         "operator",
					PasswordHash: "$2a$12$oJSuU25enGzScfVVZDtj9O4roWt1Z4OH3XId4Y109ZE7BsZhmOcGO",
				}
				app.Dao().CreateAdmin(admin)
				app.Dao().CreateAuth(&models.Auth{
					Identity:     admin.Email,
					PasswordHash: admin.PasswordHash,
				})
				BindAdminApi(app, router.Group(""))
			},
		},
		{
			Name:           "should return 400 bad request because of empty request body",
			Url:            "/admin/login",
			Method:         http.MethodPost,
			Body:           strings.NewReader(""),
			RunMigration:   true,
			RequestHeaders: map[string]string{},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedContent: []string{
				`"code":400`,
				`"message":"error handling submitted data"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				BindAdminApi(app, router.Group(""))
			},
		},
		{
			Name:   "should return 400 bad request because of invalid email",
			Url:    "/admin/login",
			Method: http.MethodPost,
			Body: strings.NewReader(`{
				"identity": "XXXXXXXXXXXXXXXXXX",
				"password": "XXXXXXXXXXX"
			}`),
			RunMigration:   true,
			RequestHeaders: map[string]string{},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedContent: []string{
				`"code":400`,
				`"message":"error validating request body`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				BindAdminApi(app, router.Group(""))
			},
		},
		{
			Name:   "should return 400 bad request because of invalid password",
			Url:    "/admin/login",
			Method: http.MethodPost,
			Body: strings.NewReader(`{
				"identity": "john.doe@gmail.com",
				"password": "incorrectpassword"
			}`),
			RunMigration:   true,
			RequestHeaders: map[string]string{},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedContent: []string{
				`"code":400`,
				`"message":"invalid login credentials"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				app.Dao().CreateAdmin(&models.Admin{
					FirstName:    "John",
					LastName:     "Doe",
					Email:        "john.doe@gmail.com",
					Role:         "operator",
					PasswordHash: "$2a$12$oJSuU25enGzScfVVZDtj9O4roWt1Z4OH3XId4Y109ZE7BsZhmOcGO",
				})
				BindAdminApi(app, router.Group(""))
			},
		},
		{
			Name:   "should return 500 internal server error",
			Url:    "/admin/login",
			Method: http.MethodPost,
			Body: strings.NewReader(`{
				"identity": "john.doe@gmail.com",
				"password": "XXXXXXXXXXX"
			}`),
			RunMigration:   false,
			RequestHeaders: map[string]string{},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedContent: []string{
				`"code":500`,
				`"message":"internal server error"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				BindAdminApi(app, router.Group(""))
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestChangePasswordAdmin(t *testing.T) {
	scenarios := []test.ApiScenario{
		{
			Name:   "should change password successfully",
			Url:    "/admin/change-password",
			Method: http.MethodPut,
			Body: strings.NewReader(`{
				"oldPassword": "XXXXXXXXXXX",
				"password": "XXXXXXXXXXX123",
				"confirmPassword": "XXXXXXXXXXX123"
			}`),
			RunMigration: true,
			RequestHeaders: map[string]string{
				"Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzYyMDI5MDUsImlkIjoiNTc5Mzg3ZmMtZDhjZC00ODAzLThiNmUtNTI3NDJkMTQ2MGE4Iiwicm9sZSI6Im9wZXJhdG9yIn0.MFj5LWW4VEJbPbFhiAKNGNuDZ5SVZwPLDUyKcYv-boo",
			},
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				admin := &models.Admin{
					FirstName:    "John",
					LastName:     "Doe",
					Email:        "john.doe@gmail.com",
					Role:         "operator",
					PasswordHash: "$2a$12$oJSuU25enGzScfVVZDtj9O4roWt1Z4OH3XId4Y109ZE7BsZhmOcGO",
				}
				admin.SetId(uuid.MustParse("579387fc-d8cd-4803-8b6e-52742d1460a8"))
				app.Dao().CreateAdmin(admin)
				app.Dao().CreateAuth(&models.Auth{
					Identity:     admin.Email,
					PasswordHash: admin.PasswordHash,
				})
				BindAdminApi(app, router.Group(""))
			},
		},
		{
			Name:   "should return 400 as current password is invalid",
			Url:    "/admin/change-password",
			Method: http.MethodPut,
			Body: strings.NewReader(`{
				"oldPassword": "invalidpassword",
				"password": "XXXXXXXXXXX123",
				"confirmPassword": "XXXXXXXXXXX123"
			}`),
			RunMigration: true,
			RequestHeaders: map[string]string{
				"Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzYyMDI5MDUsImlkIjoiNTc5Mzg3ZmMtZDhjZC00ODAzLThiNmUtNTI3NDJkMTQ2MGE4Iiwicm9sZSI6Im9wZXJhdG9yIn0.MFj5LWW4VEJbPbFhiAKNGNuDZ5SVZwPLDUyKcYv-boo",
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedContent: []string{
				`"code":400`,
				`"message":"current password is incorrect"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				admin := &models.Admin{
					FirstName:    "John",
					LastName:     "Doe",
					Email:        "john.doe@gmail.com",
					Role:         "operator",
					PasswordHash: "$2a$12$oJSuU25enGzScfVVZDtj9O4roWt1Z4OH3XId4Y109ZE7BsZhmOcGO",
				}
				admin.SetId(uuid.MustParse("579387fc-d8cd-4803-8b6e-52742d1460a8"))
				app.Dao().CreateAdmin(admin)
				app.Dao().CreateAuth(&models.Auth{
					Identity:     admin.Email,
					PasswordHash: admin.PasswordHash,
				})
				BindAdminApi(app, router.Group(""))
			},
		},
		{
			Name:         "should return 400 because of bad JSON",
			Url:          "/admin/change-password",
			Method:       http.MethodPut,
			Body:         strings.NewReader(``),
			RunMigration: true,
			RequestHeaders: map[string]string{
				"Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzYyMDI5MDUsImlkIjoiNTc5Mzg3ZmMtZDhjZC00ODAzLThiNmUtNTI3NDJkMTQ2MGE4Iiwicm9sZSI6Im9wZXJhdG9yIn0.MFj5LWW4VEJbPbFhiAKNGNuDZ5SVZwPLDUyKcYv-boo",
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedContent: []string{
				`"code":400`,
				`"message":"error handling submitted data"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				admin := &models.Admin{
					FirstName:    "John",
					LastName:     "Doe",
					Email:        "john.doe@gmail.com",
					Role:         "operator",
					PasswordHash: "$2a$12$oJSuU25enGzScfVVZDtj9O4roWt1Z4OH3XId4Y109ZE7BsZhmOcGO",
				}
				admin.SetId(uuid.MustParse("579387fc-d8cd-4803-8b6e-52742d1460a8"))
				app.Dao().CreateAdmin(admin)
				app.Dao().CreateAuth(&models.Auth{
					Identity:     admin.Email,
					PasswordHash: admin.PasswordHash,
				})
				BindAdminApi(app, router.Group(""))
			},
		},
		{
			Name:   "should return 500 internal server error because of server error",
			Url:    "/admin/change-password",
			Method: http.MethodPut,
			Body: strings.NewReader(`{
				"oldPassword": "XXXXXXXXXXX",
				"password": "XXXXXXXXXXX123",
				"confirmPassword": "XXXXXXXXXXX123"
			}`),
			RunMigration: false,
			RequestHeaders: map[string]string{
				"Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzYyMDI5MDUsImlkIjoiNTc5Mzg3ZmMtZDhjZC00ODAzLThiNmUtNTI3NDJkMTQ2MGE4Iiwicm9sZSI6Im9wZXJhdG9yIn0.MFj5LWW4VEJbPbFhiAKNGNuDZ5SVZwPLDUyKcYv-boo",
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedContent: []string{
				`"code":500`,
				`"message":"internal server error"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				admin := &models.Admin{
					FirstName:    "John",
					LastName:     "Doe",
					Email:        "john.doe@gmail.com",
					Role:         "operator",
					PasswordHash: "$2a$12$oJSuU25enGzScfVVZDtj9O4roWt1Z4OH3XId4Y109ZE7BsZhmOcGO",
				}
				admin.SetId(uuid.MustParse("579387fc-d8cd-4803-8b6e-52742d1460a8"))
				app.Dao().CreateAdmin(admin)
				app.Dao().CreateAuth(&models.Auth{
					Identity:     admin.Email,
					PasswordHash: admin.PasswordHash,
				})
				BindAdminApi(app, router.Group(""))
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestRefreshPassword(t *testing.T) {
	scenarios := []test.ApiScenario{
		{
			Name:   "should refresh token successfully",
			Url:    "/admin/refresh-token",
			Method: http.MethodPut,
			Body: strings.NewReader(`{
				"refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzYyMDI5MDUsImlkIjoiNTc5Mzg3ZmMtZDhjZC00ODAzLThiNmUtNTI3NDJkMTQ2MGE4Iiwicm9sZSI6Im9wZXJhdG9yIn0.MFj5LWW4VEJbPbFhiAKNGNuDZ5SVZwPLDUyKcYv-boo"
			}`),
			RunMigration:   true,
			RequestHeaders: map[string]string{},
			ExpectedStatus: http.StatusOK,
			ExpectedContent: []string{
				`"token":`,
				`"admin":{`,
				`"id"`,
				`"avatar":""`,
				`"email":"john.doe@gmail.com"`,
				`"firstName":"John"`,
				`"lastName":"Doe"`,
				`"role":"operator"`,
				`"created"`,
				`"updated"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				admin := &models.Admin{
					FirstName:    "John",
					LastName:     "Doe",
					Email:        "john.doe@gmail.com",
					Role:         "operator",
					PasswordHash: "$2a$12$oJSuU25enGzScfVVZDtj9O4roWt1Z4OH3XId4Y109ZE7BsZhmOcGO",
				}
				admin.SetId(uuid.MustParse("579387fc-d8cd-4803-8b6e-52742d1460a8"))
				app.Dao().CreateAdmin(admin)
				BindAdminApi(app, router.Group(""))
			},
		},
		{
			Name:   "should return 500 internal server error because of a server error",
			Url:    "/admin/refresh-token",
			Method: http.MethodPut,
			Body: strings.NewReader(`{
				"refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzYyMDI5MDUsImlkIjoiNTc5Mzg3ZmMtZDhjZC00ODAzLThiNmUtNTI3NDJkMTQ2MGE4Iiwicm9sZSI6Im9wZXJhdG9yIn0.MFj5LWW4VEJbPbFhiAKNGNuDZ5SVZwPLDUyKcYv-boo"
			}`),
			RunMigration:   false,
			RequestHeaders: map[string]string{},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedContent: []string{
				`"code":500`,
				`"message":"internal server error"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				BindAdminApi(app, router.Group(""))
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}
