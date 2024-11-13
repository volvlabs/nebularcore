package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/services/account"
)

func BindAccountApi(app core.App, rg *gin.RouterGroup) {
	api := accountApi{app: app}

	subGroup := rg.Group("/accounts")
	subGroup.Use(AuthenticateRequestThenLoadAuthContext(app))
	subGroup.POST("", api.registerAdmin)
}

type accountApi struct {
	app core.App
}

func (a *accountApi) registerAdmin(c *gin.Context) {
	var adminCreateRequest account.AdminCreateRequest
	if err := c.BindJSON(&adminCreateRequest); err != nil {
		log.Err(err).Msgf("AccountApi: error occurred marshalling submitted data")
		NewBadRequestError(c, "error handling submitted data", nil)
		return
	}

	accountService := account.New(a.app)
	admin, err := accountService.CreateAdmin(adminCreateRequest)
	if err != nil {
		HandleError(c, err)
		return
	}

	authApi := authApi{app: a.app}
	authApi.authResponseWithUserType(c, admin)
}
