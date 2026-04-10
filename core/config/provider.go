package config

// Provider defines the interface for modules that provide configuration
type Provider interface {
	ConfigKey() string
	Validate(config interface{}) error
}
