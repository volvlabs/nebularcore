package controlcenter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/jideobs/nebularcore/core/module"
)

type handler struct {
	app coreApp
}

func newHandler(app coreApp) *handler {
	return &handler{app: app}
}

func (h *handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/modules", h.modules)
	rg.GET("/config", h.config)
}

func (h *handler) modules(c *gin.Context) {
	modules := h.app.GetModulesByNamespace(module.PublicNamespace)
	c.JSON(http.StatusOK, modules)
}

func (h *handler) config(c *gin.Context) {
	config := h.app.Config()
	c.JSON(http.StatusOK, config)
}
