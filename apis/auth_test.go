package apis

import (
	"net/http"
	"strings"
	"testing"

	"github.com/volvlabs/nebularcore/entities"
	"github.com/volvlabs/nebularcore/tools/types"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/volvlabs/nebularcore/test"
	"github.com/volvlabs/nebularcore/tools/security"
)

func setupAdmin(app *test.TestApp) *entities.Admin {
	user := &entities.Admin{
		UserBase: entities.UserBase{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe23@gmail.com",
			Role:      types.Admin,
		},
	}
	user.SetId(uuid.MustParse("579387fc-d8cd-4803-8b6e-52742d1460a8"))
	app.Dao().Save(user)
	return user
}

func createAuth(t *testing.T, app *test.TestApp, admin *entities.Admin) *entities.Auth {
	auth := &entities.Auth{
		Identity:      admin.Email,
		UserId:        admin.Id,
		Role:          types.Admin,
		UserTableName: "admins",
		PasswordHash:  "$2a$12$oJSuU25enGzScfVVZDtj9O4roWt1Z4OH3XId4Y109ZE7BsZhmOcGO",
	}
	err := app.Dao().CreateAuth(auth)
	if err != nil {
		t.Fatal(err)
	}

	return auth
}

func getAuthorizationToken(app *test.TestApp, user *entities.Admin) string {
	token, _ := security.NewJWT(
		jwt.MapClaims{"id": user.Id, "identity": user.Email, "role": "Admin"},
		app.Settings().AuthTokenSecret,
		app.Settings().AuthTokenDuration,
	)

	return token
}

func TestLoginAdmin(t *testing.T) {
	scenarios := []test.ApiScenario{
		{
			Name:   "should login admin successfully",
			Url:    "/login",
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
				`"user":{`,
				`"id"`,
				`"avatar":""`,
				`"email":"john.doe23@gmail.com"`,
				`"first_name":"John"`,
				`"last_name":"Doe"`,
				`"role":"Admin"`,
				`"created"`,
				`"updated"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				admin := setupAdmin(app)
				createAuth(t, app, admin)
				BindAuthApi(app, router.Group(""), true)
			},
		},
		{
			Name:           "should return 400 bad request because of empty request body",
			Url:            "/login",
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
				BindAuthApi(app, router.Group(""), true)
			},
		},
		{
			Name:   "should return 400 bad request because of invalid password",
			Url:    "/login",
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
				admin := setupAdmin(app)
				createAuth(t, app, admin)
				BindAuthApi(app, router.Group(""), true)
			},
		},
		{
			Name:   "should return 500 internal server error",
			Url:    "/login",
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
				BindAuthApi(app, router.Group(""), true)
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
			Url:    "/change-password",
			Method: http.MethodPut,
			Body: strings.NewReader(`{
				"currentPassword": "XXXXXXXXXXX",
				"newPassword": "XXXXXXXXXXX123",
				"confirmPassword": "XXXXXXXXXXX123"
			}`),
			RunMigration: true,
			RequestHeaders: map[string]string{
				"Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4MjQyNzc3MDEsImlkIjoiNTc5Mzg3ZmMtZDhjZC00ODAzLThiNmUtNTI3NDJkMTQ2MGE4IiwiaWRlbnRpdHkiOiJqb2huLmRvZTIzQGdtYWlsLmNvbSIsInJvbGUiOiJBZG1pbiJ9.Nfl0_0weZ4jG2WhFgEAelCOJw1s6copIVJblCZel5bo",
			},
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				admin := setupAdmin(app)
				createAuth(t, app, admin)
				BindAuthApi(app, router.Group(""), true)
			},
		},
		{
			Name:   "should return 400 as current password is invalid",
			Url:    "/change-password",
			Method: http.MethodPut,
			Body: strings.NewReader(`{
				"currentPassword": "invalidpassword",
				"newPassword": "XXXXXXXXXXX123",
				"confirmPassword": "XXXXXXXXXXX123"
			}`),
			RunMigration: true,
			RequestHeaders: map[string]string{
				"Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4MjQyNzc3MDEsImlkIjoiNTc5Mzg3ZmMtZDhjZC00ODAzLThiNmUtNTI3NDJkMTQ2MGE4IiwiaWRlbnRpdHkiOiJqb2huLmRvZTIzQGdtYWlsLmNvbSIsInJvbGUiOiJBZG1pbiJ9.Nfl0_0weZ4jG2WhFgEAelCOJw1s6copIVJblCZel5bo",
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedContent: []string{
				`"code":400`,
				`"message":"current password is incorrect"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				admin := setupAdmin(app)
				createAuth(t, app, admin)
				BindAuthApi(app, router.Group(""), true)
			},
		},
		{
			Name:         "should return 400 because of bad JSON",
			Url:          "/change-password",
			Method:       http.MethodPut,
			Body:         strings.NewReader(``),
			RunMigration: true,
			RequestHeaders: map[string]string{
				"Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4MjQyNzc3MDEsImlkIjoiNTc5Mzg3ZmMtZDhjZC00ODAzLThiNmUtNTI3NDJkMTQ2MGE4IiwiaWRlbnRpdHkiOiJqb2huLmRvZTIzQGdtYWlsLmNvbSIsInJvbGUiOiJBZG1pbiJ9.Nfl0_0weZ4jG2WhFgEAelCOJw1s6copIVJblCZel5bo",
			},
			ExpectedStatus: http.StatusBadRequest,
			ExpectedContent: []string{
				`"code":400`,
				`"message":"error handling submitted data"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				admin := setupAdmin(app)
				createAuth(t, app, admin)
				BindAuthApi(app, router.Group(""), true)
			},
		},
		{
			Name:   "should return 500 internal server error because of server error",
			Url:    "/change-password",
			Method: http.MethodPut,
			Body: strings.NewReader(`{
				"currentPassword": "XXXXXXXXXXX",
				"newPassword": "XXXXXXXXXXX123",
				"confirmPassword": "XXXXXXXXXXX123"
			}`),
			RunMigration: false,
			RequestHeaders: map[string]string{
				"Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE4MjQyNzc3MDEsImlkIjoiNTc5Mzg3ZmMtZDhjZC00ODAzLThiNmUtNTI3NDJkMTQ2MGE4IiwiaWRlbnRpdHkiOiJqb2huLmRvZTIzQGdtYWlsLmNvbSIsInJvbGUiOiJBZG1pbiJ9.Nfl0_0weZ4jG2WhFgEAelCOJw1s6copIVJblCZel5bo",
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedContent: []string{
				`"code":500`,
				`"message":"internal server error"`,
			},
			BeforeTestFunc: func(t *testing.T, app *test.TestApp, router *gin.Engine) {
				BindAuthApi(app, router.Group(""), true)
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}
