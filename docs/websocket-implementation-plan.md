# WebSocket Support Implementation Plan

## Executive Summary

Add WebSocket support to NebularCore as an **optional data transport adapter** that works alongside HTTP. This enables real-time, bidirectional communication for modules while maintaining the framework's modularity and opt-in philosophy.

**Target**: 100,000 concurrent WebSocket connections (assuming adequate RAM/CPU).
**Library**: `github.com/coder/websocket` — lower per-connection allocation and context-native API vs. gorilla/websocket, better suited for 100k-scale.

---

## 1. Architecture Overview

### 1.1 Design Principles

1. **Adapter Pattern**: WebSocket is a transport adapter, separate from HTTP, not replacing it
2. **Opt-in**: Disabled by default; users enable in configuration
3. **Module-agnostic**: Modules don't know about transport mechanism
4. **Data Flow-based**: Events/messages flow through a unified interface
5. **Event-driven**: Leverage existing event system for pub/sub
6. **Multi-tenancy ready**: Support tenant isolation via path parameters
7. **Authentication-aware**: Integrate with existing auth module
8. **100k-concurrent-ready**: Sharded connection manager; no global locks on hot paths

### 1.2 Core Concept: Data Transport Layer

```
┌─────────────────────────────────────────────────────┐
│                   Application Logic                  │
│          (Modules, Handlers, Business Logic)         │
└────────────┬────────────────────────────────┬────────┘
             │                                │
    ┌────────▼──────────┐          ┌─────────▼────────┐
    │  HTTP Transport   │          │ WebSocket        │
    │  (Gin Router)     │          │ Transport        │
    │                   │          │ (New Adapter)    │
    └────────┬──────────┘          └─────────┬────────┘
             │                                │
    ┌────────▼──────────────────────────────▼────────┐
    │    Unified Message/Event Interface             │
    │    (Shared across all transports)              │
    └──────────────────────────────────────────────┘
```

### 1.3 Key Components

1. **WebSocket Module** (`modules/websocket/module.go`) — lifecycle management, route registration, event bridge startup
2. **Message Protocol** (`modules/websocket/protocol/`) — standardized wire format, parsing, serialization
3. **Connection Manager** (`modules/websocket/connections/`) — sharded lock-free-on-read connection tracking, per-user/tenant indexes
4. **Event Bridge** (`modules/websocket/bridge/`) — Watermill to WebSocket fanout with topic filtering
5. **Auth Integration** (`modules/websocket/auth/`) — JWT validation on upgrade, origin allowlist
6. **Subscription Store** (`modules/websocket/store/`) — tracks which connections are subscribed to which topics
7. **WebSocketAdapter** (`modules/websocket/store/adapters.go`) — interface exposed to other modules

---

## 2. Data Flow Architecture

### 2.1 Event-to-WebSocket Flow

```
Module publishes event via event.Bus
         ↓
Event System (Watermill)
         ↓
Event Bridge subscribes to allowed topics
         ↓
Bridge checks client subscription store
         ↓
Bridge serializes event to WebSocket message
         ↓
Connection Manager routes to subscribed clients
         ↓
Per-connection write goroutine sends to WebSocket client
```

### 2.2 WebSocket-to-Module Flow

```
Client sends WebSocket message
         ↓
WebSocket Handler parses and validates message
         ↓
Validates format and permissions
         ↓
Publishes to event.Bus as internal event
         ↓
Modules subscribe to the event type via event.Bus
         ↓
Module processes via normal event handler
```

---

## 3. High-Concurrency Design (100k Connections)

### 3.1 Sharded Connection Manager

The connection manager uses **256 shards** to avoid global lock contention at scale:

```
Connection ID → fnv32a hash → % 256 → shard index → shard.RWMutex + map[id]*conn
```

- Read operations (lookups by connection ID) hold only the per-shard `RLock`
- Write operations (register/deregister) hold only the per-shard `Lock`
- `userIndex` and `tenantIndex` use `sync.Map` for O(1) reverse lookups
- Total counter uses `atomic.Int64` — no lock needed for connection count checks

### 3.2 Per-Connection Write Goroutine

Each connection spawns **one dedicated write goroutine** backed by a buffered channel:

```
conn.Send(msg) → non-blocking push to writes chan (cap 256)
                       ↓
             write goroutine drains channel
                       ↓
          coder/websocket.Write() (serialized, safe)
```

- `Send()` is non-blocking: drops message and logs if channel is full (explicit backpressure, no goroutine leak)
- `sync.Once` guards `Close()` for idempotency
- `context.Context` propagation: connection context cancelled on close signals the write goroutine to exit

### 3.3 Memory Efficiency

- `sync.Pool` for `bytes.Buffer` reuse in encoder (reduces GC pressure under broadcast load)
- Messages serialized once, broadcast by reference to multiple connections
- No per-connection heap allocation for read buffers (reused across reads via `coder/websocket` internals)

---

## 4. Core Implementation Details

### 4.1 Configuration Structure

```yaml
modules:
  websocket:
    enabled: true

    server:
      host: localhost
      port: "8080"              # Same port as HTTP (upgrade-based)
      readBufferSize: 4096
      writeBufferSize: 4096
      readDeadline: 60s
      writeDeadline: 60s
      maxConnections: 100000    # Total server limit
      maxConnectionsPerUser: 10
      maxConnectionsPerTenant: 50000

    routing:
      maxTopicLength: 256
      maxTopicsPerConnection: 100

    security:
      authRequired: true
      allowOrigins:
        - "http://localhost:3000"
        - "https://app.example.com"

    tenantMode: "header"        # header | path | query

    events:
      allowedEventTypes:
        - "user.*"
        - "notification.*"
      internalEventPrefix: "ws:"
```

### 4.2 Message Protocol

#### Client to Server

```json
{
  "id": "msg-uuid-1234",
  "type": "subscribe | publish | unsubscribe | ping | auth",
  "topic": "notifications.user:123",
  "payload": {},
  "timestamp": "2026-02-10T10:00:00Z"
}
```

**Types:**
- `subscribe` — Subscribe to a topic or glob pattern
- `publish` — Publish a message to a topic (from client)
- `unsubscribe` — Unsubscribe from a topic
- `ping` — Keepalive
- `auth` — (Re)authenticate with a token during a long-lived connection

#### Server to Client

```json
{
  "id": "msg-uuid-1234",
  "type": "subscribed | unsubscribed | message | error | pong | auth_success",
  "topic": "notifications.user:123",
  "payload": {},
  "timestamp": "2026-02-10T10:00:00Z"
}
```

**Types:**
- `subscribed` — Confirms subscription
- `unsubscribed` — Confirms unsubscription
- `message` — Event message from subscribed topic
- `error` — Error response
- `pong` — Keepalive response
- `auth_success` — Auth confirmation

### 4.3 WebSocketAdapter Interface (for modules)

```go
// Connection represents an active WebSocket connection
type Connection interface {
    ID() string
    UserID() string
    TenantID() string
    Send(ctx context.Context, msg *Message) error
    Close() error
    Context() context.Context
}

// MessageHandler processes incoming WebSocket messages
type MessageHandler func(ctx context.Context, conn Connection, msg *Message) error

// TopicFilter determines which connections receive a message
type TopicFilter func(conn Connection, topic string) bool

// WebSocketAdapter allows other modules to use WebSocket capabilities
type WebSocketAdapter interface {
    Subscribe(topic string, handler MessageHandler) error
    Broadcast(ctx context.Context, topic string, msg *Message, filter TopicFilter) error
    SendTo(ctx context.Context, connID string, msg *Message) error
    SendToUser(ctx context.Context, userID string, msg *Message) error
    SendToTenant(ctx context.Context, tenantID string, msg *Message) error
    RegisterHandler(msgType string, handler MessageHandler) error
    GetConnections(filter func(Connection) bool) []Connection
}
```

### 4.4 Module Structure

```
modules/websocket/
├── module.go                   # Module interface impl: Name, Dependencies, Initialize, Shutdown
├── config/
│   └── config.go              # Config struct, DefaultConfig(), Validate()
├── connections/
│   ├── connection.go          # Connection interface + conn impl (write goroutine, Send, Close)
│   ├── manager.go             # 256-shard manager, userIndex, tenantIndex, atomic counter
│   └── pool.go                # Per-user/tenant limit enforcement; bytes.Buffer sync.Pool
├── protocol/
│   ├── message.go             # MessageType enum, Message (client to server), ServerMessage structs
│   ├── parser.go              # Parse([]byte) — JSON decode + field validation
│   └── encoder.go             # Encode(*ServerMessage) + helper constructors
├── handlers/
│   ├── websocket.go           # Gin GET /ws: auth middleware → upgrade → register → read loop
│   └── message.go             # Read loop: subscribe/unsubscribe/publish/ping dispatch
├── bridge/
│   ├── event_bridge.go        # event.Bus subscriber → fanout to matching connections
│   ├── router.go              # MessageType → registered MessageHandler dispatch
│   └── filter.go              # MatchTopic(pattern, topic) glob matching
├── auth/
│   ├── validator.go           # Wraps tools/security/jwt.go ParseToken(); origin check
│   └── middleware.go          # Gin middleware: Bearer/query token → context
└── store/
    ├── subscriptions.go       # Subscribe/Unsubscribe/GetTopics/GetSubscribedConns (sync.Map)
    └── adapters.go            # Implements WebSocketAdapter
```

---

## 5. Integration Points

### 5.1 With Event Module

- **Dependency**: `Dependencies() []string{"event"}` — registry initializes event module first
- Event Bridge subscribes to `event.Bus` for configured `allowedEventTypes` (glob patterns)
- Per-event fanout: checks subscription store, sends to matching connections only
- Event payload serialized as `message` type on the WebSocket wire

### 5.2 With Auth Module

- JWT validation on the HTTP upgrade request (before WebSocket handshake completes)
- `tools/security/jwt.go` `ParseToken()` reused — no new JWT logic
- Extracts `userID` and `tenantID` from claims; stored on `Connection`
- Future: per-connection token refresh via `auth` message type

### 5.3 With Tenant Module

- Tenant extracted from header/path/query per `tenantMode` config
- Subscriptions and broadcasts automatically scoped to `tenantID`
- Per-tenant connection limits enforced at registration time
- Cross-tenant message delivery is impossible by design (tenant index is separate)

---

## 6. Testing Strategy

### 6.1 Unit Tests (run with `go test ./...`)

| Package | Coverage |
|---|---|
| `protocol/` | Parse: valid messages, malformed JSON, unknown type, oversized topic; Encode: round-trip, all helper constructors |
| `connections/` | Manager: 10k goroutines concurrent register/deregister (`go test -race`), shard distribution uniformity, user/tenant lookup correctness, limit enforcement; Connection: Send, backpressure drop on full channel, graceful close idempotency |
| `auth/` | Validator: valid JWT, expired JWT, wrong secret, missing `userID`/`tenantID` claims, origin allowlist (allowed + blocked) |
| `bridge/` | Filter: glob patterns (exact, wildcard, multi-segment); EventBridge: mock Bus + mock manager — event delivered to subscribed connections, not delivered to unsubscribed, tenant isolation |
| `module` | Configure/Initialize/Shutdown lifecycle with mock event.Bus; disabled module skips route registration |
| `handlers/` | httptest WebSocket upgrade, 401 on missing/invalid token, subscribe+message dispatch end-to-end |

All unit tests use testify mocks (matching `modules/event/mocks/` pattern) and must pass `go test -race`.

### 6.2 Load Tests (excluded from `go test ./...` via build tag)

Stored in `tests/load/` with `//go:build loadtest` — **never executed by `go test ./...`**.

#### Invocation Options

| Method | Command |
|---|---|
| Makefile | `make load-test` |
| Shell | `./tests/load/run.sh` |
| Go (CI-friendly) | `go test -tags loadtest -run TestWebSocketLoad ./tests/load/` |
| GitLab CI | `load_test` job (`when: manual` or scheduled) |

#### Load Profile (k6)

| Stage | Duration | Target VUs |
|---|---|---|
| Ramp up | 2 min | 0 to 10,000 |
| Ramp up | 5 min | 10,000 to 100,000 |
| Sustain | 10 min | 100,000 |
| Ramp down | 2 min | 100,000 to 0 |

Each VU: connect to subscribe to a topic to receive N messages to disconnect.

#### SLO Thresholds (k6)

- Connection success rate: **>99%**
- `ws_session_duration` p95: **<30s**
- Error rate: **<1%**

#### Load Test Infrastructure

```
tests/load/
├── testapp/
│   ├── main.go             # Minimal NebularCore app: SQLite, WebSocket enabled, auth off
│   │                       # Background goroutine publishes events every 100ms
│   ├── Dockerfile          # Multi-stage: golang:1.24-alpine builder to alpine runner
│   └── config.yml          # SQLite, WebSocket on :8080, auth disabled
├── k6/
│   └── websocket_load.js   # k6 script: ramp profile, thresholds, reads WS_HOST/WS_PORT env
├── docker-compose.yml      # Services: testapp (health-checked on /health) + k6 (grafana/k6)
│                           # k6 depends_on: testapp: condition: service_healthy
│                           # k6 writes result JSON to /results/ (mounted volume)
├── run.sh                  # docker compose up --build --exit-code-from k6
│                           # docker compose down on exit; propagates k6 exit code
└── load_test.go            # //go:build loadtest
                            # TestWebSocketLoad: runs run.sh via os/exec, asserts exit code 0
```

---

## 7. Implementation Phases

### Phase 1: Core Protocol
- [ ] Add `github.com/coder/websocket` to `go.mod`
- [ ] `protocol/message.go` — `MessageType` enum, `Message`, `ServerMessage` structs
- [ ] `protocol/parser.go` — `Parse([]byte) (*Message, error)` with JSON + field validation
- [ ] `protocol/encoder.go` — `Encode(*ServerMessage) ([]byte, error)` + helper constructors
- [ ] Unit tests: `protocol/message_test.go`, `protocol/parser_test.go`

**Deliverable**: Wire format defined, tested, and round-trip verified

### Phase 2: High-Concurrency Connection Manager
- [ ] `connections/connection.go` — write goroutine, buffered channel `Send()`, idempotent `Close()`
- [ ] `connections/manager.go` — 256-shard design, `userIndex`, `tenantIndex`, `atomic.Int64` counter
- [ ] `connections/pool.go` — per-user/tenant limit enforcement; `bytes.Buffer` `sync.Pool`
- [ ] Unit tests: `connections/manager_test.go` (race-safe, 10k goroutine concurrency); `connections/connection_test.go`

**Deliverable**: Connection layer capable of 100k concurrent connections with no data races

### Phase 3: Auth & Security (parallel with Phase 2)
- [ ] `auth/validator.go` — `ParseToken()` wrapper, origin allowlist check
- [ ] `auth/middleware.go` — Gin upgrade middleware (Bearer header + `?token=` query param)
- [ ] Unit tests: `auth/validator_test.go`

**Deliverable**: Secure, authenticated WebSocket upgrades

### Phase 4: Event Bridge (depends on Phases 2 and 3)
- [ ] `bridge/filter.go` — `MatchTopic(pattern, topic string) bool` glob matching
- [ ] `bridge/router.go` — `MessageType` to registered `MessageHandler` dispatch
- [ ] `bridge/event_bridge.go` — `event.Bus` subscriber to per-connection fanout
- [ ] Unit tests: `bridge/filter_test.go`, `bridge/event_bridge_test.go`

**Deliverable**: Events flow from Watermill bus to subscribed WebSocket clients

### Phase 5: Handlers & Module Assembly (depends on all prior phases)
- [ ] `store/subscriptions.go` — `Subscribe/Unsubscribe/GetTopics/GetSubscribedConns`
- [ ] `store/adapters.go` — implements `WebSocketAdapter`
- [ ] `handlers/websocket.go` — Gin `GET /ws` route handler
- [ ] `handlers/message.go` — read loop dispatcher
- [ ] `module.go` — full `Module` interface: `Name="websocket"`, `Dependencies=["event"]`, `Namespace=PublicNamespace`
- [ ] Unit tests: `module_test.go`, `handlers/websocket_test.go`

**Deliverable**: Fully functional WebSocket module, registerable with `app.RegisterModule()`

### Phase 6: Load Testing Infrastructure (independent; parallel with Phases 1-5)
- [ ] `tests/load/testapp/main.go` — minimal NebularCore app for load testing
- [ ] `tests/load/testapp/Dockerfile`
- [ ] `tests/load/testapp/config.yml`
- [ ] `tests/load/k6/websocket_load.js` — ramp profile, thresholds, `WS_HOST`/`WS_PORT` env vars
- [ ] `tests/load/docker-compose.yml`
- [ ] `tests/load/run.sh`
- [ ] `tests/load/load_test.go` — `//go:build loadtest`, `TestWebSocketLoad` via `os/exec`
- [ ] `Makefile` — `load-test` and `load-test-clean` targets
- [ ] `.gitlab-ci.yml` — `load_test` job (`when: manual`, artifact: k6 result JSON)

**Deliverable**: Automated load test proving the 100k concurrent connection target

---

## 8. Configuration Examples

### Example 1: Basic Setup (HTTP Upgrade, Same Port)

```yaml
modules:
  websocket:
    enabled: true
    server:
      port: "8080"
```

Result: WebSocket at `ws://localhost:8080/ws`

### Example 2: Multi-Tenant with Path-Based Tenants

```yaml
modules:
  websocket:
    enabled: true
    tenantMode: "path"
    security:
      authRequired: true
    events:
      allowedEventTypes:
        - "notification.*"
        - "data.*"
```

Result: `ws://localhost:8080/ws/tenant-123/` — isolated per tenant

### Example 3: Publishing from a Module (No WebSocket Code Needed)

```go
// In any module — WebSocket bridge picks this up automatically
func (m *MyModule) NotifyUser(ctx context.Context, data interface{}) error {
    return m.eventBus.Publish(ctx, event.Message{
        Source:    "mymodule",
        EventType: "mymodule.notification",
        Payload:   data,
    })
}
```

### Example 4: WebSocket-Aware Module

```go
func (m *RealtimeModule) Initialize(ctx context.Context, db *gorm.DB, router *gin.Engine) error {
    if m.wsAdapter != nil {
        m.wsAdapter.RegisterHandler("data:request", m.handleDataRequest)
    }
    return nil
}

func (m *RealtimeModule) handleDataRequest(ctx context.Context, conn websocket.Connection, msg *websocket.Message) error {
    result := m.processData(msg.Payload)
    return conn.Send(ctx, &websocket.Message{
        Type:    "data:response",
        Payload: result,
    })
}
```

### Example 5: Broadcasting from a Module

```go
func (m *CollabModule) onDocumentUpdate(ctx context.Context, docID string, update interface{}) error {
    if m.wsAdapter == nil {
        return nil
    }
    return m.wsAdapter.Broadcast(
        ctx,
        fmt.Sprintf("docs:%s", docID),
        &websocket.Message{Type: "doc:updated", Payload: update},
        func(conn websocket.Connection, topic string) bool {
            return m.canViewDoc(ctx, conn.UserID(), docID)
        },
    )
}
```

---

## 9. Key Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| WebSocket library | `github.com/coder/websocket` | Lower per-connection memory; context-native; actively maintained |
| Connection manager | 256-shard map | No global lock on read hot path; scales linearly with shard count |
| Write pattern | Dedicated goroutine + buffered channel per conn | Safe concurrent writes; non-blocking `Send()`; no goroutine leak |
| Backpressure | Drop + log on full write channel | Prevents slow clients from stalling fast ones or leaking goroutines |
| Load test tool | k6 (`grafana/k6`) | Native `k6/ws` WebSocket support; Docker-first; SLO thresholds built-in |
| Load test container | Docker Compose | No `testcontainers-go` dep in `go.mod`; portable; CI-friendly |
| Load test exclusion | `//go:build loadtest` | `go test ./...` safe without any extra configuration |
| Transport model | Adapter (not RPC) | Decouples modules from transport; aligns with event-driven architecture |
| Auth on upgrade | JWT via HTTP header/query | Standard; validates before WebSocket handshake completes; reuses `tools/security/jwt.go` |

---

## 10. Out of Scope

- Prometheus/metrics export (observability — deferred to a later phase)
- Separate WebSocket server port (upgrade-based same port only; extensible later)
- Distributed k6 (k6-operator/Kubernetes) — single-node load test assumed
- Persistent message queues (Watermill uses in-memory `gochannel` by default)
- gRPC or Server-Sent Events transports
- WebSocket compression (permessage-deflate) — can be enabled later via `coder/websocket` options
