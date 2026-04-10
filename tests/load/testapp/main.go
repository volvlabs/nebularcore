package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/volvlabs/nebularcore/modules/event"
	websocket "github.com/volvlabs/nebularcore/modules/websocket"
	wsconfig "github.com/volvlabs/nebularcore/modules/websocket/config"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Health endpoint.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Event bus.
	eventModule, err := event.New()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create event module")
	}

	// WebSocket module with load-test-friendly config.
	wsModule := websocket.New(eventModule)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "load-test-secret"
	}

	cfg := &wsconfig.Config{
		Enabled: true,
		Server: wsconfig.ServerConfig{
			Host:                    "0.0.0.0",
			Port:                    port,
			ReadBufferSize:          4096,
			WriteBufferSize:         4096,
			ReadDeadline:            120 * time.Second,
			WriteDeadline:           60 * time.Second,
			MaxConnections:          200000,
			MaxConnectionsPerUser:   100,
			MaxConnectionsPerTenant: 200000,
		},
		Routing: wsconfig.RoutingConfig{
			MaxTopicLength:         256,
			MaxTopicsPerConnection: 100,
		},
		Security: wsconfig.SecurityConfig{
			AuthRequired: false,
			JWTSecret:    jwtSecret,
		},
		TenantMode: "header",
		Events: wsconfig.EventsConfig{
			AllowedEventTypes:   []string{"**"},
			InternalEventPrefix: "ws:",
		},
	}

	if err := wsModule.Configure(cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to configure websocket module")
	}

	ctx := context.Background()
	if err := wsModule.Initialize(ctx, nil, router); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize websocket module")
	}

	// Metrics endpoint.
	router.GET("/metrics", func(c *gin.Context) {
		manager := wsModule.Manager()
		c.JSON(http.StatusOK, gin.H{
			"total_connections": manager.Total(),
		})
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	go func() {
		log.Info().Str("port", port).Msg("load test server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wsModule.Shutdown(shutdownCtx)
	srv.Shutdown(shutdownCtx)
}
