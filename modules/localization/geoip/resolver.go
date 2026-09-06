// Package geoip resolves a client IP address to a best-guess ISO country
// code, for DetectedCountry only (display/discovery convenience — never
// authoritative for billing; see the localization module's middleware).
package geoip

import "context"

// Resolver looks up the country an IP address most likely belongs to.
// ok=false means "no confident answer" (private/unroutable IP, provider
// disabled, lookup failure, or the DB has no entry) — callers should fall
// back to the module's configured default rather than treating it as an
// error.
type Resolver interface {
	Lookup(ctx context.Context, ip string) (countryCode string, ok bool)
}

// NoneResolver never resolves anything — used when GeoIP detection is
// disabled (Config.GeoIP.Provider == "none").
type NoneResolver struct{}

func (NoneResolver) Lookup(_ context.Context, _ string) (string, bool) {
	return "", false
}
