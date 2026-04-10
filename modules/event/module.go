package event

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/gin-gonic/gin"
	"github.com/volvlabs/nebularcore/core/migration_runner"
	"github.com/volvlabs/nebularcore/core/module"
	"gorm.io/gorm"
)

// Module implements the event system using Watermill
type Module struct {
	publisher  message.Publisher
	subscriber message.Subscriber
	router     *message.Router
	logger     watermill.LoggerAdapter

	runCtx context.Context
}

// New creates a new event module
func New() (*Module, error) {
	logger := watermill.NewStdLogger(false, false)

	pubSub := gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer: 1024,
			Persistent:          false,
		},
		logger,
	)

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	return &Module{
		publisher:  pubSub,
		subscriber: pubSub,
		router:     router,
		logger:     logger,
	}, nil
}

// Configure implements module.Module.
func (m *Module) Configure(config any) error {
	return nil
}

// Dependencies implements module.Module.
func (m *Module) Dependencies() []string {
	return nil
}

// GetMigrationSources implements module.Module.
func (m *Module) GetMigrationSources(projectRoot string) []migration_runner.Source {
	return nil
}

// MigrationsDir implements module.Module.
func (m *Module) MigrationsDir() string {
	return ""
}

// Namespace implements module.Module.
func (m *Module) Namespace() module.ModuleNamespace {
	return module.PublicNamespace
}

// NewConfig implements module.Module.
func (m *Module) NewConfig() any {
	return nil
}

// ProvidesMigrations implements module.Module.
func (m *Module) ProvidesMigrations() bool {
	return false
}

// Version implements module.Module.
func (m *Module) Version() string {
	return "1.0.0"
}

// Name returns the module name
func (m *Module) Name() string {
	return "event"
}

// Initialize implements core.Module interface
func (m *Module) Initialize(ctx context.Context, db *gorm.DB, router *gin.Engine) error {
	return nil
}

// Shutdown implements core.Module interface
func (m *Module) Shutdown(ctx context.Context) error {
	return m.router.Close()
}
