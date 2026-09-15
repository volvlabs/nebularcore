package http

import (
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

// FileServer defines the interface for serving files over HTTP
type FileServer interface {
	ServeFile(w http.ResponseWriter, r *http.Request, path string)
}

type Handler struct {
	provider FileServer
	prefix   string
}

func NewHandler(provider FileServer, prefix string) *Handler {
	return &Handler{
		provider: provider,
		prefix:   strings.TrimRight(prefix, "/"),
	}
}

// Register registers the storage routes with the given router
func (h *Handler) Register(router gin.IRouter) {
	if h.prefix != "" {
		router = router.Group(h.prefix)
	}

	router.GET("/*path", h.serveFile)
}

func (h *Handler) serveFile(c *gin.Context) {
	filePath := path.Clean(c.Param("path"))
	if filePath == "/" {
		c.Status(http.StatusNotFound)
		return
	}

	filePath = strings.TrimPrefix(filePath, "/")

	h.provider.ServeFile(c.Writer, c.Request, filePath)
}
