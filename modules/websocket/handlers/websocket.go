package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	wsauth "github.com/volvlabs/nebularcore/modules/websocket/auth"
	"github.com/volvlabs/nebularcore/modules/websocket/bridge"
	wsconfig "github.com/volvlabs/nebularcore/modules/websocket/config"
	"github.com/volvlabs/nebularcore/modules/websocket/connections"
	"github.com/volvlabs/nebularcore/modules/websocket/protocol"
	"github.com/volvlabs/nebularcore/modules/websocket/store"
)

// WebSocketHandler manages WebSocket upgrade and read loop.
type WebSocketHandler struct {
	config  *wsconfig.Config
	manager *connections.Manager
	pool    *connections.Pool
	subs    *store.Subscriptions
	router  *bridge.Router
}

// NewWebSocketHandler creates a WebSocketHandler.
func NewWebSocketHandler(
	cfg *wsconfig.Config,
	manager *connections.Manager,
	pool *connections.Pool,
	subs *store.Subscriptions,
	router *bridge.Router,
) *WebSocketHandler {
	return &WebSocketHandler{
		config:  cfg,
		manager: manager,
		pool:    pool,
		subs:    subs,
		router:  router,
	}
}

// Handle processes a WebSocket upgrade request.
func (h *WebSocketHandler) Handle(c *gin.Context) {
	claims := wsauth.GetClaims(c)

	userID := ""
	tenantID := ""
	orgID := ""
	if claims != nil {
		userID = claims.UserID
		tenantID = claims.TenantID
		orgID = claims.OrgID
	}

	if len(h.config.Security.AllowOrigins) > 0 {
		origin := c.GetHeader("Origin")
		if !wsauth.ValidateOrigin(origin, h.config.Security.AllowOrigins) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  false,
				"message": "origin not allowed",
			})
			return
		}
	}

	if err := h.pool.CheckLimits(userID, tenantID); err != nil {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	opts := &websocket.AcceptOptions{
		InsecureSkipVerify: len(h.config.Security.AllowOrigins) == 0,
		OriginPatterns:     h.config.Security.AllowOrigins,
	}

	ws, err := websocket.Accept(c.Writer, c.Request, opts)
	if err != nil {
		log.Err(err).Msg("websocket upgrade failed")
		return
	}

	connID := uuid.NewString()
	conn := connections.NewConnection(connID, userID, tenantID, orgID, c.Request.Context())

	conn.WriteFn = func(msg *protocol.ServerMessage) error {
		data, encErr := protocol.Encode(msg)
		if encErr != nil {
			return encErr
		}
		writeCtx, cancel := context.WithTimeout(conn.Context(), h.config.Server.WriteDeadline)
		defer cancel()
		return ws.Write(writeCtx, websocket.MessageText, data)
	}

	if !h.manager.Register(conn) {
		_ = ws.Close(websocket.StatusTryAgainLater, "server at capacity")
		return
	}

	log.Info().Str("conn_id", connID).Str("user_id", userID).Msg("websocket connected")

	go conn.StartWriter()

	if h.config.Server.PingInterval > 0 {
		go h.pingLoop(conn, ws)
	}

	h.readLoop(conn, ws)

	conn.Close()
	h.manager.Deregister(connID)
	h.subs.UnsubscribeAll(connID)

	_ = ws.Close(websocket.StatusNormalClosure, "connection closed")
	log.Info().Str("conn_id", connID).Msg("websocket disconnected")
}

// pingLoop periodically sends a WebSocket ping frame so idle connections
// (and any intermediate proxies/load balancers) are not torn down purely
// for lack of application traffic. Ping is safe to call concurrently with
// an in-progress Read, since coder/websocket relies on an active reader to
// observe the resulting pong frame.
func (h *WebSocketHandler) pingLoop(conn connections.Connection, ws *websocket.Conn) {
	ticker := time.NewTicker(h.config.Server.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-conn.Context().Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(conn.Context(), h.config.Server.PingInterval)
			err := ws.Ping(pingCtx)
			cancel()
			if err != nil {
				return
			}
		}
	}
}

func (h *WebSocketHandler) readLoop(conn connections.Connection, ws *websocket.Conn) {
	for {
		readCtx := conn.Context()
		var cancel context.CancelFunc
		if h.config.Server.ReadDeadline > 0 {
			readCtx, cancel = context.WithTimeout(conn.Context(), h.config.Server.ReadDeadline)
		}

		_, data, err := ws.Read(readCtx)
		if cancel != nil {
			cancel()
		}
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				return
			}
			if conn.Context().Err() != nil {
				return
			}
			log.Debug().Err(err).Str("conn_id", conn.ID()).Msg("read error")
			return
		}

		msg, parseErr := protocol.Parse(data, h.config.Routing.MaxTopicLength)
		if parseErr != nil {
			conn.Send(protocol.NewErrorMsg("", parseErr.Error()))
			continue
		}

		h.router.Dispatch(conn.Context(), conn.ID(), msg)
	}
}
