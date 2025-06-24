package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/jideobs/nebularcore/modules/auth/models/responses"
	"gitlab.com/jideobs/nebularcore/modules/auth/resource"
	"gitlab.com/jideobs/nebularcore/tools/handlers"
)

type ResourceHandler struct {
	manager resource.Manager
}

func NewResourceHandler(manager resource.Manager) *ResourceHandler {
	return &ResourceHandler{manager: manager}
}

func (r *ResourceHandler) RegisterRoutes(router *gin.RouterGroup) {
	resourceRouter := router.Group("/resource")
	{
		resourceRouter.GET("", r.listResources)
		resourceRouter.GET("/filter", r.filterResource)
	}
}

func (r *ResourceHandler) listResources(c *gin.Context) {
	page := r.manager.ListResources(c.Request.Context(), c.Request)
	if page.Error {
		handlers.HandleError(c, page.RawError)
		return
	}

	c.JSON(http.StatusOK, page)
}

func (r *ResourceHandler) filterResource(c *gin.Context) {
	filter := c.Query("filter")
	if filter == "" {
		c.JSON(http.StatusOK, []responses.FilterResourcePayload{})
	}

	resources, err := r.manager.FilterResource(c.Request.Context(), filter)
	if err != nil {
		handlers.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resources)
}
