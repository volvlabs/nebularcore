package bridge

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
	"gitlab.com/jideobs/nebularcore/modules/websocket/protocol"
)

// MessageHandler processes an incoming client message for a specific type.
type MessageHandler func(ctx context.Context, connID string, msg *protocol.Message) error

// Router dispatches incoming client messages to registered handlers by type.
type Router struct {
	mu       sync.RWMutex
	handlers map[protocol.MessageType]MessageHandler
}

// NewRouter creates a new Router.
func NewRouter() *Router {
	return &Router{
		handlers: make(map[protocol.MessageType]MessageHandler),
	}
}

// Register adds a handler for a message type.
func (r *Router) Register(msgType protocol.MessageType, handler MessageHandler) {
	r.mu.Lock()
	r.handlers[msgType] = handler
	r.mu.Unlock()
}

// Dispatch routes a parsed message to the registered handler.
// Returns false if no handler is registered for the message type.
func (r *Router) Dispatch(ctx context.Context, connID string, msg *protocol.Message) bool {
	r.mu.RLock()
	handler, ok := r.handlers[msg.Type]
	r.mu.RUnlock()

	if !ok {
		log.Warn().Str("type", string(msg.Type)).Msg("no handler registered for message type")
		return false
	}

	if err := handler(ctx, connID, msg); err != nil {
		log.Err(err).Str("type", string(msg.Type)).Str("conn_id", connID).Msg("handler error")
	}
	return true
}
