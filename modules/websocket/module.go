package websocket

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"gitlab.com/jideobs/nebularcore/core/migration_runner"
	"gitlab.com/jideobs/nebularcore/core/module"
	"gitlab.com/jideobs/nebularcore/modules/event"
	wsauth "gitlab.com/jideobs/nebularcore/modules/websocket/auth"
	"gitlab.com/jideobs/nebularcore/modules/websocket/bridge"
	wsconfig "gitlab.com/jideobs/nebularcore/modules/websocket/config"
	"gitlab.com/jideobs/nebularcore/modules/websocket/connections"
	"gitlab.com/jideobs/nebularcore/modules/websocket/handlers"
	"gitlab.com/jideobs/nebularcore/modules/websocket/store"
)

// Module implements the WebSocket module for NebularCore.
type Module struct {
	name    string
	version string
	config  *wsconfig.Config

	eventBus event.Bus
	manager  *connections.Manager
	pool     *connections.Pool
	subs     *store.Subscriptions
	adapter  *store.Adapter
	router   *bridge.Router
	evBridge *bridge.EventBridge
}

// New creates a new WebSocket module. The event.Bus is required for event
// bridging between WebSocket clients and the internal event system.
func New(bus event.Bus) *Module {
	return &Module{
		name:     "websocket",
		version:  "1.0.0",
		config:   wsconfig.DefaultConfig(),
		eventBus: bus,
	}
}

// Name implements module.Module.
func (m *Module) Name() string { return m.name }

// Version implements module.Module.
func (m *Module) Version() string { return m.version }

// Dependencies implements module.Module.
func (m *Module) Dependencies() []string { return []string{"event"} }

// Namespace implements module.Module.
func (m *Module) Namespace() module.ModuleNamespace { return module.PublicNamespace }

// NewConfig implements module.Module.
func (m *Module) NewConfig() any { return wsconfig.DefaultConfig() }

// Configure implements module.Module.
func (m *Module) Configure(config any) error {
	cfg, ok := config.(*wsconfig.Config)
	if !ok {
		return fmt.Errorf("websocket: invalid config type %T", config)
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	m.config = cfg
	return nil
}

// ProvidesMigrations implements module.Module.
func (m *Module) ProvidesMigrations() bool { return false }

// MigrationsDir implements module.Module.
func (m *Module) MigrationsDir() string { return "" }

// GetMigrationSources implements module.Module.
func (m *Module) GetMigrationSources(projectRoot string) []migration_runner.Source { return nil }

// Initialize implements module.Module.
func (m *Module) Initialize(ctx context.Context, db *gorm.DB, ginRouter *gin.Engine) error {
	if !m.config.Enabled {
		log.Info().Msg("websocket module is disabled")
		return nil
	}

	m.manager = connections.NewManager(m.config.Server.MaxConnections)
	m.pool = connections.NewPool(
		m.manager,
		m.config.Server.MaxConnectionsPerUser,
		m.config.Server.MaxConnectionsPerTenant,
	)

	m.subs = store.NewSubscriptions()
	validators := store.NewValidatorRegistry()
	m.adapter = store.NewAdapter(m.manager, m.subs, validators)
	m.evBridge = bridge.NewEventBridge(m.eventBus, m.manager, m.subs, m.config.Events.AllowedEventTypes)

	m.router = bridge.NewRouter()
	handlers.RegisterMessageHandlers(
		m.router,
		m.manager,
		m.subs,
		m.eventBus,
		m.config.Routing.MaxTopicsPerConnection,
		validators,
		m.evBridge,
	)

	if len(m.config.Events.AllowedEventTypes) > 0 {
		if err := m.evBridge.Start(ctx); err != nil {
			return fmt.Errorf("websocket: failed to start event bridge: %w", err)
		}
	}

	wsHandler := handlers.NewWebSocketHandler(m.config, m.manager, m.pool, m.subs, m.router)

	ginRouter.GET("/ws",
		wsauth.Middleware(m.config.Security.JWTSecret, m.config.Security.AuthRequired),
		wsHandler.Handle,
	)

	log.Info().
		Int64("max_connections", m.config.Server.MaxConnections).
		Msg("websocket module initialized")

	return nil
}

// Shutdown implements module.Module.
func (m *Module) Shutdown(ctx context.Context) error {
	if m.manager == nil {
		return nil
	}

	conns := m.manager.GetAll()
	for _, c := range conns {
		c.Close()
	}

	log.Info().Int("closed", len(conns)).Msg("websocket module shut down")
	return nil
}

// Manager returns the connection manager for external access.
func (m *Module) Manager() *connections.Manager { return m.manager }

// Adapter returns the adapter for other modules to interact with WebSocket connections.
func (m *Module) Adapter() *store.Adapter { return m.adapter }

// Subscriptions returns the subscription store.
func (m *Module) Subscriptions() *store.Subscriptions { return m.subs }
