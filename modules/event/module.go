package event

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	watermillkafka "github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/gin-gonic/gin"
	"github.com/volvlabs/nebularcore/core/config"
	"github.com/volvlabs/nebularcore/core/migration_runner"
	"github.com/volvlabs/nebularcore/core/module"
	evtconfig "github.com/volvlabs/nebularcore/modules/event/config"
	"github.com/volvlabs/nebularcore/modules/event/types"
	"gorm.io/gorm"
)

// Module implements the event system using Watermill
type Module struct {
	publisher  message.Publisher
	subscriber message.Subscriber
	router     *message.Router
	logger     watermill.LoggerAdapter

	cfg *evtconfig.Config

	runCtx context.Context
}

// New creates a new event module
func New() (*Module, error) {
	logger := watermill.NewStdLogger(false, false)

	return &Module{
		logger: logger,
	}, nil
}

// Configure implements module.Module.
func (m *Module) Configure(config config.Config) error {
	cfg, ok := config.(*evtconfig.Config)
	if !ok {
		return fmt.Errorf("invalid configuration for event module")
	}

	m.cfg = cfg
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
func (m *Module) NewConfig() config.Config {
	return evtconfig.Default()
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

func (m *Module) createPubSub() (message.Publisher, message.Subscriber, error) {
	if m.cfg.Backend == types.BackendGoChannels {
		channel := gochannel.NewGoChannel(
			gochannel.Config{
				OutputChannelBuffer: 1024,
				Persistent:          false,
			},
			m.logger,
		)
		return channel, channel, nil
	}

	if m.cfg.Backend == types.BackendKafka {
		kafkaPublisher, err := watermillkafka.NewPublisher(
			watermillkafka.PublisherConfig{
				Brokers:   m.cfg.Kafka.Brokers,
				Marshaler: watermillkafka.DefaultMarshaler{},
			},
			watermill.NewStdLogger(false, false),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating kafka publisher: %w", err)
		}

		kafkaSubscriber, err := watermillkafka.NewSubscriber(
			watermillkafka.SubscriberConfig{
				Brokers:     m.cfg.Kafka.Brokers,
				Unmarshaler: watermillkafka.DefaultMarshaler{},
			},
			watermill.NewStdLogger(false, false),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating kafka subscriber: %w", err)
		}
		return kafkaPublisher, kafkaSubscriber, nil
	}

	return nil, nil, fmt.Errorf("could not create pub-sub for watermill, unsupported, %s", m.cfg.Backend)
}

// Initialize implements core.Module interface
func (m *Module) Initialize(ctx context.Context, db *gorm.DB, router *gin.Engine) error {
	publisher, subscriber, err := m.createPubSub()
	if err != nil {
		return fmt.Errorf("error initializing event module: %w", err)
	}

	m.publisher = publisher
	m.subscriber = subscriber

	m.router, err = message.NewRouter(message.RouterConfig{}, m.logger)
	if err != nil {
		return err
	}
	return nil
}

// Shutdown implements core.Module interface
func (m *Module) Shutdown(ctx context.Context) error {
	return m.router.Close()
}
