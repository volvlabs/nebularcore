package config

import "fmt"

// GeoIPProvider selects how DetectedCountry (Section A2 — display/discovery
// only, never authoritative for billing) is resolved from a client IP.
type GeoIPProvider string

const (
	// GeoIPProviderNone disables IP-based country detection entirely;
	// only AccountCountry resolution (if the host app registered a
	// resolver) is available.
	GeoIPProviderNone GeoIPProvider = "none"
	// GeoIPProviderMaxMind reads a self-hosted MaxMind GeoLite2-Country
	// (or commercial GeoIP2-Country) database file.
	GeoIPProviderMaxMind GeoIPProvider = "maxmind"
	// GeoIPProviderCloudflareHeader trusts a CF-IPCountry (or configurable
	// header name) header injected by a Cloudflare-fronted deployment.
	GeoIPProviderCloudflareHeader GeoIPProvider = "cloudflare-header"
)

// GeoIPConfig configures IP-based country detection.
type GeoIPConfig struct {
	Provider GeoIPProvider `yaml:"provider"`

	// MaxMindDBPath is required when Provider == "maxmind" — a filesystem
	// path to a GeoLite2-Country (or GeoIP2-Country) .mmdb file.
	MaxMindDBPath string `yaml:"maxMindDBPath"`

	// HeaderName is the header read when Provider == "cloudflare-header".
	// Defaults to "CF-IPCountry" if empty.
	HeaderName string `yaml:"headerName"`
}

// Config is the localization module's configuration.
type Config struct {
	// DefaultCountryCode is used whenever a country cannot otherwise be
	// resolved (no account-country resolver registered / no match, IP
	// lookup inconclusive or disabled, or the resolved country isn't
	// active). Must be a code present in the countries table. Defaults to
	// "ZZ" ("Global", USD, seeded by migration 000002) rather than any
	// specific country — a generic framework module has no basis for
	// assuming a host app's home country, and "Global" is honest about
	// being a fallback rather than silently impersonating one country's
	// data. Host apps with a strong home-market bias (e.g. before their
	// own GeoIP is configured) can override this back to their own
	// country code.
	DefaultCountryCode string `yaml:"defaultCountryCode" validate:"required,len=2"`

	GeoIP GeoIPConfig `yaml:"geoIP"`
}

func Default() *Config {
	return &Config{
		DefaultCountryCode: "ZZ",
		GeoIP: GeoIPConfig{
			Provider:   GeoIPProviderNone,
			HeaderName: "CF-IPCountry",
		},
	}
}

func (c *Config) Key() string {
	return "localization"
}

func (c *Config) Validate() error {
	if len(c.DefaultCountryCode) != 2 {
		return fmt.Errorf("localization: defaultCountryCode must be a 2-letter ISO code, got %q", c.DefaultCountryCode)
	}

	switch c.GeoIP.Provider {
	case "", GeoIPProviderNone:
		// no-op, detection disabled
	case GeoIPProviderMaxMind:
		if c.GeoIP.MaxMindDBPath == "" {
			return fmt.Errorf("localization: geoIP.maxMindDBPath is required when geoIP.provider is %q", GeoIPProviderMaxMind)
		}
	case GeoIPProviderCloudflareHeader:
		// HeaderName defaults below if empty; nothing to validate.
	default:
		return fmt.Errorf("localization: unknown geoIP.provider %q", c.GeoIP.Provider)
	}

	if c.GeoIP.HeaderName == "" {
		c.GeoIP.HeaderName = "CF-IPCountry"
	}

	return nil
}
