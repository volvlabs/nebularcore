package config

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Config represents the health module configuration
type Config struct {
	Enabled      bool          `yaml:"enabled" validate:"required" default:"true"`
	Path         string        `yaml:"path" validate:"required,startswith=/" default:"/health"`
	Interval     time.Duration `yaml:"interval" validate:"required,gt=0" default:"30s"`
	Timeout      time.Duration `yaml:"timeout" validate:"required,gt=0,ltfield=Interval" default:"5s"`
	InitialDelay time.Duration `yaml:"initialDelay" validate:"required,gte=0" default:"5s"`
}

func (c *Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}