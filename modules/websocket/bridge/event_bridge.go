package bridge

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

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

	subscribedTopics sync.Map
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
		b.subscribedTopics.Store(p, struct{}{})
		log.Info().Str("pattern", p).Msg("event bridge subscribed")
	}
	return nil
}

// SubscribeTopic dynamically subscribes to an exact topic on the event bus so
// that events published by other modules are forwarded to WebSocket clients.
// Wildcard topics (containing * or **) are silently skipped because Watermill
// requires exact topic names — glob matching is handled at the WebSocket
// subscription store level.
// If allowedPatterns is configured, the topic must match at least one pattern.
func (b *EventBridge) SubscribeTopic(topic string) error {
	if strings.Contains(topic, "*") {
		log.Debug().Str("topic", topic).Msg("event bridge: wildcard topic skipped for dynamic subscribe")
		return nil
	}

	if _, loaded := b.subscribedTopics.LoadOrStore(topic, struct{}{}); loaded {
		log.Debug().Str("topic", topic).Msg("event bridge: topic already subscribed, skipping")
		return nil
	}

	// If allowedPatterns is configured, enforce it as a security filter.
	if len(b.allowedPatterns) > 0 && !protocol.MatchAny(b.allowedPatterns, topic) {
		b.subscribedTopics.Delete(topic)
		log.Debug().Str("topic", topic).Msg("event bridge: topic not allowed by configured patterns")
		return nil
	}

	if err := b.bus.Subscribe(topic, func(ctx context.Context, msg event.Message) error {
		b.fanout(msg.EventType, msg.Payload)
		return nil
	}); err != nil {
		b.subscribedTopics.Delete(topic)
		return err
	}

	log.Info().Str("topic", topic).Msg("event bridge dynamically subscribed")
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
