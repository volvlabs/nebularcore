package websocket

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/volvlabs/nebularcore/core/config"
	"github.com/volvlabs/nebularcore/core/migration_runner"
	"github.com/volvlabs/nebularcore/core/module"
	"github.com/volvlabs/nebularcore/modules/event"
	wsauth "github.com/volvlabs/nebularcore/modules/websocket/auth"
	"github.com/volvlabs/nebularcore/modules/websocket/bridge"
	wsconfig "github.com/volvlabs/nebularcore/modules/websocket/config"
	"github.com/volvlabs/nebularcore/modules/websocket/connections"
	"github.com/volvlabs/nebularcore/modules/websocket/handlers"
	"github.com/volvlabs/nebularcore/modules/websocket/store"
)

// Module implements the WebSocket module for NebularCore.
type Module struct {
	name    string
	version string
	config  *wsconfig.Config

	eventBus          event.Bus
	manager           *connections.Manager
	pool              *connections.Pool
	subs              *store.Subscriptions
	adapter           *store.Adapter
	router            *bridge.Router
	evBridge          *bridge.EventBridge
	validators        *store.ValidatorRegistry
	pendingValidators []pendingValidator
	claimsResolver    wsauth.ClaimsResolverFunc
}

type pendingValidator struct {
	pattern string
	fn      store.TopicValidatorFunc
}

// Option configures a Module before Initialize runs.
type Option func(*Module)

// WithValidator registers a topic validator for topics matching pattern,
// applied once Initialize builds the module's ValidatorRegistry. Multiple
// options may be supplied; all are applied in order.
func WithValidator(pattern string, fn store.TopicValidatorFunc) Option {
	return func(m *Module) {
		m.pendingValidators = append(m.pendingValidators, pendingValidator{pattern: pattern, fn: fn})
	}
}

// WithClaimsResolver overrides token verification for the /ws upgrade
// handshake to use resolve instead of Config.Security's fixed HS256
// shared secret — for hosts whose own tokens are signed with a different
// algorithm (e.g. RS256) and already have a TokenIssuer that knows how to
// validate them.
func WithClaimsResolver(resolve wsauth.ClaimsResolverFunc) Option {
	return func(m *Module) {
		m.claimsResolver = resolve
	}
}

// New creates a new WebSocket module. The event.Bus is required for event
// bridging between WebSocket clients and the internal event system. Options
// (e.g. WithValidator) are applied before Initialize starts the event
// bridge, so callers can safely register topic-authorization rules here.
func New(bus event.Bus, opts ...Option) *Module {
	m := &Module{
		name:     "websocket",
		version:  "1.0.0",
		config:   wsconfig.DefaultConfig(),
		eventBus: bus,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
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
func (m *Module) NewConfig() config.Config { return wsconfig.DefaultConfig() }

// Configure implements module.Module.
func (m *Module) Configure(config config.Config) error {
	cfg, ok := config.(*wsconfig.Config)
	if !ok {
		return fmt.Errorf("websocket: invalid config type %T", config)
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
	m.manager.SetPresenceListener(bridge.NewPresenceBroadcaster(m.eventBus, m.config.Events.PresenceDebounce))
	m.pool = connections.NewPool(
		m.manager,
		m.config.Server.MaxConnectionsPerUser,
		m.config.Server.MaxConnectionsPerTenant,
	)

	m.subs = store.NewSubscriptions()
	validators := store.NewValidatorRegistry()
	for _, pv := range m.pendingValidators {
		if err := validators.Register(pv.pattern, pv.fn); err != nil {
			return fmt.Errorf("websocket: failed to register validator for pattern %q: %w", pv.pattern, err)
		}
	}
	m.validators = validators
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

	authMiddleware := wsauth.Middleware(m.config.Security.JWTSecret, m.config.Security.AuthRequired)
	if m.claimsResolver != nil {
		authMiddleware = wsauth.MiddlewareWithResolver(m.claimsResolver, m.config.Security.AuthRequired)
	}
	ginRouter.GET("/ws",
		authMiddleware,
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

// Validators returns the topic validator registry, populated from any
// WithValidator options after Initialize has run.
func (m *Module) Validators() *store.ValidatorRegistry { return m.validators }
