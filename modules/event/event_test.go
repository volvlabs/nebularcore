package event

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestPayload struct {
	Data string `json:"data"`
}

func setupTestModule(t *testing.T) *Module {
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{},
		watermill.NewStdLogger(false, false),
	)

	router, err := message.NewRouter(message.RouterConfig{}, watermill.NewStdLogger(false, false))
	require.NoError(t, err)

	return &Module{
		router:     router,
		publisher:  pubSub,
		subscriber: pubSub,
		logger:     watermill.NewStdLogger(false, false),
	}
}

func TestNewEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		source    string
		payload   interface{}
		wantErr   bool
	}{
		{
			name:      "valid event with string payload",
			eventType: "test.event",
			source:    "test-source",
			payload:   "test payload",
			wantErr:   false,
		},
		{
			name:      "valid event with struct payload",
			eventType: "test.event",
			source:    "test-source",
			payload:   TestPayload{Data: "test data"},
			wantErr:   false,
		},
		{
			name:      "valid event with nil payload",
			eventType: "test.event",
			source:    "test-source",
			payload:   nil,
			wantErr:   false,
		},
		{
			name:      "invalid payload (channel - cannot be marshaled)",
			eventType: "test.event",
			source:    "test-source",
			payload:   make(chan int),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := NewMessage(tt.eventType, tt.source, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.eventType, event.EventType)
			assert.Equal(t, tt.source, event.Source)
			assert.NotNil(t, event.Message)
			assert.NotEmpty(t, event.UUID)
			assert.WithinDuration(t, time.Now(), event.Timestamp, 2*time.Second)

			if tt.payload != nil {
				expectedPayload, _ := json.Marshal(tt.payload)
				actualPayload := []byte(event.Message.Payload)
				assert.Equal(t, expectedPayload, actualPayload)
			}
		})
	}
}

func TestModule_Publish(t *testing.T) {
	module := setupTestModule(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		event   Message
		wantErr bool
	}{
		{
			name: "valid event",
			event: Message{
				Message:   message.NewMessage(watermill.NewUUID(), []byte("test")),
				Source:    "test-source",
				EventType: "test.event",
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "nil message",
			event: Message{
				Message:   nil,
				Source:    "test-source",
				EventType: "test.event",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "empty metadata",
			event: Message{
				Message:   message.NewMessage(watermill.NewUUID(), []byte("test")),
				Source:    "test-source",
				EventType: "test.event",
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := module.Publish(ctx, tt.event)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestModule_PublishAsync(t *testing.T) {
	module := setupTestModule(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		event   Message
		wantErr bool
	}{
		{
			name: "valid event",
			event: Message{
				Message:   message.NewMessage(watermill.NewUUID(), []byte("test")),
				Source:    "test-source",
				EventType: "test.event",
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "nil message",
			event: Message{
				Message:   nil,
				Source:    "test-source",
				EventType: "test.event",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errCh, err := module.PublishAsync(ctx, tt.event)
			require.NoError(t, err)

			err = <-errCh
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestModule_Subscribe(t *testing.T) {
	ctx := context.Background()

	eventType := "test.event"
	source := "test-source"
	payload := TestPayload{Data: "test data"}

	event, err := NewMessage(eventType, source, payload)
	require.NoError(t, err)

	tests := []struct {
		name      string
		handler   Handler
		wantErr   bool
		checkFunc func(t *testing.T, receivedEvent Message)
	}{
		{
			name: "successful handler",
			handler: func(ctx context.Context, event Message) error {
				return nil
			},
			wantErr: false,
			checkFunc: func(t *testing.T, receivedEvent Message) {
				assert.Equal(t, eventType, receivedEvent.EventType)
				assert.Equal(t, source, receivedEvent.Source)

				var receivedPayload TestPayload
				err := json.Unmarshal(receivedEvent.Message.Payload, &receivedPayload)
				require.NoError(t, err)
				assert.Equal(t, payload.Data, receivedPayload.Data)
			},
		},
		{
			name: "handler returns error",
			handler: func(ctx context.Context, event Message) error {
				return assert.AnError
			},
			wantErr: true,
			checkFunc: func(t *testing.T, receivedEvent Message) {
				assert.Equal(t, eventType, receivedEvent.EventType)
				assert.Equal(t, source, receivedEvent.Source)
			},
		},
		{
			name: "handler with invalid timestamp",
			handler: func(ctx context.Context, event Message) error {
				return nil
			},
			wantErr: false,
			checkFunc: func(t *testing.T, receivedEvent Message) {
				// Compare timestamps truncated to seconds and ensure both are in UTC
				expectedTime := event.Timestamp.UTC().Truncate(time.Second)
				actualTime := receivedEvent.Timestamp.UTC().Truncate(time.Second)
				assert.True(t, expectedTime.Equal(actualTime),
					"timestamps should be equal (expected: %s, actual: %s)",
					expectedTime.Format(time.RFC3339),
					actualTime.Format(time.RFC3339))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module := setupTestModule(t)
			defer func() { _ = module.router.Close() }()

			eventCh := make(chan Message, 1)
			errCh := make(chan error, 1)
			handlerWrapper := func(ctx context.Context, e Message) error {
				eventCh <- e
				err := tt.handler(ctx, e)
				if err != nil {
					errCh <- err
				}
				return err
			}

			// Subscribe
			err := module.Subscribe(eventType, handlerWrapper)
			require.NoError(t, err)

			// Start router
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			go func() {
				if err := module.router.Run(ctx); err != nil && err != context.Canceled {
					t.Errorf("router run error: %v", err)
				}
			}()

			// Wait for router to start and stabilize
			time.Sleep(50 * time.Millisecond)

			// Publish test event
			err = module.Publish(ctx, event)
			require.NoError(t, err)

			// Wait for event to be handled
			select {
			case receivedEvent := <-eventCh:
				tt.checkFunc(t, receivedEvent)
				// Wait for potential error
				select {
				case err := <-errCh:
					if tt.wantErr {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}
				case <-time.After(100 * time.Millisecond):
					if tt.wantErr {
						t.Error("expected error but got none")
					}
				}
			case <-time.After(time.Second):
				t.Fatal("timeout waiting for event")
			}
		})
	}
}

func TestModule_Unsubscribe(t *testing.T) {
	module := setupTestModule(t)
	eventType := "test.event"

	// Test unsubscribe (currently a no-op)
	err := module.Unsubscribe(eventType)
	assert.NoError(t, err)
}
