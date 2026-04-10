package connections

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/volvlabs/nebularcore/modules/websocket/protocol"
)

// Connection represents an active WebSocket connection.
type Connection interface {
	ID() string
	UserID() string
	TenantID() string
	Send(msg *protocol.ServerMessage) bool
	Close()
	Context() context.Context
}

// conn is the default Connection implementation.
type conn struct {
	id       string
	userID   string
	tenantID string
	writes   chan *protocol.ServerMessage
	ctx      context.Context
	cancel   context.CancelFunc
	once     sync.Once

	// WriteFn is called by the write goroutine to actually send data.
	// It is set by the handler after upgrade.
	WriteFn func(*protocol.ServerMessage) error
}

// NewConnection creates a new connection. The caller must set WriteFn and then
// call StartWriter before using Send.
func NewConnection(id, userID, tenantID string, parentCtx context.Context) *conn {
	ctx, cancel := context.WithCancel(parentCtx)
	return &conn{
		id:       id,
		userID:   userID,
		tenantID: tenantID,
		writes:   make(chan *protocol.ServerMessage, 256),
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (c *conn) ID() string               { return c.id }
func (c *conn) UserID() string           { return c.userID }
func (c *conn) TenantID() string         { return c.tenantID }
func (c *conn) Context() context.Context { return c.ctx }

// Send enqueues a message for writing. Returns false if the channel is full
// (message dropped) or the connection is closed.
func (c *conn) Send(msg *protocol.ServerMessage) bool {
	select {
	case c.writes <- msg:
		return true
	default:
		log.Warn().Str("conn_id", c.id).Msg("write channel full, dropping message")
		return false
	}
}

// Close idempotently cancels the context and drains the write channel.
func (c *conn) Close() {
	c.once.Do(func() {
		c.cancel()
		// Drain remaining messages so the write goroutine can exit.
		for {
			select {
			case <-c.writes:
			default:
				return
			}
		}
	})
}

// StartWriter runs the write goroutine. It blocks until the connection context
// is cancelled. Must be called in a separate goroutine.
func (c *conn) StartWriter() {
	defer func() {
		// Drain on exit.
		for {
			select {
			case <-c.writes:
			default:
				return
			}
		}
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.writes:
			if c.WriteFn != nil {
				if err := c.WriteFn(msg); err != nil {
					log.Err(err).Str("conn_id", c.id).Msg("write error, closing connection")
					c.Close()
					return
				}
			}
		}
	}
}

// Writes returns the write channel for direct access by the handler read loop
// (e.g. for the write goroutine).
func (c *conn) Writes() <-chan *protocol.ServerMessage {
	return c.writes
}
