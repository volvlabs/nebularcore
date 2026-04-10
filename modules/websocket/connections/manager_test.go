package connections

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volvlabs/nebularcore/modules/websocket/protocol"
)

func makeConn(id, userID, tenantID string) *conn {
	return NewConnection(id, userID, tenantID, context.Background())
}

func TestManager_RegisterAndGet(t *testing.T) {
	m := NewManager(100)
	c := makeConn("c1", "u1", "t1")

	ok := m.Register(c)
	require.True(t, ok)
	assert.Equal(t, int64(1), m.Total())

	got := m.Get("c1")
	require.NotNil(t, got)
	assert.Equal(t, "c1", got.ID())
	assert.Equal(t, "u1", got.UserID())
	assert.Equal(t, "t1", got.TenantID())
}

func TestManager_Deregister(t *testing.T) {
	m := NewManager(100)
	c := makeConn("c1", "u1", "t1")
	m.Register(c)

	m.Deregister("c1")
	assert.Equal(t, int64(0), m.Total())
	assert.Nil(t, m.Get("c1"))
}

func TestManager_DeregisterNonExistent(t *testing.T) {
	m := NewManager(100)
	m.Deregister("nope") // should not panic
	assert.Equal(t, int64(0), m.Total())
}

func TestManager_MaxConnectionsLimit(t *testing.T) {
	m := NewManager(2)
	m.Register(makeConn("c1", "", ""))
	m.Register(makeConn("c2", "", ""))

	ok := m.Register(makeConn("c3", "", ""))
	assert.False(t, ok)
	assert.Equal(t, int64(2), m.Total())
}

func TestManager_GetByUser(t *testing.T) {
	m := NewManager(100)
	m.Register(makeConn("c1", "u1", "t1"))
	m.Register(makeConn("c2", "u1", "t1"))
	m.Register(makeConn("c3", "u2", "t1"))

	conns := m.GetByUser("u1")
	assert.Len(t, conns, 2)

	conns = m.GetByUser("u2")
	assert.Len(t, conns, 1)

	conns = m.GetByUser("nope")
	assert.Len(t, conns, 0)
}

func TestManager_GetByTenant(t *testing.T) {
	m := NewManager(100)
	m.Register(makeConn("c1", "u1", "t1"))
	m.Register(makeConn("c2", "u2", "t1"))
	m.Register(makeConn("c3", "u3", "t2"))

	conns := m.GetByTenant("t1")
	assert.Len(t, conns, 2)

	conns = m.GetByTenant("t2")
	assert.Len(t, conns, 1)
}

func TestManager_GetAll(t *testing.T) {
	m := NewManager(100)
	for i := 0; i < 10; i++ {
		m.Register(makeConn(fmt.Sprintf("c%d", i), "u", "t"))
	}
	assert.Len(t, m.GetAll(), 10)
}

func TestManager_UserIndexCleanup(t *testing.T) {
	m := NewManager(100)
	m.Register(makeConn("c1", "u1", "t1"))
	assert.Equal(t, 1, m.UserConnectionCount("u1"))

	m.Deregister("c1")
	assert.Equal(t, 0, m.UserConnectionCount("u1"))
}

func TestManager_TenantIndexCleanup(t *testing.T) {
	m := NewManager(100)
	m.Register(makeConn("c1", "u1", "t1"))
	assert.Equal(t, 1, m.TenantConnectionCount("t1"))

	m.Deregister("c1")
	assert.Equal(t, 0, m.TenantConnectionCount("t1"))
}

// TestManager_ConcurrentRegisterDeregister runs 10k goroutines registering and
// deregistering connections concurrently. Must pass with -race.
func TestManager_ConcurrentRegisterDeregister(t *testing.T) {
	m := NewManager(100000)
	const n = 10000

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("c%d", i)
			c := makeConn(id, fmt.Sprintf("u%d", i%100), fmt.Sprintf("t%d", i%10))
			m.Register(c)
		}(i)
	}
	wg.Wait()
	assert.Equal(t, int64(n), m.Total())

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			m.Deregister(fmt.Sprintf("c%d", i))
		}(i)
	}
	wg.Wait()
	assert.Equal(t, int64(0), m.Total())
}

// TestManager_ConcurrentMixed tests interleaved register/deregister operations.
func TestManager_ConcurrentMixed(t *testing.T) {
	m := NewManager(100000)
	const n = 5000

	var wg sync.WaitGroup
	wg.Add(n * 2)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("cm%d", i)
			m.Register(makeConn(id, "u", "t"))
		}(i)
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("cm%d", i)
			m.Deregister(id)
		}(i)
	}
	wg.Wait()
	// Total should be >= 0 and <= n
	total := m.Total()
	assert.GreaterOrEqual(t, total, int64(0))
	assert.LessOrEqual(t, total, int64(n))
}

func TestManager_ShardDistribution(t *testing.T) {
	// Verify that shardFor distributes across multiple shards.
	shardCounts := make(map[int]int, numShards)
	for i := 0; i < 10000; i++ {
		idx := shardFor(fmt.Sprintf("conn-%d", i))
		shardCounts[idx]++
	}
	// Should use multiple shards, not just one.
	assert.Greater(t, len(shardCounts), 100, "expected connections to distribute across >100 shards")
}

func TestConnection_Send(t *testing.T) {
	c := makeConn("c1", "u1", "t1")
	ok := c.Send(&protocol.ServerMessage{Type: protocol.TypePong})
	assert.True(t, ok)
}

func TestConnection_SendDropsOnFull(t *testing.T) {
	c := makeConn("c1", "u1", "t1")
	// Fill the channel.
	for i := 0; i < 256; i++ {
		c.Send(&protocol.ServerMessage{Type: protocol.TypePong})
	}
	// Next send should drop.
	ok := c.Send(&protocol.ServerMessage{Type: protocol.TypePong})
	assert.False(t, ok)
}

func TestConnection_CloseIdempotent(t *testing.T) {
	c := makeConn("c1", "u1", "t1")
	c.Close()
	c.Close() // must not panic
	assert.Error(t, c.Context().Err())
}

func TestPool_CheckLimits(t *testing.T) {
	m := NewManager(100)
	p := NewPool(m, 2, 5)

	m.Register(makeConn("c1", "u1", "t1"))
	m.Register(makeConn("c2", "u1", "t1"))

	err := p.CheckLimits("u1", "t1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user u1")

	err = p.CheckLimits("u2", "t1")
	assert.NoError(t, err)
}

func TestPool_Buffer(t *testing.T) {
	m := NewManager(100)
	p := NewPool(m, 10, 100)

	buf := p.GetBuffer()
	buf.WriteString("test")
	p.PutBuffer(buf)

	buf2 := p.GetBuffer()
	assert.Equal(t, 0, buf2.Len()) // should be reset
}
