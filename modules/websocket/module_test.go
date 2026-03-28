package websocket

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"gitlab.com/jideobs/nebularcore/core/module"
	"gitlab.com/jideobs/nebularcore/modules/event"
	wsconfig "gitlab.com/jideobs/nebularcore/modules/websocket/config"
)

// mockBus implements event.Bus for testing.
type mockBus struct {
	subscribed map[string]bool
}

func newMockBus() *mockBus {
	return &mockBus{subscribed: make(map[string]bool)}
}

func (b *mockBus) Publish(ctx context.Context, evt event.Message) error {
	return nil
}

func (b *mockBus) PublishAsync(ctx context.Context, evt event.Message) (<-chan error, error) {
	ch := make(chan error, 1)
	close(ch)
	return ch, nil
}

func (b *mockBus) Subscribe(eventType string, handler event.Handler) error {
	b.subscribed[eventType] = true
	return nil
}

func (b *mockBus) Unsubscribe(eventType string) error {
	delete(b.subscribed, eventType)
	return nil
}

func TestModuleInterface(t *testing.T) {
	bus := newMockBus()
	m := New(bus)

	var _ module.Module = m

	assert.Equal(t, "websocket", m.Name())
	assert.Equal(t, "1.0.0", m.Version())
	assert.Equal(t, []string{"event"}, m.Dependencies())
	assert.Equal(t, module.PublicNamespace, m.Namespace())
	assert.False(t, m.ProvidesMigrations())
	assert.Empty(t, m.MigrationsDir())
	assert.Nil(t, m.GetMigrationSources(""))
}

func TestModuleNewConfig(t *testing.T) {
	bus := newMockBus()
	m := New(bus)

	cfg := m.NewConfig()
	require.NotNil(t, cfg)
	wsCfg, ok := cfg.(*wsconfig.Config)
	require.True(t, ok)
	assert.Equal(t, int64(100000), wsCfg.Server.MaxConnections)
}

func TestModuleConfigure(t *testing.T) {
	bus := newMockBus()
	m := New(bus)

	t.Run("valid config", func(t *testing.T) {
		cfg := wsconfig.DefaultConfig()
		cfg.Enabled = true
		err := m.Configure(cfg)
		assert.NoError(t, err)
	})

	t.Run("invalid config type", func(t *testing.T) {
		err := m.Configure("not a config")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid config type")
	})

	t.Run("validation error", func(t *testing.T) {
		cfg := wsconfig.DefaultConfig()
		cfg.Enabled = true
		cfg.Server.Port = ""
		err := m.Configure(cfg)
		assert.Error(t, err)
	})
}

func TestModuleInitializeDisabled(t *testing.T) {
	bus := newMockBus()
	m := New(bus)

	err := m.Initialize(context.Background(), &gorm.DB{}, gin.New())
	assert.NoError(t, err)
	assert.Nil(t, m.Manager())
}

func TestModuleShutdownNoop(t *testing.T) {
	bus := newMockBus()
	m := New(bus)

	err := m.Shutdown(context.Background())
	assert.NoError(t, err)
}
