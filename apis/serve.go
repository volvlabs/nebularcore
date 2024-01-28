package apis

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/models/config"

	"github.com/gin-contrib/cors"
	"github.com/rs/zerolog/log"
)

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

	router.Use(cors.New(cors.Config{
		AllowOrigins: strings.Split(config.AllowedOrigins, ","),
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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
		server.Shutdown(ctx)
		wg.Done()

		return nil
	})

	defer wg.Wait()

	return server.ListenAndServe()
}
