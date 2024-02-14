package apis

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/jideobs/nebularcore/core"
)

func BindFilesApi(app core.App, r *gin.RouterGroup) {
	api := NewFilesApi(app)
	r.GET("/files", api.Get)
}

type filesApi struct {
	app core.App
}

func NewFilesApi(app core.App) *filesApi {
	return &filesApi{app: app}
}

// Todo: write tests coverage for this handler.
func (api *filesApi) Get(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		NewBadRequestError(c, "file key is required", nil)
		return
	}

	fs, err := api.app.NewFileSystem()
	if err != nil {
		log.Err(err).Msgf("error occurred creating file system")
		NewInternalServerError(c)
		return
	}
	defer fs.Close()

	content, contentType, err := fs.Download(key)
	if err != nil {
		log.Err(err).Msgf("error occurred downloading file")
		NewInternalServerError(c)
		return
	}

	c.Data(http.StatusOK, contentType, content)
}
