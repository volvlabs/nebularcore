package localization

import (
	"context"
	"testing"

	"github.com/volvlabs/nebularcore/modules/localization/models"
	"github.com/volvlabs/nebularcore/modules/localization/repositories"
)

type fakeCountryRepo struct {
	countries map[string]*models.Country
	calls     int
}

func (f *fakeCountryRepo) FindByCode(_ context.Context, code string) (*models.Country, error) {
	f.calls++
	c, ok := f.countries[code]
	if !ok {
		return nil, repositories.ErrCountryNotFound
	}
	return c, nil
}

func (f *fakeCountryRepo) ListActive(_ context.Context) ([]models.Country, error) {
	var out []models.Country
	for _, c := range f.countries {
		if c.IsActive {
			out = append(out, *c)
		}
	}
	return out, nil
}

func (f *fakeCountryRepo) SetActive(_ context.Context, code string, active bool) error {
	c, ok := f.countries[code]
	if !ok {
		return repositories.ErrCountryNotFound
	}
	c.IsActive = active
	return nil
}

func newFakeRepo() *fakeCountryRepo {
	return &fakeCountryRepo{
		countries: map[string]*models.Country{
			"NG": {Code: "NG", CurrencyCode: "NGN", IsActive: true},
			"GH": {Code: "GH", CurrencyCode: "GHS", IsActive: false},
		},
	}
}

func TestCountryResolver_ResolveActiveAndInactive(t *testing.T) {
	repo := newFakeRepo()
	resolver := NewCountryResolver(repo, "NG")
	ctx := context.Background()

	ng, err := resolver.Resolve(ctx, "NG")
	if err != nil {
		t.Fatalf("resolving active country NG: %v", err)
	}
	if ng.CurrencyCode != "NGN" {
		t.Errorf("expected NGN, got %s", ng.CurrencyCode)
	}

	gh, err := resolver.Resolve(ctx, "GH")
	if err != nil {
		t.Fatalf("resolving inactive country GH should still succeed: %v", err)
	}
	if gh.IsActive {
		t.Errorf("GH should still be inactive")
	}
}

func TestCountryResolver_ResolveUnknown(t *testing.T) {
	repo := newFakeRepo()
	resolver := NewCountryResolver(repo, "NG")

	if _, err := resolver.Resolve(context.Background(), "ZZ"); err != repositories.ErrCountryNotFound {
		t.Errorf("expected ErrCountryNotFound, got %v", err)
	}
}

func TestCountryResolver_Default(t *testing.T) {
	repo := newFakeRepo()
	resolver := NewCountryResolver(repo, "NG")

	def, err := resolver.Default(context.Background())
	if err != nil {
		t.Fatalf("resolving default: %v", err)
	}
	if def.Code != "NG" {
		t.Errorf("expected default NG, got %s", def.Code)
	}
}

func TestCountryResolver_Caching(t *testing.T) {
	repo := newFakeRepo()
	resolver := NewCountryResolver(repo, "NG")
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		if _, err := resolver.Resolve(ctx, "NG"); err != nil {
			t.Fatalf("resolve %d: %v", i, err)
		}
	}

	if repo.calls != 1 {
		t.Errorf("expected repo to be hit once due to caching, got %d calls", repo.calls)
	}
}
