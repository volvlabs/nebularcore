package bridge

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	eventmocks "github.com/volvlabs/nebularcore/modules/event/mocks"
	"github.com/volvlabs/nebularcore/modules/websocket/connections"
	"github.com/volvlabs/nebularcore/modules/websocket/store"
)

func newTestBridge(bus *eventmocks.EventBus, allowedPatterns []string) *EventBridge {
	mgr := connections.NewManager(100)
	subs := store.NewSubscriptions()
	return NewEventBridge(bus, mgr, subs, allowedPatterns)
}

func TestSubscribeTopic_ExactTopic(t *testing.T) {
	bus := new(eventmocks.EventBus)
	bus.On("Subscribe", "user.created", mock.AnythingOfType("event.Handler")).Return(nil)

	eb := newTestBridge(bus, nil)
	err := eb.SubscribeTopic("user.created")

	require.NoError(t, err)
	bus.AssertCalled(t, "Subscribe", "user.created", mock.AnythingOfType("event.Handler"))
}

func TestSubscribeTopic_SkipsWildcard(t *testing.T) {
	bus := new(eventmocks.EventBus)
	eb := newTestBridge(bus, nil)

	err := eb.SubscribeTopic("user.*")
	require.NoError(t, err)
	bus.AssertNotCalled(t, "Subscribe", mock.Anything, mock.Anything)

	err = eb.SubscribeTopic("**")
	require.NoError(t, err)
	bus.AssertNotCalled(t, "Subscribe", mock.Anything, mock.Anything)
}

func TestSubscribeTopic_Idempotent(t *testing.T) {
	bus := new(eventmocks.EventBus)
	bus.On("Subscribe", "order.placed", mock.AnythingOfType("event.Handler")).Return(nil).Once()

	eb := newTestBridge(bus, nil)

	require.NoError(t, eb.SubscribeTopic("order.placed"))
	require.NoError(t, eb.SubscribeTopic("order.placed"))

	bus.AssertNumberOfCalls(t, "Subscribe", 1)
}

func TestSubscribeTopic_AllowedPatternsFilter(t *testing.T) {
	bus := new(eventmocks.EventBus)
	eb := newTestBridge(bus, []string{"user.*", "order.*"})

	// Allowed topic — should subscribe.
	bus.On("Subscribe", "user.created", mock.AnythingOfType("event.Handler")).Return(nil)
	require.NoError(t, eb.SubscribeTopic("user.created"))
	bus.AssertCalled(t, "Subscribe", "user.created", mock.AnythingOfType("event.Handler"))

	// Disallowed topic — should NOT subscribe.
	err := eb.SubscribeTopic("admin.secret")
	require.NoError(t, err)
	bus.AssertNotCalled(t, "Subscribe", "admin.secret", mock.Anything)
}

func TestSubscribeTopic_EmptyAllowedPatternsAllowsAll(t *testing.T) {
	bus := new(eventmocks.EventBus)
	bus.On("Subscribe", "anything.goes", mock.AnythingOfType("event.Handler")).Return(nil)

	eb := newTestBridge(bus, nil)
	require.NoError(t, eb.SubscribeTopic("anything.goes"))
	bus.AssertCalled(t, "Subscribe", "anything.goes", mock.AnythingOfType("event.Handler"))
}

func TestStart_StaticPatterns(t *testing.T) {
	bus := new(eventmocks.EventBus)
	bus.On("Subscribe", "user.*", mock.AnythingOfType("event.Handler")).Return(nil)
	bus.On("Subscribe", "order.*", mock.AnythingOfType("event.Handler")).Return(nil)

	eb := newTestBridge(bus, []string{"user.*", "order.*"})
	err := eb.Start(context.Background())

	require.NoError(t, err)
	bus.AssertNumberOfCalls(t, "Subscribe", 2)

	// After Start, these patterns should be tracked as subscribed.
	// Re-subscribing to the same patterns should be a no-op.
	assert.NoError(t, eb.SubscribeTopic("user.*"))
	assert.NoError(t, eb.SubscribeTopic("order.*"))
	bus.AssertNumberOfCalls(t, "Subscribe", 2)
}
