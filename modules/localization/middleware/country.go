// Package middleware resolves per-request country context and injects it
// into the gin context, mirroring modules/auth/middleware's c.Set("user", ...)
// pattern.
package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/volvlabs/nebularcore/modules/localization/geoip"
	"github.com/volvlabs/nebularcore/modules/localization/models"
)

// CountryResolver is a local, structural copy of the parent localization
// package's CountryResolver interface — defined here rather than imported
// to avoid an import cycle (the parent localization package imports this
// middleware package to wire it up in Module.Initialize). Any
// localization.CountryResolver implementation satisfies this automatically.
type CountryResolver interface {
	Resolve(ctx context.Context, code string) (*models.Country, error)
	Default(ctx context.Context) (*models.Country, error)
}

const (
	// AccountCountryKey is the gin context key for the authenticated
	// user's home country — authoritative for anything money-shaped.
	// Absent if no AccountCountryResolver was registered, or the current
	// request is unauthenticated.
	AccountCountryKey = "accountCountry"

	// DetectedCountryKey is the gin context key for the country inferred
	// from the request's client IP — display/discovery convenience only,
	// never authoritative for billing. Always set (falling back to the
	// module's configured default) unless GeoIP detection is disabled and
	// no default could be resolved either.
	DetectedCountryKey = "detectedCountry"
)

// AccountCountryResolverFunc looks up the ISO country code for an
// authenticated user's account/profile. Supplied by the host app — the
// localization module has no way to know a host app's own user/profile
// schema (see Module.WithAccountCountryResolver).
type AccountCountryResolverFunc func(ctx context.Context, userID string) (countryCode string, ok bool)

// ClientIPFunc extracts the trusted client IP (or, for a
// CloudflareHeaderResolver, the pre-extracted header value) from the
// request. Deliberately not hardcoded to gin's c.ClientIP() — that call is
// only trustworthy once the host app has correctly configured
// SetTrustedProxies/ForwardedByClientIP for its actual proxy/LB, and this
// middleware should not silently paper over a host app that hasn't done so.
type ClientIPFunc func(c *gin.Context) string

// AuthenticatedUserIDFunc extracts the current request's authenticated
// user ID, if any (e.g. from whatever the auth middleware set on the gin
// context). Supplied by the host app for the same reason as
// AccountCountryResolverFunc.
type AuthenticatedUserIDFunc func(c *gin.Context) (userID string, ok bool)

// New builds the country-resolution middleware. accountResolver and
// getUserID may be nil (AccountCountry resolution is then simply skipped
// for every request — DetectedCountry still works).
//
// AccountCountry is always set for an authenticated request once a
// resolver/getUserID pair is registered: it resolves to the user's real
// country when that country is known and active, and otherwise falls back
// to countryResolver.Default() (the module's configured "Global" sentinel
// by default — see config.Default). Callers therefore never need to
// handle a third "nothing resolved" case; the fallback is explicit rather
// than the key being silently absent.
func New(
	countryResolver CountryResolver,
	geoResolver geoip.Resolver,
	getClientIP ClientIPFunc,
	getUserID AuthenticatedUserIDFunc,
	accountResolver AccountCountryResolverFunc,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		if country := resolveDetected(ctx, c, countryResolver, geoResolver, getClientIP); country != nil {
			c.Set(DetectedCountryKey, country)
		}

		if accountResolver != nil && getUserID != nil {
			if userID, ok := getUserID(c); ok {
				if country := resolveAccount(ctx, countryResolver, accountResolver, userID); country != nil {
					c.Set(AccountCountryKey, country)
				}
			}
		}

		c.Next()
	}
}

// resolveAccount resolves userID's real, active home country, falling back
// to countryResolver.Default() whenever it has no resolvable country, the
// code doesn't match a known country, or the matched country isn't active
// (e.g. seeded but not yet launched) — never returning nil once a default
// exists, so AccountCountry is always present for an authenticated request.
func resolveAccount(
	ctx context.Context,
	countryResolver CountryResolver,
	accountResolver AccountCountryResolverFunc,
	userID string,
) *models.Country {
	if code, ok := accountResolver(ctx, userID); ok {
		country, err := countryResolver.Resolve(ctx, code)
		switch {
		case err != nil:
			log.Warn().Err(err).Str("userID", userID).Str("code", code).
				Msg("localization: account country code did not resolve to a known country, falling back to default")
		case !country.IsActive:
			log.Info().Str("userID", userID).Str("code", code).
				Msg("localization: account country is not active, falling back to default")
		default:
			return country
		}
	}

	def, err := countryResolver.Default(ctx)
	if err != nil {
		log.Warn().Err(err).Str("userID", userID).Msg("localization: failed to resolve default country")
		return nil
	}
	return def
}

func resolveDetected(
	ctx context.Context,
	c *gin.Context,
	countryResolver CountryResolver,
	geoResolver geoip.Resolver,
	getClientIP ClientIPFunc,
) *models.Country {
	if geoResolver != nil && getClientIP != nil {
		if code, ok := geoResolver.Lookup(ctx, getClientIP(c)); ok {
			if country, err := countryResolver.Resolve(ctx, code); err == nil && country.IsActive {
				return country
			}
		}
	}

	if def, err := countryResolver.Default(ctx); err == nil {
		return def
	}
	return nil
}

// FromContext reads a resolved country back off the gin context. ok=false
// means that key was never set (see the doc comments on AccountCountryKey /
// DetectedCountryKey for when that happens).
func FromContext(c *gin.Context, key string) (*models.Country, bool) {
	val, exists := c.Get(key)
	if !exists {
		return nil, false
	}
	country, ok := val.(*models.Country)
	return country, ok
}
