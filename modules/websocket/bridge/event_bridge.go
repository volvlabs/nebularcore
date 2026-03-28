package bridge

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"
	"gitlab.com/jideobs/nebularcore/modules/event"
	"gitlab.com/jideobs/nebularcore/modules/websocket/connections"
	"gitlab.com/jideobs/nebularcore/modules/websocket/protocol"
	"gitlab.com/jideobs/nebularcore/modules/websocket/store"
)

// EventBridge subscribes to the event.Bus and fans out matching events to
// connected WebSocket clients.
type EventBridge struct {
	bus             event.Bus
	manager         *connections.Manager
	subscriptions   *store.Subscriptions
	allowedPatterns []string
}

// NewEventBridge creates a new EventBridge.
func NewEventBridge(bus event.Bus, manager *connections.Manager, subs *store.Subscriptions, allowedPatterns []string) *EventBridge {
	return &EventBridge{
		bus:             bus,
		manager:         manager,
		subscriptions:   subs,
		allowedPatterns: allowedPatterns,
	}
}

// Start subscribes to all allowed event patterns on the event bus. For each
// incoming event, it fans out to all WebSocket clients subscribed to a matching
// topic.
func (b *EventBridge) Start(ctx context.Context) error {
	for _, pattern := range b.allowedPatterns {
		p := pattern // capture
		if err := b.bus.Subscribe(p, func(ctx context.Context, msg event.Message) error {
			b.fanout(msg.EventType, msg.Payload)
			return nil
		}); err != nil {
			return err
		}
		log.Info().Str("pattern", p).Msg("event bridge subscribed")
	}
	return nil
}

// fanout sends an event to all WebSocket connections subscribed to a matching
// topic.
func (b *EventBridge) fanout(eventType string, payload []byte) {
	connIDs := b.subscriptions.GetSubscribedConns(eventType)
	if len(connIDs) == 0 {
		return
	}

	serverMsg := protocol.NewEventMsg(eventType, json.RawMessage(payload))

	for _, connID := range connIDs {
		c := b.manager.Get(connID)
		if c == nil {
			continue
		}
		c.Send(serverMsg)
	}
}
