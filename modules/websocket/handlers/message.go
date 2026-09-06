package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rs/zerolog/log"
	"github.com/volvlabs/nebularcore/modules/event"
	"github.com/volvlabs/nebularcore/modules/websocket/bridge"
	"github.com/volvlabs/nebularcore/modules/websocket/connections"
	"github.com/volvlabs/nebularcore/modules/websocket/protocol"
	"github.com/volvlabs/nebularcore/modules/websocket/store"
)

// RegisterMessageHandlers wires up all client message type handlers on the
// given router. The evBridge parameter is optional; when non-nil, client
// subscriptions automatically create corresponding event bus subscriptions so
// that events published by other modules are forwarded to WebSocket clients.
func RegisterMessageHandlers(
	router *bridge.Router,
	manager *connections.Manager,
	subs *store.Subscriptions,
	bus event.Bus,
	maxTopicsPerConn int,
	validators *store.ValidatorRegistry,
	evBridge *bridge.EventBridge,
) {
	router.Register(protocol.TypeSubscribe, subscribeHandler(manager, subs, maxTopicsPerConn, validators, evBridge))
	router.Register(protocol.TypeUnsubscribe, unsubscribeHandler(subs))
	router.Register(protocol.TypePublish, publishHandler(manager, bus, validators))
	router.Register(protocol.TypePing, pingHandler(manager))
}

func subscribeHandler(manager *connections.Manager, subs *store.Subscriptions, maxTopics int, validators *store.ValidatorRegistry, evBridge *bridge.EventBridge) bridge.MessageHandler {
	return func(ctx context.Context, connID string, msg *protocol.Message) error {
		c := manager.Get(connID)

		if validators != nil && c != nil {
			if err := validators.Validate(ctx, c, msg.Topic); err != nil {
				c.Send(protocol.NewErrorMsg(msg.ID, err.Error()))
				return nil
			}
		}

		topics := subs.GetTopics(connID)
		if maxTopics > 0 && len(topics) >= maxTopics {
			if c != nil {
				c.Send(protocol.NewErrorMsg(msg.ID, "max topic subscriptions reached"))
			}
			return nil
		}

		subs.Subscribe(connID, msg.Topic)

		if evBridge != nil {
			if err := evBridge.SubscribeTopic(msg.Topic); err != nil {
				log.Warn().Err(err).Str("topic", msg.Topic).Msg("event bridge dynamic subscribe failed")
			}
		}

		if c != nil {
			c.Send(protocol.NewSubscribedMsg(msg.ID, msg.Topic))
		}

		log.Debug().Str("conn_id", connID).Str("topic", msg.Topic).Msg("subscribed")
		return nil
	}
}

func unsubscribeHandler(subs *store.Subscriptions) bridge.MessageHandler {
	return func(ctx context.Context, connID string, msg *protocol.Message) error {
		subs.Unsubscribe(connID, msg.Topic)
		log.Debug().Str("conn_id", connID).Str("topic", msg.Topic).Msg("unsubscribed")
		return nil
	}
}

func publishHandler(manager *connections.Manager, bus event.Bus, validators *store.ValidatorRegistry) bridge.MessageHandler {
	return func(ctx context.Context, connID string, msg *protocol.Message) error {
		c := manager.Get(connID)

		if validators != nil && c != nil {
			if err := validators.Validate(ctx, c, msg.Topic); err != nil {
				c.Send(protocol.NewErrorMsg(msg.ID, err.Error()))
				return nil
			}
		}

		userID := ""
		if c != nil {
			userID = c.UserID()
		}

		payload, _ := json.Marshal(map[string]any{
			"sender":  connID,
			"user_id": userID,
			"data":    msg.Payload,
		})

		watermillMsg := message.NewMessage(msg.ID, payload)
		watermillMsg.Metadata.Set("source", "websocket")
		watermillMsg.Metadata.Set("event_type", msg.Topic)
		watermillMsg.Metadata.Set("conn_id", connID)

		return bus.Publish(ctx, event.Message{
			Message:   watermillMsg,
			Source:    "websocket",
			EventType: msg.Topic,
			Timestamp: time.Now().UTC(),
		})
	}
}

func pingHandler(manager *connections.Manager) bridge.MessageHandler {
	return func(ctx context.Context, connID string, msg *protocol.Message) error {
		c := manager.Get(connID)
		if c != nil {
			c.Send(protocol.NewPongMsg(msg.ID))
		}
		return nil
	}
}
