package config

// Config defines the interface for modules that provide configuration
type Config interface {
	Key() string
	Validate() error
}
