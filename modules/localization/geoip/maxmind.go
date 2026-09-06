package geoip

import (
	"context"
	"fmt"
	"net"

	geoip2 "github.com/oschwald/geoip2-golang"
)

// MaxMindResolver looks up countries from a self-hosted GeoLite2-Country (or
// commercial GeoIP2-Country) .mmdb file — no per-request external API call.
type MaxMindResolver struct {
	db *geoip2.Reader
}

// NewMaxMindResolver opens the .mmdb file at path. The returned resolver
// owns the file handle; call Close when the host app shuts down.
func NewMaxMindResolver(path string) (*MaxMindResolver, error) {
	db, err := geoip2.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening maxmind db at %q: %w", path, err)
	}
	return &MaxMindResolver{db: db}, nil
}

func (r *MaxMindResolver) Close() error {
	return r.db.Close()
}

func (r *MaxMindResolver) Lookup(_ context.Context, ip string) (string, bool) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "", false
	}

	record, err := r.db.Country(parsed)
	if err != nil || record == nil {
		return "", false
	}

	code := record.Country.IsoCode
	if code == "" {
		return "", false
	}
	return code, true
}
