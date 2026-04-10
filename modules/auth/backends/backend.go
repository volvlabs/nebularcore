package backends

import (
	"context"

	"github.com/volvlabs/nebularcore/modules/auth/interfaces"
)

// AuthenticationBackend defines the interface for authentication backends
type AuthenticationBackend interface {
	// Authenticate attempts to authenticate a user with the given credentials
	Authenticate(ctx context.Context, credentials map[string]any) (interfaces.User, error)
	// ValidateToken validates a token and returns the associated user
	ValidateToken(ctx context.Context, token string) (interfaces.User, error)
}
