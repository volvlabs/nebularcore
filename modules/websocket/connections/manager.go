package connections

import (
	"hash/fnv"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog/log"
)

const numShards = 256

// shard holds a subset of connections.
type shard struct {
	mu    sync.RWMutex
	conns map[string]*conn
}

// PresenceListener is notified when a user transitions between having zero
// and having at least one live connection. It is invoked exactly on the 0->1
// ("came online") and 1->0 ("went offline") edges, never on every individual
// connect/disconnect — a user with several simultaneous connections (e.g.
// multiple browser tabs plus mobile) only fires this once per edge.
type PresenceListener interface {
	OnPresenceChange(userID string, online bool)
}

// Manager tracks all active WebSocket connections using a sharded map for
// high-concurrency access.
type Manager struct {
	shards   [numShards]shard
	total    atomic.Int64
	maxConns int64

	// userIndex maps userID -> set of connIDs.
	userIndex sync.Map
	// tenantIndex maps tenantID -> set of connIDs.
	tenantIndex sync.Map

	// userConnCounts maps userID -> *atomic.Int64, tracking live connection
	// count per user so presence transitions can be detected race-free
	// (ranging userIndex's sync.Map is not a reliable way to detect the
	// exact 0<->1 edge under concurrent Register/Deregister).
	userConnCounts sync.Map

	presence PresenceListener
}

// SetPresenceListener registers a listener notified on user presence
// transitions (0->1 and 1->0 live-connection edges). Not safe to call
// concurrently with Register/Deregister — set once during setup.
func (m *Manager) SetPresenceListener(l PresenceListener) {
	m.presence = l
}

// NewManager creates a connection manager with the given max connections limit.
func NewManager(maxConns int64) *Manager {
	m := &Manager{maxConns: maxConns}
	for i := range m.shards {
		m.shards[i].conns = make(map[string]*conn)
	}
	return m
}

// shardFor returns the shard index for a connection ID.
func shardFor(id string) int {
	h := fnv.New32a()
	h.Write([]byte(id))
	return int(h.Sum32() % numShards)
}

// Register adds a connection. Returns false if the max connections limit
// has been reached.
func (m *Manager) Register(c *conn) bool {
	if m.total.Load() >= m.maxConns {
		log.Warn().Int64("max", m.maxConns).Msg("max connections reached, rejecting")
		return false
	}

	idx := shardFor(c.id)
	s := &m.shards[idx]
	s.mu.Lock()
	s.conns[c.id] = c
	s.mu.Unlock()

	m.total.Add(1)

	// Update user index.
	if c.userID != "" {
		actual, _ := m.userIndex.LoadOrStore(c.userID, &sync.Map{})
		actual.(*sync.Map).Store(c.id, struct{}{})

		if newCount := m.incrUserConnCount(c.userID); newCount == 1 && m.presence != nil {
			m.presence.OnPresenceChange(c.userID, true)
		}
	}

	// Update tenant index.
	if c.tenantID != "" {
		actual, _ := m.tenantIndex.LoadOrStore(c.tenantID, &sync.Map{})
		actual.(*sync.Map).Store(c.id, struct{}{})
	}

	return true
}

// Deregister removes a connection.
func (m *Manager) Deregister(id string) {
	idx := shardFor(id)
	s := &m.shards[idx]

	s.mu.Lock()
	c, ok := s.conns[id]
	if ok {
		delete(s.conns, id)
	}
	s.mu.Unlock()

	if !ok {
		return
	}

	m.total.Add(-1)

	// Clean up user index.
	if c.userID != "" {
		if v, ok := m.userIndex.Load(c.userID); ok {
			v.(*sync.Map).Delete(id)
		}

		if newCount := m.decrUserConnCount(c.userID); newCount == 0 && m.presence != nil {
			m.presence.OnPresenceChange(c.userID, false)
		}
	}

	// Clean up tenant index.
	if c.tenantID != "" {
		if v, ok := m.tenantIndex.Load(c.tenantID); ok {
			v.(*sync.Map).Delete(id)
		}
	}
}

// Get returns a connection by ID, or nil.
func (m *Manager) Get(id string) Connection {
	idx := shardFor(id)
	s := &m.shards[idx]
	s.mu.RLock()
	c := s.conns[id]
	s.mu.RUnlock()
	if c == nil {
		return nil
	}
	return c
}

// Total returns the total number of active connections.
func (m *Manager) Total() int64 {
	return m.total.Load()
}

// GetByUser returns all connections for a given user ID.
func (m *Manager) GetByUser(userID string) []Connection {
	v, ok := m.userIndex.Load(userID)
	if !ok {
		return nil
	}
	var result []Connection
	v.(*sync.Map).Range(func(key, _ any) bool {
		if c := m.Get(key.(string)); c != nil {
			result = append(result, c)
		}
		return true
	})
	return result
}

// GetByTenant returns all connections for a given tenant ID.
func (m *Manager) GetByTenant(tenantID string) []Connection {
	v, ok := m.tenantIndex.Load(tenantID)
	if !ok {
		return nil
	}
	var result []Connection
	v.(*sync.Map).Range(func(key, _ any) bool {
		if c := m.Get(key.(string)); c != nil {
			result = append(result, c)
		}
		return true
	})
	return result
}

// GetAll returns all active connections. Use sparingly — iterates all shards.
func (m *Manager) GetAll() []Connection {
	result := make([]Connection, 0, m.total.Load())
	for i := range m.shards {
		s := &m.shards[i]
		s.mu.RLock()
		for _, c := range s.conns {
			result = append(result, c)
		}
		s.mu.RUnlock()
	}
	return result
}

// incrUserConnCount increments the live connection counter for userID and
// returns the new count.
func (m *Manager) incrUserConnCount(userID string) int64 {
	actual, _ := m.userConnCounts.LoadOrStore(userID, &atomic.Int64{})
	return actual.(*atomic.Int64).Add(1)
}

// decrUserConnCount decrements the live connection counter for userID and
// returns the new count.
func (m *Manager) decrUserConnCount(userID string) int64 {
	actual, _ := m.userConnCounts.LoadOrStore(userID, &atomic.Int64{})
	return actual.(*atomic.Int64).Add(-1)
}

// UserConnectionCount returns the number of connections for a given user.
func (m *Manager) UserConnectionCount(userID string) int {
	v, ok := m.userIndex.Load(userID)
	if !ok {
		return 0
	}
	count := 0
	v.(*sync.Map).Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

// TenantConnectionCount returns the number of connections for a given tenant.
func (m *Manager) TenantConnectionCount(tenantID string) int {
	v, ok := m.tenantIndex.Load(tenantID)
	if !ok {
		return 0
	}
	count := 0
	v.(*sync.Map).Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}
