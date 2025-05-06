package config

// Provider defines the interface for modules that provide configuration
type Provider interface {
	// ConfigKey returns the key under which this module's config should be stored
	ConfigKey() string

	// Validate validates the module's configuration
	Validate(config interface{}) error
}
