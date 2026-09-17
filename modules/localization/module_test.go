package localization

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/volvlabs/nebularcore/modules/localization/middleware"
)

func TestModule_Dependencies(t *testing.T) {
	noopAccountResolver := func(_ context.Context, _ string) (string, bool) { return "", false }
	noopGetUserID := func(_ *gin.Context) (string, bool) { return "", false }

	t.Run("no dependencies when AccountCountry is never wired", func(t *testing.T) {
		m := New()
		assert.Equal(t, []string{}, m.Dependencies())
	})

	t.Run("no dependencies when only the account resolver is wired", func(t *testing.T) {
		m := New().WithAccountCountryResolver(noopAccountResolver)
		assert.Equal(t, []string{}, m.Dependencies())
	})

	t.Run("no dependencies when only getUserID is wired", func(t *testing.T) {
		m := New().WithAuthenticatedUserIDFunc(noopGetUserID)
		assert.Equal(t, []string{}, m.Dependencies())
	})

	t.Run("depends on auth once both are wired", func(t *testing.T) {
		m := New().WithAuthenticatedUserIDFunc(noopGetUserID)
		m.WithAccountCountryResolver(noopAccountResolver)
		assert.Equal(t, []string{"auth"}, m.Dependencies())
	})

	var _ middleware.AccountCountryResolverFunc = noopAccountResolver
	var _ middleware.AuthenticatedUserIDFunc = noopGetUserID
}
