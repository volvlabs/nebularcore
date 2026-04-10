package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

var validate = validator.New()

// ConfigLoader handles loading and validation of all configurations
type ConfigLoader[T Settings] struct {
	raw     map[string]interface{}
	core    *CoreConfig
	modules map[string]any
	project T
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
	Section string
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader[T Settings]() *ConfigLoader[T] {
	return &ConfigLoader[T]{
		modules: make(map[string]any),
	}
}

// LoadFromFile loads configuration from a YAML/JSON file
func (l *ConfigLoader[T]) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	var raw RawConfig[T]
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}

	if err := validate.Struct(&raw); err != nil {
		return fmt.Errorf("validating config structure: %w", err)
	}

	l.raw = make(map[string]interface{})
	if err := yaml.Unmarshal(data, &l.raw); err != nil {
		return fmt.Errorf("storing raw config: %w", err)
	}
	l.core = &raw.Core
	l.project = raw.Project

	return nil
}

// ValidateAll performs validation on all configurations
func (l *ConfigLoader[T]) ValidateAll() []ValidationError {
	var errors []ValidationError

	if err := l.validateCore(); err != nil {
		errors = append(errors, ValidationError{
			Section: "core",
			Message: err.Error(),
		})
	}

	if err := l.validateProject(); err != nil {
		errors = append(errors, ValidationError{
			Section: "project",
			Message: err.Error(),
		})
	}

	return errors
}

func (l *ConfigLoader[T]) validateCore() error {
	return l.core.Validate()
}

func (l *ConfigLoader[T]) validateProject() error {
	return l.project.Validate()
}

// GetCore returns the core configuration
func (l *ConfigLoader[T]) GetCore() *CoreConfig {
	return l.core
}

// GetProject returns the project settings
func (l *ConfigLoader[T]) GetProject() T {
	return l.project
}

// GetModuleConfig retrieves configuration for a specific module
func (l *ConfigLoader[T]) GetModuleConfig(key string, config any) error {
	raw, ok := l.raw["modules"].(map[string]any)[key]
	if !ok {
		return fmt.Errorf("no configuration found for module: %s", key)
	}

	yamlBytes, err := yaml.Marshal(raw)
	if err != nil {
		return fmt.Errorf("marshaling module config: %w", err)
	}

	if err := yaml.Unmarshal(yamlBytes, config); err != nil {
		return fmt.Errorf("parsing module config: %w", err)
	}

	return nil
}
