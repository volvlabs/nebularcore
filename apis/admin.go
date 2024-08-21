package apis

import (
	"net/http"

	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/models"
	"gitlab.com/jideobs/nebularcore/models/requests"
	"gitlab.com/jideobs/nebularcore/services"
	"gitlab.com/jideobs/nebularcore/tools/security"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func BindAdminApi(app core.App, rg *gin.RouterGroup) {
	api := adminApi{app: app}

	subGroup := rg.Group("/admin")
	subGroup.POST("/login", api.login)
	subGroup.PUT("/refresh-token", api.refreshToken)

	authGroup := subGroup.Group("")
	authGroup.Use(AuthenticateRequestThenLoadAuthContext(app))
	authGroup.POST("", api.create)
	authGroup.PUT("/change-password", api.changePassword)
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

	refreshToken, err := security.NewJWT(
		jwt.MapClaims{"id": admin.Id, "identity": admin.Email, "role": "user"},
		api.app.Settings().AuthTokenRefreshSecret,
		api.app.Settings().AuthRefreshTokenExpiryDuration,
	)
	if err != nil {
		log.Err(err).Msg("UserAuth: error creating refresh token")
		NewInternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, map[string]any{
		"token":        token,
		"admin":        admin,
		"refreshToken": refreshToken,
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
	var refreshTokenRequest requests.RefreshTokenRequest
	if err := c.BindJSON(&refreshTokenRequest); err != nil {
		NewBadRequestError(c, "error handling submitted data", nil)
		return
	}

	auth := services.NewAuth(api.app)
	admin, err := auth.RefreshToken(refreshTokenRequest)
	if err != nil {
		if err == security.ErrInvalidRefreshToken {
			NewBadRequestError(c, "invalid refresh token", nil)
			return
		}

		HandleError(c, err)
		return
	}

	api.authResponse(c, admin)
}
