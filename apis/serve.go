package apis

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/volvlabs/nebularcore/core"
	"github.com/volvlabs/nebularcore/models/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Endpoints(app core.App, endpoints config.Endpoints) {
	router := app.Router()

	// Add request logger middleware
	router.Use(RequestLogger(app))

	routerGroup := router.Group("")
	if endpoints.AuthEnabled {
		BindAuthApi(app, routerGroup, false)
	}
}

func Cors(allowedOrigins string, router *gin.Engine) {
	router.Use(cors.New(cors.Config{
		AllowOrigins: strings.Split(allowedOrigins, ","),
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Origin", "Content-Type", "Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

func Serve(app core.App, config config.ServeConfig) error {
	router := app.Router()

	baseCtx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", config.Host, config.Port),
		Handler: router,
		BaseContext: func(l net.Listener) context.Context {
			return baseCtx
		},
	}

	// Wait for server.Shutdown to return.
	var wg sync.WaitGroup

	// handle server graceful shutdown
	app.OnTerminate(func() error {
		log.Info().Msg("terminating server")
		cancelCtx()

		ctx, cancel := context.WithTimeout(
			context.Background(),
			time.Duration(config.ShutdownTimeout)*time.Second)
		defer cancel()

		wg.Add(1)
		shutdownErr := server.Shutdown(ctx)
		wg.Done()

		return shutdownErr
	})

	defer wg.Wait()

	return server.ListenAndServe()
}
