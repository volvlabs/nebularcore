package store

import (
	"context"
	"encoding/json"

	"gitlab.com/jideobs/nebularcore/modules/websocket/connections"
	"gitlab.com/jideobs/nebularcore/modules/websocket/protocol"
)

// Adapter implements the WebSocketAdapter interface, providing a high-level API
// for other modules to interact with WebSocket connections.
type Adapter struct {
	manager       *connections.Manager
	subscriptions *Subscriptions
}

// NewAdapter creates a new Adapter.
func NewAdapter(manager *connections.Manager, subs *Subscriptions) *Adapter {
	return &Adapter{
		manager:       manager,
		subscriptions: subs,
	}
}

// Broadcast sends a message to all connections subscribed to the given topic,
// optionally filtering by a predicate.
func (a *Adapter) Broadcast(ctx context.Context, topic string, payload json.RawMessage, filter func(connections.Connection) bool) {
	connIDs := a.subscriptions.GetSubscribedConns(topic)
	msg := protocol.NewEventMsg(topic, payload)

	for _, connID := range connIDs {
		c := a.manager.Get(connID)
		if c == nil {
			continue
		}
		if filter != nil && !filter(c) {
			continue
		}
		c.Send(msg)
	}
}

// SendTo sends a message to a specific connection by ID.
func (a *Adapter) SendTo(connID string, msg *protocol.ServerMessage) bool {
	c := a.manager.Get(connID)
	if c == nil {
		return false
	}
	return c.Send(msg)
}

// SendToUser sends a message to all connections of a specific user.
func (a *Adapter) SendToUser(userID string, msg *protocol.ServerMessage) int {
	conns := a.manager.GetByUser(userID)
	sent := 0
	for _, c := range conns {
		if c.Send(msg) {
			sent++
		}
	}
	return sent
}

// SendToTenant sends a message to all connections of a specific tenant.
func (a *Adapter) SendToTenant(tenantID string, msg *protocol.ServerMessage) int {
	conns := a.manager.GetByTenant(tenantID)
	sent := 0
	for _, c := range conns {
		if c.Send(msg) {
			sent++
		}
	}
	return sent
}

// GetConnections returns all connections matching a filter predicate.
func (a *Adapter) GetConnections(filter func(connections.Connection) bool) []connections.Connection {
	all := a.manager.GetAll()
	if filter == nil {
		return all
	}
	var result []connections.Connection
	for _, c := range all {
		if filter(c) {
			result = append(result, c)
		}
	}
	return result
}
