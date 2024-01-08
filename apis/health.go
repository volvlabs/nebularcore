package apis

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func BindHealthApi(rg *gin.RouterGroup) {
	api := healthApi{}

	subGroup := rg.Group("/health")
	subGroup.GET("", api.healthCheck)
}

type healthApi struct {
}

type healthCheckResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (api *healthApi) healthCheck(c *gin.Context) {
	resp := new(healthCheckResponse)
	resp.Code = http.StatusOK
	resp.Message = "Ok"

	c.JSON(http.StatusOK, resp)
}
