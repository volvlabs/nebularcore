package state_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/jideobs/nebularcore/modules/auth/config"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces"
	"gitlab.com/jideobs/nebularcore/modules/auth/state"
)

type mockUser struct {
	mock.Mock
	interfaces.User
}

func (m *mockUser) GetID() string {
	return "test-user-id"
}

func (m *mockUser) GetUsername() string {
	return "test-user"
}

func (m *mockUser) GetEmail() string {
	return "test@example.com"
}

func (m *mockUser) GetPhoneNumber() string {
	return "+2348091607293"
}

func TestJWTTokenIssuer(t *testing.T) {
	t.Run("IssueToken - Success", func(t *testing.T) {
		// Setup
		cfg := config.JWTConfig{
			AccessTokenSecret:  "test-access-secret",
			RefreshTokenSecret: "test-refresh-secret",
			AccessTokenExpiry:  time.Hour,
			RefreshTokenExpiry: time.Hour * 24,
		}
		issuer := state.NewJWTTokenIssuer(cfg)

		// Test
		user := &mockUser{}
		response, err := issuer.IssueToken(user)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Greater(t, response.ExpiresIn, int64(0))
	})

	t.Run("ValidateToken - Success", func(t *testing.T) {
		// Setup
		cfg := config.JWTConfig{
			AccessTokenSecret:  "test-access-secret",
			RefreshTokenSecret: "test-refresh-secret",
			AccessTokenExpiry:  time.Hour,
			RefreshTokenExpiry: time.Hour * 24,
		}
		issuer := state.NewJWTTokenIssuer(cfg)

		// Create a token first
		user := &mockUser{}
		response, err := issuer.IssueToken(user)
		assert.NoError(t, err)

		// Test
		claims, err := issuer.ValidateToken(response.AccessToken)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, claims)
	})

	t.Run("ValidateToken - Invalid Token", func(t *testing.T) {
		// Setup
		cfg := config.JWTConfig{
			AccessTokenSecret:  "test-access-secret",
			RefreshTokenSecret: "test-refresh-secret",
			AccessTokenExpiry:  time.Hour,
			RefreshTokenExpiry: time.Hour * 24,
		}
		issuer := state.NewJWTTokenIssuer(cfg)

		// Test
		_, err := issuer.ValidateToken("invalid.token.string")

		// Assert
		assert.Error(t, err)
	})

	t.Run("RefreshToken - Success", func(t *testing.T) {
		// Setup
		cfg := config.JWTConfig{
			AccessTokenSecret:  "test-access-secret",
			RefreshTokenSecret: "test-refresh-secret",
			AccessTokenExpiry:  time.Hour,
			RefreshTokenExpiry: time.Hour * 24,
		}
		issuer := state.NewJWTTokenIssuer(cfg)

		// Create initial tokens
		user := &mockUser{}
		initialResponse, err := issuer.IssueToken(user)
		assert.NoError(t, err)

		// Test
		refreshedResponse, err := issuer.RefreshToken(initialResponse.RefreshToken)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, refreshedResponse.AccessToken)
		assert.NotEmpty(t, refreshedResponse.RefreshToken)
		assert.Greater(t, refreshedResponse.ExpiresIn, int64(0))

		// Validate new token claims
		newClaims, err := issuer.ValidateToken(refreshedResponse.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, "test-user-id", newClaims["sub"])
		assert.Equal(t, "test-user", newClaims["username"])
		assert.Equal(t, "test@example.com", newClaims["email"])
		assert.NotNil(t, newClaims["iat"])
	})

	t.Run("RefreshToken - Invalid Token", func(t *testing.T) {
		// Setup
		cfg := config.JWTConfig{
			AccessTokenSecret:  "test-access-secret",
			RefreshTokenSecret: "test-refresh-secret",
			AccessTokenExpiry:  time.Hour,
			RefreshTokenExpiry: time.Hour * 24,
		}
		issuer := state.NewJWTTokenIssuer(cfg)

		// Test
		_, err := issuer.RefreshToken("invalid.refresh.token")

		// Assert
		assert.Error(t, err)
	})
}
