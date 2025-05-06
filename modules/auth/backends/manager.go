package backends

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
	"gitlab.com/jideobs/nebularcore/modules/auth/emitter"
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
	mu           sync.RWMutex
	backends     map[string]AuthenticationBackend
	backendOrder []string
	eventEmitter emitter.EventEmitter
}

// NewAuthenticationManager creates a new authentication manager
func NewAuthenticationManager(eventEmitter emitter.EventEmitter) AuthenticationManager {
	return &authManager{
		backends:     make(map[string]AuthenticationBackend),
		eventEmitter: eventEmitter,
	}
}

// RegisterBackend registers a new authentication backend
func (m *authManager) RegisterBackend(name string, backend AuthenticationBackend) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.backends[name]; exists {
		panic("backend already registered: " + name)
	}

	log.Info().Msgf("registering authentication backend: %s", name)
	m.backends[name] = backend
	m.backendOrder = append(m.backendOrder, name)
}

// Authenticate authenticates a user using all registered backends
func (m *authManager) Authenticate(ctx context.Context, credentials map[string]any) (interfaces.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.backends) == 0 {
		return nil, fmt.Errorf("no authentication backends registered")
	}

	var (
		user interfaces.User
		err  error
	)

	for _, name := range m.backendOrder {
		backend := m.backends[name]
		user, err = backend.Authenticate(ctx, credentials)
		if err == nil && user != nil {
			log.Info().Str("userID", user.GetID().String()).Msg("authentication successful using backend: " + name)
			m.eventEmitter.EmitLoginEvent(ctx, user, "", "", true)
			return user, nil
		}
	}

	if err != nil {
		if errors.Is(err, autherrors.ErrInvalidCredentials) {
			m.eventEmitter.EmitLoginEvent(ctx, user, "", "", false)
		}
		return nil, err
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
	for _, name := range m.backendOrder {
		backend := m.backends[name]
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
