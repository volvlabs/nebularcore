package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

func (m *Module) Publish(ctx context.Context, event Message) error {
	if event.Message == nil {
		return fmt.Errorf("event message cannot be nil")
	}
	if event.Metadata == nil {
		event.Metadata = message.Metadata{}
	}

	event.Metadata.Set("source", event.Source)
	event.Metadata.Set("event_type", event.EventType)
	event.Metadata.Set("timestamp", event.Timestamp.Format(time.RFC3339))

	return m.publisher.Publish(event.EventType, event.Message)
}

func (m *Module) PublishAsync(ctx context.Context, event Message) (<-chan error, error) {
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		if err := m.Publish(ctx, event); err != nil {
			errCh <- err
		}
	}()

	return errCh, nil
}

func (m *Module) Subscribe(eventType string, handler Handler) error {
	m.router.AddNoPublisherHandler(
		eventType,
		eventType,
		m.subscriber,
		func(msg *message.Message) error {
			event := Message{
				Message:   msg,
				Source:    msg.Metadata.Get("source"),
				EventType: msg.Metadata.Get("event_type"),
			}

			timestampStr := msg.Metadata.Get("timestamp")
			if timestampStr != "" {
				if ts, err := time.Parse(time.RFC3339, timestampStr); err == nil {
					event.Timestamp = ts
				}
			}

			if err := handler(context.Background(), event); err != nil {
				m.logger.Error("Event handler error", err, watermill.LogFields{
					"event_type": event.EventType,
					"source":     event.Source,
				})
				return err
			}

			return nil
		},
	)

	// If the router is already running, start the newly registered handler
	// goroutine immediately so it subscribes to the GoChannel before the
	// first publish on this topic arrives.
	if m.runCtx != nil {
		go func() {
			if err := m.router.RunHandlers(m.runCtx); err != nil {
				m.logger.Error("Error running handlers", err, nil)
			}
		}()
	}

	return nil
}

func (m *Module) Unsubscribe(eventType string) error {
	// m.router.RemoveHandler(eventType)
	return nil
}

func (m *Module) RunHandlers(ctx context.Context) error {
	return m.router.RunHandlers(ctx)
}

func (m *Module) Run(ctx context.Context) error {
	m.runCtx = ctx
	return m.router.Run(ctx)
}

func NewMessage(eventType, source string, payload any) (Message, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return Message{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	msg := message.NewMessage(
		watermill.NewUUID(),
		payloadBytes,
	)

	return Message{
		Message:   msg,
		Source:    source,
		EventType: eventType,
		Timestamp: time.Now().UTC(),
	}, nil
}
