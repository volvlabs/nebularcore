package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSettings implements Settings interface for testing
type TestSettings struct {
	Name        string `yaml:"name" validate:"required"`
	Environment string `yaml:"environment" validate:"required,oneof=development staging production"`
}

func (s TestSettings) Validate() error {
	return validate.Struct(s)
}

func (s TestSettings) IsProduction() bool {
	return s.Environment == "production"
}

// TestModuleConfig represents a test module configuration
type TestModuleConfig struct {
	Enabled bool   `yaml:"enabled"`
	Name    string `yaml:"name" validate:"required"`
}

func TestConfigLoader_LoadFromFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErr  bool
		validate func(*testing.T, *ConfigLoader[TestSettings])
	}{
		{
			name: "valid configuration",
			content: `
core:
  environment: development
  debug: true
  server:
    host: localhost
    port: "8080"
  database:
    driver: sqlite
    name: test.db
project:
  name: test-project
  environment: development
modules:
  test:
    enabled: true
    name: test-module
`,
			wantErr: false,
			validate: func(t *testing.T, loader *ConfigLoader[TestSettings]) {
				assert.Equal(t, "development", loader.GetCore().Environment)
				assert.True(t, loader.GetCore().Debug)
				assert.Equal(t, "localhost", loader.GetCore().Server.Host)
				assert.Equal(t, "8080", loader.GetCore().Server.Port)
				assert.Equal(t, "sqlite", loader.GetCore().Database.Driver)
				assert.Equal(t, "test.db", loader.GetCore().Database.Name)

				project := loader.GetProject()
				assert.Equal(t, "test-project", project.Name)
				assert.Equal(t, "development", project.Environment)

				// Test module config
				var moduleConfig TestModuleConfig
				err := loader.GetModuleConfig("test", &moduleConfig)
				require.NoError(t, err)
				assert.True(t, moduleConfig.Enabled)
				assert.Equal(t, "test-module", moduleConfig.Name)
			},
		},
		{
			name: "missing required fields",
			content: `
core:
  debug: true
project:
  name: test-project
`,
			wantErr: true,
			validate: func(t *testing.T, loader *ConfigLoader[TestSettings]) {
				errs := loader.ValidateAll()
				assert.NotEmpty(t, errs)
			},
		},
		{
			name: "invalid environment value",
			content: `
core:
  environment: invalid
  server:
    host: localhost
    port: "8080"
  database:
    driver: sqlite
    name: test.db
project:
  name: test-project
  environment: development
`,
			wantErr: true,
			validate: func(t *testing.T, loader *ConfigLoader[TestSettings]) {
				errs := loader.ValidateAll()
				assert.NotEmpty(t, errs)
				for _, err := range errs {
					if err.Section == "core" {
						return
					}
				}
				t.Error("expected core validation error")
			},
		},
		{
			name: "invalid yaml syntax",
			content: `
		core: {
		  invalid yaml
		}
		`,
			wantErr:  true,
			validate: func(t *testing.T, loader *ConfigLoader[TestSettings]) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			err := os.WriteFile(configPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			// Load configuration
			loader := NewConfigLoader[TestSettings]()
			err = loader.LoadFromFile(configPath)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			tt.validate(t, loader)
		})
	}
}

func TestConfigLoader_GetModuleConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
		module  string
		wantErr bool
	}{
		{
			name: "existing module",
			content: `
core:
  environment: development
  server:
    host: localhost
    port: "8080"
  database:
    driver: sqlite
    name: test.db
project:
  name: test-project
  environment: development
modules:
  test:
    enabled: true
    name: test-module
`,
			module:  "test",
			wantErr: false,
		},
		{
			name: "non-existent module",
			content: `
core:
  environment: development
  server:
    host: localhost
    port: "8080"
  database:
    driver: sqlite
    name: test.db
project:
  name: test-project
  environment: development
modules: {}
`,
			module:  "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			err := os.WriteFile(configPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			// Load configuration
			loader := NewConfigLoader[TestSettings]()
			err = loader.LoadFromFile(configPath)
			require.NoError(t, err)

			// Test module config
			var moduleConfig TestModuleConfig
			err = loader.GetModuleConfig(tt.module, &moduleConfig)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.True(t, moduleConfig.Enabled)
			assert.Equal(t, "test-module", moduleConfig.Name)
		})
	}
}
