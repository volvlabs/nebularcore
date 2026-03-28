package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/rs/zerolog/log"
	"gitlab.com/jideobs/nebularcore/modules/event"
	"gitlab.com/jideobs/nebularcore/modules/websocket/bridge"
	"gitlab.com/jideobs/nebularcore/modules/websocket/connections"
	"gitlab.com/jideobs/nebularcore/modules/websocket/protocol"
	"gitlab.com/jideobs/nebularcore/modules/websocket/store"
)

// RegisterMessageHandlers wires up all client message type handlers on the
// given router.
func RegisterMessageHandlers(
	router *bridge.Router,
	manager *connections.Manager,
	subs *store.Subscriptions,
	bus event.Bus,
	maxTopicsPerConn int,
) {
	router.Register(protocol.TypeSubscribe, subscribeHandler(manager, subs, maxTopicsPerConn))
	router.Register(protocol.TypeUnsubscribe, unsubscribeHandler(subs))
	router.Register(protocol.TypePublish, publishHandler(manager, bus))
	router.Register(protocol.TypePing, pingHandler(manager))
}

func subscribeHandler(manager *connections.Manager, subs *store.Subscriptions, maxTopics int) bridge.MessageHandler {
	return func(ctx context.Context, connID string, msg *protocol.Message) error {
		topics := subs.GetTopics(connID)
		if maxTopics > 0 && len(topics) >= maxTopics {
			c := manager.Get(connID)
			if c != nil {
				c.Send(protocol.NewErrorMsg(msg.ID, "max topic subscriptions reached"))
			}
			return nil
		}

		subs.Subscribe(connID, msg.Topic)

		c := manager.Get(connID)
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

func publishHandler(manager *connections.Manager, bus event.Bus) bridge.MessageHandler {
	return func(ctx context.Context, connID string, msg *protocol.Message) error {
		c := manager.Get(connID)
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
