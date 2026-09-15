// Package localization provides generic, app-agnostic country/currency/
// timezone/phone reference data and resolution — a Country registry, an
// AccountCountry vs DetectedCountry resolution middleware, and IP-based
// country detection — for any nebularcore-based product operating in more
// than one country. It intentionally carries no payment-provider or other
// app-specific business logic; host apps layer that on top, keyed by the
// ISO country code this module resolves.
package localization

import (
	"context"
	"embed"
	"fmt"

	"github.com/gin-gonic/gin"
	coreConfig "github.com/volvlabs/nebularcore/core/config"
	migrationRunner "github.com/volvlabs/nebularcore/core/migration_runner"
	"github.com/volvlabs/nebularcore/core/module"
	"github.com/volvlabs/nebularcore/modules/localization/config"
	"github.com/volvlabs/nebularcore/modules/localization/geoip"
	"github.com/volvlabs/nebularcore/modules/localization/middleware"
	"github.com/volvlabs/nebularcore/modules/localization/repositories"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Module implements the nebularcore module interface for country/locale
// resolution.
type Module struct {
	name    string
	version string
	config  *config.Config

	repo        repositories.CountryRepository
	resolver    CountryResolver
	geoResolver geoip.Resolver

	getClientIP     middleware.ClientIPFunc
	getUserID       middleware.AuthenticatedUserIDFunc
	accountResolver middleware.AccountCountryResolverFunc

	maxmindCloser interface{ Close() error }
}

// New creates a new localization module.
func New() *Module {
	return &Module{
		name:    "localization",
		version: "1.0.0",
		config:  config.Default(),
	}
}

func (m *Module) Name() string    { return m.name }
func (m *Module) Version() string { return m.version }

// Dependencies returns the module's dependencies. Localization has none —
// it's foundational reference data other modules (auth, billing, etc.)
// depend on, not the reverse.
func (m *Module) Dependencies() []string { return []string{} }

func (m *Module) MigrationsDir() string    { return "migrations" }
func (m *Module) ProvidesMigrations() bool { return true }

func (m *Module) GetMigrationSources(_ string) []migrationRunner.Source {
	return []migrationRunner.Source{
		{
			FS:       migrations,
			Path:     "migrations",
			Priority: 10,
			Exclude:  []string{},
		},
	}
}

func (m *Module) Namespace() module.ModuleNamespace {
	return module.PublicNamespace
}

func (m *Module) NewConfig() coreConfig.Config {
	return config.Default()
}

func (m *Module) Configure(cfg coreConfig.Config) error {
	c, ok := cfg.(*config.Config)
	if !ok {
		return fmt.Errorf("invalid config type for localization module")
	}
	if err := c.Validate(); err != nil {
		return err
	}
	m.config = c
	return nil
}

func (m *Module) Initialize(_ context.Context, db *gorm.DB, router *gin.Engine) error {
	if m.repo == nil {
		m.repo = repositories.NewCountryRepository(db)
	}
	if m.resolver == nil {
		m.resolver = NewCountryResolver(m.repo, m.config.DefaultCountryCode)
	}

	if m.geoResolver == nil {
		resolver, closer, err := buildGeoResolver(m.config.GeoIP)
		if err != nil {
			return fmt.Errorf("initializing localization geoip resolver: %w", err)
		}
		m.geoResolver = resolver
		m.maxmindCloser = closer
	}

	router.Use(middleware.New(m.resolver, m.geoResolver, m.getClientIP, m.getUserID, m.accountResolver))

	return nil
}

func buildGeoResolver(cfg config.GeoIPConfig) (geoip.Resolver, interface{ Close() error }, error) {
	switch cfg.Provider {
	case config.GeoIPProviderMaxMind:
		r, err := geoip.NewMaxMindResolver(cfg.MaxMindDBPath)
		if err != nil {
			return nil, nil, err
		}
		return r, r, nil
	case config.GeoIPProviderCloudflareHeader:
		return geoip.NewCloudflareHeaderResolver(cfg.HeaderName), nil, nil
	default:
		return geoip.NoneResolver{}, nil, nil
	}
}

func (m *Module) Shutdown(_ context.Context) error {
	if m.maxmindCloser != nil {
		return m.maxmindCloser.Close()
	}
	return nil
}

// Resolver exposes the module's CountryResolver for other modules/host-app
// code to use directly (e.g. billing sourcing a currency code, or
// account-country resolvers wanting to validate a code exists).
func (m *Module) Resolver() CountryResolver {
	return m.resolver
}

// Repository exposes the module's CountryRepository, e.g. for an admin
// handler toggling IsActive.
func (m *Module) Repository() repositories.CountryRepository {
	return m.repo
}

// WithClientIPFunc sets how the middleware extracts a trusted client IP (or
// pre-extracted header value, for a CloudflareHeaderResolver) from each
// request. Must be set before Initialize for GeoIP-based DetectedCountry
// resolution to do anything — deliberately not defaulted to gin's
// c.ClientIP(), since that's only trustworthy once the host app has
// correctly configured SetTrustedProxies/ForwardedByClientIP for its real
// proxy/LB.
func (m *Module) WithClientIPFunc(fn middleware.ClientIPFunc) *Module {
	m.getClientIP = fn
	return m
}

// WithAuthenticatedUserIDFunc sets how the middleware reads the current
// request's authenticated user ID (e.g. from whatever the host app's auth
// middleware set on the gin context).
func (m *Module) WithAuthenticatedUserIDFunc(fn middleware.AuthenticatedUserIDFunc) *Module {
	m.getUserID = fn
	return m
}

// WithAccountCountryResolver registers the host app's lookup from user ID
// to home-country code (e.g. reading Address.Country off the user's
// profile). Without this, AccountCountry is never resolved and only
// DetectedCountry (IP-based) is available.
func (m *Module) WithAccountCountryResolver(fn middleware.AccountCountryResolverFunc) *Module {
	m.accountResolver = fn
	return m
}

// WithCountryRepository overrides the repository (e.g. to inject a fake in
// tests).
func (m *Module) WithCountryRepository(repo repositories.CountryRepository) *Module {
	m.repo = repo
	return m
}

// WithCountryResolver overrides the resolver (e.g. to inject a fake in
// tests, or a differently-tuned cache).
func (m *Module) WithCountryResolver(resolver CountryResolver) *Module {
	m.resolver = resolver
	return m
}

// WithGeoResolver overrides IP-resolution entirely (e.g. to inject a fake
// in tests, or a provider not covered by config.GeoIPProvider).
func (m *Module) WithGeoResolver(resolver geoip.Resolver) *Module {
	m.geoResolver = resolver
	return m
}
