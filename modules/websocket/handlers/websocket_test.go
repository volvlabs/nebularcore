package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wsauth "github.com/volvlabs/nebularcore/modules/websocket/auth"
	"github.com/volvlabs/nebularcore/modules/websocket/bridge"
	wsconfig "github.com/volvlabs/nebularcore/modules/websocket/config"
	"github.com/volvlabs/nebularcore/modules/websocket/connections"
	"github.com/volvlabs/nebularcore/modules/websocket/protocol"
	"github.com/volvlabs/nebularcore/modules/websocket/store"
)

func setupTestServer(t *testing.T, authRequired bool) (*httptest.Server, *connections.Manager, *store.Subscriptions) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := wsconfig.DefaultConfig()
	cfg.Enabled = true
	cfg.Security.AuthRequired = authRequired
	cfg.Security.JWTSecret = "test-secret"

	manager := connections.NewManager(100)
	pool := connections.NewPool(manager, 10, 100)
	subs := store.NewSubscriptions()
	router := bridge.NewRouter()

	router.Register(protocol.TypePing, func(ctx context.Context, connID string, msg *protocol.Message) error {
		c := manager.Get(connID)
		if c != nil {
			c.Send(protocol.NewPongMsg(msg.ID))
		}
		return nil
	})

	handler := NewWebSocketHandler(cfg, manager, pool, subs, router)

	r.GET("/ws",
		wsauth.Middleware(cfg.Security.JWTSecret, cfg.Security.AuthRequired),
		handler.Handle,
	)

	srv := httptest.NewServer(r)
	t.Cleanup(func() { srv.Close() })

	return srv, manager, subs
}

func TestWebSocketUpgradeNoAuth(t *testing.T) {
	srv, manager, _ := setupTestServer(t, false)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ws, _, err := websocket.Dial(ctx, "ws"+srv.URL[4:]+"/ws", nil)
	require.NoError(t, err)
	defer ws.Close(websocket.StatusNormalClosure, "done")

	require.Eventually(t, func() bool {
		return manager.Total() == 1
	}, time.Second, 10*time.Millisecond)
}

func TestWebSocketAuthRequired(t *testing.T) {
	srv, _, _ := setupTestServer(t, true)

	resp, err := http.Get(srv.URL + "/ws")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWebSocketPingPong(t *testing.T) {
	srv, _, _ := setupTestServer(t, false)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ws, _, err := websocket.Dial(ctx, "ws"+srv.URL[4:]+"/ws", nil)
	require.NoError(t, err)
	defer ws.Close(websocket.StatusNormalClosure, "done")

	pingMsg := protocol.Message{
		ID:   "test-ping-1",
		Type: protocol.TypePing,
	}
	data, err := json.Marshal(pingMsg)
	require.NoError(t, err)

	err = ws.Write(ctx, websocket.MessageText, data)
	require.NoError(t, err)

	_, resp, err := ws.Read(ctx)
	require.NoError(t, err)

	var serverMsg protocol.ServerMessage
	err = json.Unmarshal(resp, &serverMsg)
	require.NoError(t, err)

	assert.Equal(t, protocol.TypePong, serverMsg.Type)
	assert.Equal(t, "test-ping-1", serverMsg.ID)
}

func TestWebSocketDisconnectCleanup(t *testing.T) {
	srv, manager, _ := setupTestServer(t, false)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ws, _, err := websocket.Dial(ctx, "ws"+srv.URL[4:]+"/ws", nil)
	require.NoError(t, err)

	assert.Equal(t, int64(1), manager.Total())

	ws.Close(websocket.StatusNormalClosure, "done")

	assert.Eventually(t, func() bool {
		return manager.Total() == 0
	}, 2*time.Second, 50*time.Millisecond)
}
