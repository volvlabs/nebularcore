package localization

import (
	"context"
	"sync"
	"time"

	"github.com/volvlabs/nebularcore/modules/localization/models"
	"github.com/volvlabs/nebularcore/modules/localization/repositories"
)

// CountryResolver resolves ISO country codes to full Country reference
// records. Implementations may cache, since the countries table is
// admin-editable reference data read on every request via the module's
// middleware, not something that needs a fresh query each time.
type CountryResolver interface {
	Resolve(ctx context.Context, code string) (*models.Country, error)
	Default(ctx context.Context) (*models.Country, error)
}

const defaultCacheTTL = 5 * time.Minute

// cachedResolver is a short-TTL, in-process cache in front of
// CountryRepository. It caches misses as well as hits (a zero-value entry
// with found=false) so a resolver.Resolve("XX") for a code that will never
// exist doesn't hit the DB on every request.
type cachedResolver struct {
	repo        repositories.CountryRepository
	defaultCode string
	ttl         time.Duration
	mu          sync.RWMutex
	cache       map[string]cacheEntry
}

type cacheEntry struct {
	country   *models.Country
	found     bool
	expiresAt time.Time
}

// NewCountryResolver builds a cached CountryResolver. defaultCode is the
// country returned by Default (and used as a resolution fallback by
// consumers) — typically the host app's primary/home country.
func NewCountryResolver(repo repositories.CountryRepository, defaultCode string) CountryResolver {
	return &cachedResolver{
		repo:        repo,
		defaultCode: defaultCode,
		ttl:         defaultCacheTTL,
		cache:       make(map[string]cacheEntry),
	}
}

func (r *cachedResolver) Resolve(ctx context.Context, code string) (*models.Country, error) {
	r.mu.RLock()
	entry, ok := r.cache[code]
	r.mu.RUnlock()
	if ok && time.Now().Before(entry.expiresAt) {
		if !entry.found {
			return nil, repositories.ErrCountryNotFound
		}
		return entry.country, nil
	}

	country, err := r.repo.FindByCode(ctx, code)
	r.mu.Lock()
	defer r.mu.Unlock()
	if err != nil {
		r.cache[code] = cacheEntry{found: false, expiresAt: time.Now().Add(r.ttl)}
		return nil, err
	}
	r.cache[code] = cacheEntry{country: country, found: true, expiresAt: time.Now().Add(r.ttl)}
	return country, nil
}

func (r *cachedResolver) Default(ctx context.Context) (*models.Country, error) {
	return r.Resolve(ctx, r.defaultCode)
}
