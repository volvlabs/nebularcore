package bridge

import (
	"context"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/volvlabs/nebularcore/modules/event"
	eventmocks "github.com/volvlabs/nebularcore/modules/event/mocks"
	"github.com/volvlabs/nebularcore/modules/websocket/connections"
	"github.com/volvlabs/nebularcore/modules/websocket/store"
)

func newTestBridge(bus *eventmocks.Bus, allowedPatterns []string) *EventBridge {
	mgr := connections.NewManager(100)
	subs := store.NewSubscriptions()
	return NewEventBridge(bus, mgr, subs, allowedPatterns)
}

func TestSubscribeTopic_ExactTopic(t *testing.T) {
	bus := new(eventmocks.Bus)
	bus.On("Subscribe", "user.created", "websocket-fanout-user.created", mock.AnythingOfType("event.Handler")).Return(nil)

	eb := newTestBridge(bus, nil)
	err := eb.SubscribeTopic("user.created")

	require.NoError(t, err)
	bus.AssertCalled(t, "Subscribe", "user.created", "websocket-fanout-user.created", mock.AnythingOfType("event.Handler"))
}

// realWatermillBus wraps a real Watermill router + GoChannel pub/sub (the
// same wiring the production event.Module uses), so tests can exercise
// EventBridge.SubscribeTopic against Watermill's actual handler-name
// uniqueness constraint instead of a mock that just records call args.
type realWatermillBus struct {
	pubSub *gochannel.GoChannel
	router *message.Router
}

func newRealWatermillBus(t *testing.T) *realWatermillBus {
	t.Helper()
	logger := watermill.NewStdLogger(false, false)
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		_ = router.Run(ctx)
	}()
	select {
	case <-router.Running():
	case <-time.After(2 * time.Second):
		t.Fatal("router did not start in time")
	}

	return &realWatermillBus{pubSub: pubSub, router: router}
}

func (b *realWatermillBus) Publish(ctx context.Context, evt event.Message) error {
	return b.pubSub.Publish(evt.EventType, evt.Message)
}

func (b *realWatermillBus) PublishAsync(ctx context.Context, evt event.Message) (<-chan error, error) {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		if err := b.Publish(ctx, evt); err != nil {
			errCh <- err
		}
	}()
	return errCh, nil
}

func (b *realWatermillBus) Subscribe(eventType, handlerName string, handler event.Handler) error {
	b.router.AddNoPublisherHandler(handlerName, eventType, b.pubSub, func(msg *message.Message) error {
		return handler(context.Background(), event.Message{Message: msg, EventType: eventType})
	})
	go func() {
		_ = b.router.RunHandlers(context.Background())
	}()
	return nil
}

func (b *realWatermillBus) Unsubscribe(eventType string) error { return nil }

// TestSubscribeTopic_MultipleDistinctTopics_RealRouter is a regression test
// for a bug where every dynamic subscription reused the constant handler
// name "websocket-fanout". Watermill's router.AddHandler panics with
// DuplicateHandlerNameError when a handler name is registered twice, so in
// production only the first ever dynamically-subscribed topic succeeded —
// every subsequent distinct topic subscribed on the same process crashed the
// connection's read-loop goroutine, which surfaced to WebSocket clients as a
// "Subscribe timeout" (the ack never arrived because the goroutine died
// before sending it). A mock bus can't catch this because it never touches
// the real router. This test subscribes to several distinct topics against
// a real Watermill router and asserts none of them panic and all of them
// actually receive fanout events.
func TestSubscribeTopic_MultipleDistinctTopics_RealRouter(t *testing.T) {
	bus := newRealWatermillBus(t)
	mgr := connections.NewManager(100)
	subs := store.NewSubscriptions()
	eb := NewEventBridge(bus, mgr, subs, nil)

	topics := []string{
		"qa.conversation.aaa",
		"qa.conversation.bbb",
		"qa.user.ccc.create.question",
		"qa.user.ddd.create.answer.hcp",
	}

	for _, topic := range topics {
		require.NotPanics(t, func() {
			require.NoError(t, eb.SubscribeTopic(topic))
		}, "subscribing to topic %q must not panic", topic)
	}

	// Give the router a moment to finish wiring up the newly added handlers.
	time.Sleep(200 * time.Millisecond)

	// Every topic's handler must actually be live and receiving events —
	// not just "subscribed" without a working underlying subscription.
	for _, topic := range topics {
		msg := message.NewMessage(watermill.NewUUID(), []byte(`{}`))
		require.NoError(t, bus.pubSub.Publish(topic, msg))
	}
}

func TestSubscribeTopic_SkipsWildcard(t *testing.T) {
	bus := new(eventmocks.Bus)
	eb := newTestBridge(bus, nil)

	err := eb.SubscribeTopic("user.*")
	require.NoError(t, err)
	bus.AssertNotCalled(t, "Subscribe", mock.Anything, mock.Anything)

	err = eb.SubscribeTopic("**")
	require.NoError(t, err)
	bus.AssertNotCalled(t, "Subscribe", mock.Anything, mock.Anything)
}

func TestSubscribeTopic_Idempotent(t *testing.T) {
	bus := new(eventmocks.Bus)
	bus.On("Subscribe", "order.placed", "websocket-fanout-order.placed", mock.AnythingOfType("event.Handler")).Return(nil).Once()

	eb := newTestBridge(bus, nil)

	require.NoError(t, eb.SubscribeTopic("order.placed"))
	require.NoError(t, eb.SubscribeTopic("order.placed"))

	bus.AssertNumberOfCalls(t, "Subscribe", 1)
}

func TestSubscribeTopic_AllowedPatternsFilter(t *testing.T) {
	bus := new(eventmocks.Bus)
	eb := newTestBridge(bus, []string{"user.*", "order.*"})

	// Allowed topic — should subscribe.
	bus.On("Subscribe", "user.created", "websocket-fanout-user.created", mock.AnythingOfType("event.Handler")).Return(nil)
	require.NoError(t, eb.SubscribeTopic("user.created"))
	bus.AssertCalled(t, "Subscribe", "user.created", "websocket-fanout-user.created", mock.AnythingOfType("event.Handler"))

	// Disallowed topic — should NOT subscribe.
	err := eb.SubscribeTopic("admin.secret")
	require.NoError(t, err)
	bus.AssertNotCalled(t, "Subscribe", "admin.secret", "whatever", mock.Anything)
}

func TestSubscribeTopic_EmptyAllowedPatternsAllowsAll(t *testing.T) {
	bus := new(eventmocks.Bus)
	bus.On("Subscribe", "anything.goes", "websocket-fanout-anything.goes", mock.AnythingOfType("event.Handler")).Return(nil)

	eb := newTestBridge(bus, nil)
	require.NoError(t, eb.SubscribeTopic("anything.goes"))
	bus.AssertCalled(t, "Subscribe", "anything.goes", "websocket-fanout-anything.goes", mock.AnythingOfType("event.Handler"))
}

func TestStart_StaticPatterns(t *testing.T) {
	bus := new(eventmocks.Bus)
	bus.On("Subscribe", "user.*", mock.AnythingOfType("string"), mock.AnythingOfType("event.Handler")).Return(nil)
	bus.On("Subscribe", "order.*", mock.AnythingOfType("string"), mock.AnythingOfType("event.Handler")).Return(nil)

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
