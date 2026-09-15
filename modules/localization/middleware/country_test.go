package middleware

import (
	"context"
	"errors"
	"testing"

	"github.com/volvlabs/nebularcore/modules/localization/models"
)

type fakeResolver struct {
	countries map[string]*models.Country
	def       *models.Country
	defErr    error
}

func (f *fakeResolver) Resolve(_ context.Context, code string) (*models.Country, error) {
	c, ok := f.countries[code]
	if !ok {
		return nil, errors.New("not found")
	}
	return c, nil
}

func (f *fakeResolver) Default(_ context.Context) (*models.Country, error) {
	return f.def, f.defErr
}

func newFakeResolver() *fakeResolver {
	global := &models.Country{Code: "ZZ", Name: "Global", CurrencyCode: "USD", IsActive: true}
	return &fakeResolver{
		countries: map[string]*models.Country{
			"NG": {Code: "NG", CurrencyCode: "NGN", IsActive: true},
			"GH": {Code: "GH", CurrencyCode: "GHS", IsActive: false},
			"ZZ": global,
		},
		def: global,
	}
}

func TestResolveAccount_ActiveCountry(t *testing.T) {
	resolver := newFakeResolver()
	accountResolver := func(_ context.Context, _ string) (string, bool) { return "NG", true }

	got := resolveAccount(context.Background(), resolver, accountResolver, "user-1")
	if got == nil || got.Code != "NG" {
		t.Fatalf("expected NG, got %+v", got)
	}
}

func TestResolveAccount_InactiveCountryFallsBackToDefault(t *testing.T) {
	resolver := newFakeResolver()
	accountResolver := func(_ context.Context, _ string) (string, bool) { return "GH", true }

	got := resolveAccount(context.Background(), resolver, accountResolver, "user-1")
	if got == nil || got.Code != "ZZ" {
		t.Fatalf("expected fallback to ZZ (Global) for an inactive country, got %+v", got)
	}
}

func TestResolveAccount_UnknownCodeFallsBackToDefault(t *testing.T) {
	resolver := newFakeResolver()
	accountResolver := func(_ context.Context, _ string) (string, bool) { return "XX", true }

	got := resolveAccount(context.Background(), resolver, accountResolver, "user-1")
	if got == nil || got.Code != "ZZ" {
		t.Fatalf("expected fallback to ZZ (Global) for an unresolvable code, got %+v", got)
	}
}

func TestResolveAccount_NoAccountCountryFallsBackToDefault(t *testing.T) {
	resolver := newFakeResolver()
	accountResolver := func(_ context.Context, _ string) (string, bool) { return "", false }

	got := resolveAccount(context.Background(), resolver, accountResolver, "user-1")
	if got == nil || got.Code != "ZZ" {
		t.Fatalf("expected fallback to ZZ (Global) when no account country is known, got %+v", got)
	}
}
