package connections

import (
	"bytes"
	"fmt"
	"sync"
)

// Pool provides limit enforcement for per-user and per-tenant connections,
// and a bytes.Buffer pool for reducing GC pressure during encoding.
type Pool struct {
	manager                 *Manager
	maxConnectionsPerUser   int
	maxConnectionsPerTenant int
	bufPool                 sync.Pool
}

// NewPool creates a new Pool.
func NewPool(manager *Manager, maxPerUser, maxPerTenant int) *Pool {
	return &Pool{
		manager:                 manager,
		maxConnectionsPerUser:   maxPerUser,
		maxConnectionsPerTenant: maxPerTenant,
		bufPool: sync.Pool{
			New: func() any { return new(bytes.Buffer) },
		},
	}
}

// CheckLimits returns an error if the user or tenant has reached their
// connection limit.
func (p *Pool) CheckLimits(userID, tenantID string) error {
	if userID != "" && p.maxConnectionsPerUser > 0 {
		if p.manager.UserConnectionCount(userID) >= p.maxConnectionsPerUser {
			return fmt.Errorf("user %s has reached max connections (%d)", userID, p.maxConnectionsPerUser)
		}
	}
	if tenantID != "" && p.maxConnectionsPerTenant > 0 {
		if p.manager.TenantConnectionCount(tenantID) >= p.maxConnectionsPerTenant {
			return fmt.Errorf("tenant %s has reached max connections (%d)", tenantID, p.maxConnectionsPerTenant)
		}
	}
	return nil
}

// GetBuffer returns a bytes.Buffer from the pool.
func (p *Pool) GetBuffer() *bytes.Buffer {
	buf := p.bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutBuffer returns a bytes.Buffer to the pool.
func (p *Pool) PutBuffer(buf *bytes.Buffer) {
	p.bufPool.Put(buf)
}
