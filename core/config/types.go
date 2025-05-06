package config

import (
	"fmt"
	"time"
)

// Settings defines the base interface that all settings must implement
type Settings interface {
	Validate() error
	IsProduction() bool
}

// CoreConfig represents the essential framework configuration
type CoreConfig struct {
	Environment string         `yaml:"environment" json:"environment" validate:"required,oneof=development staging production test"`
	Debug       bool           `yaml:"debug" json:"debug"`
	Database    DatabaseConfig `yaml:"database" json:"database"`
	Server      ServerConfig   `yaml:"server" json:"server" validate:"required"`
	ProjectRoot string         `yaml:"-" json:"-"`
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	Driver   string `yaml:"driver" json:"driver" validate:"omitempty,oneof=postgres sqlite cloudsqlpostgres"`
	Host     string `yaml:"host" json:"host"`
	Port     string `yaml:"port" json:"port"`
	Name     string `yaml:"name" json:"name"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	SSLMode  string `yaml:"sslmode" json:"sslmode"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Host            string        `yaml:"host" json:"host" validate:"required"`
	Port            string        `yaml:"port" json:"port" validate:"required"`
	ReadTimeout     time.Duration `yaml:"readTimeout" json:"readTimeout"`
	WriteTimeout    time.Duration `yaml:"writeTimeout" json:"writeTimeout"`
	ShutdownTimeout time.Duration `yaml:"shutdownTimeout" json:"shutdownTimeout"`
}

// RawConfig represents the structure of the configuration file
type RawConfig[T Settings] struct {
	Core    CoreConfig     `yaml:"core" json:"core" validate:"required"`
	Modules map[string]any `yaml:"modules" json:"modules"`
	Project T              `yaml:"project" json:"project" validate:"required"`
}

func (c *CoreConfig) Validate() error {
	if c.Database.Driver != "" {
		c.Database.Enabled = true
		if c.Database.Name == "" {
			return fmt.Errorf("database name is required")
		}
	}
	return validate.Struct(c)
}

func (c *CoreConfig) IsProduction() bool {
	return c.Environment == "production"
}
