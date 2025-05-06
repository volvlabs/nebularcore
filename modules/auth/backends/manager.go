package backends

import (
	"context"
	"errors"
	"fmt"
	"sync"

	autherrors "gitlab.com/jideobs/nebularcore/modules/auth/errors"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces"
)

// AuthenticationManager manages multiple authentication backends
type AuthenticationManager interface {
	RegisterBackend(name string, backend AuthenticationBackend)
	Authenticate(ctx context.Context, credentials map[string]any) (interfaces.User, error)
	ValidateToken(ctx context.Context, token string) (interfaces.User, error)
}

// authManager implements the AuthenticationManager interface
type authManager struct {
	mu       sync.RWMutex
	backends map[string]AuthenticationBackend
}

// NewAuthenticationManager creates a new authentication manager
func NewAuthenticationManager() AuthenticationManager {
	return &authManager{
		backends: make(map[string]AuthenticationBackend),
	}
}

// RegisterBackend registers a new authentication backend
func (m *authManager) RegisterBackend(name string, backend AuthenticationBackend) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.backends[name]; exists {
		panic("backend already registered: " + name)
	}

	m.backends[name] = backend
}

// Authenticate authenticates a user using all registered backends
func (m *authManager) Authenticate(ctx context.Context, credentials map[string]any) (interfaces.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.backends) == 0 {
		return nil, fmt.Errorf("no authentication backends registered")
	}

	var lastErr error
	for _, backend := range m.backends {
		user, err := backend.Authenticate(ctx, credentials)
		if err == nil {
			return user, nil
		}
		if !errors.Is(err, autherrors.ErrInvalidCredentials) {
			return nil, err
		}
		lastErr = err
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("no authentication backend could handle the credentials")
}

// ValidateToken validates a token using all registered backends
func (m *authManager) ValidateToken(ctx context.Context, token string) (interfaces.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.backends) == 0 {
		return nil, fmt.Errorf("no authentication backends registered")
	}

	var lastErr error
	for _, backend := range m.backends {
		user, err := backend.ValidateToken(ctx, token)
		if err == nil {
			return user, nil
		}
		if !errors.Is(err, autherrors.ErrInvalidToken) {
			return nil, err
		}
		lastErr = err
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("no authentication backend could validate the token")
}

// IsAuthenticationError checks if the error is an authentication error
func IsAuthenticationError(err error) bool {
	// TODO: Implement proper error type checking
	return true
}
