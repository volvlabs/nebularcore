package auth

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestBaseProvider(t *testing.T) {
	// Arrange:
	b := baseProvider{}

	// Act-Assert:
	b.SetDisplayName("Google")
	assert.Equal(t, "Google", b.DisplayName())

	b.SetClientId("test")
	assert.Equal(t, "test", b.ClientId())

	b.SetClientSecret("test-secret")
	assert.Equal(t, "test-secret", b.ClientSecret())

	b.SetRedirectUrl("redirect-url")
	assert.Equal(t, "redirect-url", b.RedirectUrl())

	b.SetScopes([]string{"test"})
	assert.Equal(t, []string{"test"}, b.Scopes())

	b.SetAuthUrl("auth-url")
	assert.Equal(t, "auth-url", b.AuthUrl())

	b.SetTokenUrl("token-url")
	assert.Equal(t, "token-url", b.TokenUrl())

	b.SetUserApiUrl("api-url")
	assert.Equal(t, "api-url", b.UserApiUrl())
}
