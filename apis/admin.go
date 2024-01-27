package apis

import (
	"net/http"

	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/models"
	"gitlab.com/jideobs/nebularcore/services"
	"gitlab.com/jideobs/nebularcore/tools/security"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func BindAdminApi(app core.App, rg *gin.RouterGroup) {
	api := adminApi{app: app}

	subGroup := rg.Group("/admin")
	subGroup.POST("", api.create)
	subGroup.POST("/login", api.login)

	authGroup := subGroup.Group("")
	authGroup.Use(AuthenticateRequestThenLoadAuthContext(app))
	authGroup.PUT("/change-password", api.changePassword)
	authGroup.POST("/refresh-token", api.refreshToken)
}

type adminApi struct {
	app core.App
}

func (api *adminApi) authResponse(c *gin.Context, admin *models.Admin) {
	token, err := security.NewJWT(
		jwt.MapClaims{"id": admin.Id, "role": admin.Role},
		api.app.Settings().AuthTokenSecret,
		api.app.Settings().AuthTokenDuration,
	)
	if err != nil {
		NewInternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, map[string]any{
		"token": token,
		"admin": admin,
	})
}

func (api *adminApi) create(c *gin.Context) {
	var adminCreateRequest services.AdminCreateRequest
	if err := c.BindJSON(&adminCreateRequest); err != nil {
		NewBadRequestError(c, "error handling submitted data", nil)
		return
	}

	adminCreateSrv := services.NewAdminCreate(api.app)
	admin, err := adminCreateSrv.Create(adminCreateRequest)
	if err != nil {
		HandleError(c, err)
		return
	}

	api.authResponse(c, admin)
}

func (api *adminApi) login(c *gin.Context) {
	var adminLoginRequest services.AdminLoginRequest
	if err := c.BindJSON(&adminLoginRequest); err != nil {
		NewBadRequestError(c, "error handling submitted data", nil)
		return
	}

	adminLoginSrv := services.NewAdminLogin(api.app)
	admin, err := adminLoginSrv.Submit(adminLoginRequest)
	if err != nil {
		HandleError(c, err)
		return
	}

	api.authResponse(c, admin)
}

func (api *adminApi) changePassword(c *gin.Context) {
	var adminChangePasswordRequest services.AdminPasswordChangeRequest
	if err := c.BindJSON(&adminChangePasswordRequest); err != nil {
		NewBadRequestError(c, "error handling submitted data", nil)
		return
	}

	adminChangePassword := services.NewAdminPasswordChange(api.app)
	claims, _ := c.Get("claims")
	id := claims.(jwt.MapClaims)["id"]

	err := adminChangePassword.ChangePassword(uuid.MustParse(id.(string)), adminChangePasswordRequest)
	if err != nil {
		HandleError(c, err)

		return
	}

	c.JSON(http.StatusOK, map[string]string{})
}

func (api *adminApi) refreshToken(c *gin.Context) {
	claimsRaw, _ := c.Get(ContextClaimsKey)
	claims := claimsRaw.(jwt.MapClaims)
	adminId := uuid.MustParse(claims["id"].(string))
	admin, err := api.app.Dao().FindAdminById(adminId)
	if err != nil {
		NewInternalServerError(c)
		return
	}
	api.authResponse(c, admin)
}
