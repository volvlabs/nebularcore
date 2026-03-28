package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/jideobs/nebularcore/tools/security"
)

const testSecret = "test-secret-key-12345"

func TestValidateToken_Valid(t *testing.T) {
	token, err := security.NewJWT(jwt.MapClaims{
		"user_id":   "user-123",
		"tenant_id": "tenant-456",
	}, testSecret, 3600)
	require.NoError(t, err)

	claims, err := ValidateToken(token, testSecret)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "tenant-456", claims.TenantID)
}

func TestValidateToken_SubFallback(t *testing.T) {
	token, err := security.NewJWT(jwt.MapClaims{
		"sub":       "user-789",
		"tenant_id": "tenant-1",
	}, testSecret, 3600)
	require.NoError(t, err)

	claims, err := ValidateToken(token, testSecret)
	require.NoError(t, err)
	assert.Equal(t, "user-789", claims.UserID)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	_, err := ValidateToken("not-a-valid-token", testSecret)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestValidateToken_WrongSecret(t *testing.T) {
	token, err := security.NewJWT(jwt.MapClaims{
		"user_id": "user-1",
	}, testSecret, 3600)
	require.NoError(t, err)

	_, err = ValidateToken(token, "wrong-secret")
	require.Error(t, err)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	// Create a token that expired 1 hour ago.
	claims := jwt.MapClaims{
		"user_id": "user-1",
		"exp":     time.Now().Add(-time.Hour).Unix(),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(testSecret))
	require.NoError(t, err)

	_, err = ValidateToken(token, testSecret)
	require.Error(t, err)
}

func TestValidateOrigin(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		allowed []string
		want    bool
	}{
		{"empty allowlist permits all", "http://any.com", nil, true},
		{"allowed origin", "http://localhost:3000", []string{"http://localhost:3000"}, true},
		{"blocked origin", "http://evil.com", []string{"http://localhost:3000"}, false},
		{"multiple allowed", "http://b.com", []string{"http://a.com", "http://b.com"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ValidateOrigin(tt.origin, tt.allowed))
		})
	}
}

func TestExtractTokenFromQuery(t *testing.T) {
	assert.Equal(t, "abc123", ExtractTokenFromQuery("ws://localhost/ws?token=abc123"))
	assert.Equal(t, "", ExtractTokenFromQuery("ws://localhost/ws"))
	assert.Equal(t, "", ExtractTokenFromQuery("://bad"))
}
