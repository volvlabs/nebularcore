package storage

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/volvlabs/nebularcore/core/migration_runner"
	"github.com/volvlabs/nebularcore/core/module"
	"github.com/volvlabs/nebularcore/modules/storage/gcs"
	"github.com/volvlabs/nebularcore/modules/storage/local"
	"github.com/volvlabs/nebularcore/modules/storage/models"
	"github.com/volvlabs/nebularcore/modules/storage/s3"
	"gorm.io/gorm"
)

type Config struct {
	Provider string      `yaml:"provider"`
	S3       *s3.Config  `yaml:"s3,omitempty"`
	GCS      *gcs.Config `yaml:"gcs,omitempty"`

	LocalPath string `yaml:"localPath,omitempty"`
	BaseURL   string `yaml:"baseUrl,omitempty"`
}

type Module struct {
	config   *Config
	provider StorageProvider
	mu       sync.RWMutex
}

func New() *Module {
	return &Module{}
}

// MigrationsDir implements module.Module.
func (m *Module) MigrationsDir() string {
	return ""
}

// Namespace implements module.Module.
func (m *Module) Namespace() module.ModuleNamespace {
	return module.PublicNamespace
}

// NewConfig implements module.Module.
func (m *Module) NewConfig() any {
	return &Config{}
}

// ProvidesMigrations implements module.Module.
func (m *Module) ProvidesMigrations() bool {
	return false
}

// Name returns the module name
func (m *Module) Name() string {
	return "storage"
}

// Version returns the module version
func (m *Module) Version() string {
	return "v1.0.0"
}

// Dependencies returns the module dependencies
func (m *Module) Dependencies() []string {
	return []string{}
}

// Initialize sets up the storage module
func (m *Module) Initialize(
	ctx context.Context,
	db *gorm.DB,
	router *gin.Engine,
) error {
	provider, err := m.initializeProvider()
	if err != nil {
		return fmt.Errorf("failed to initialize storage provider: %w", err)
	}

	m.provider = provider
	return nil
}

// Configure updates the module configuration
func (m *Module) Configure(config interface{}) error {
	if cfg, ok := config.(*Config); ok {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.config = cfg

		provider, err := m.initializeProvider()
		if err != nil {
			return fmt.Errorf("failed to reinitialize storage provider: %w", err)
		}
		m.provider = provider
		return nil
	}
	return fmt.Errorf("invalid configuration type")
}

// Shutdown cleans up resources
func (m *Module) Shutdown(ctx context.Context) error {
	return nil
}

// GetMigrationSources returns the migration sources for the storage module
func (m *Module) GetMigrationSources(projectRoot string) []migration_runner.Source {
	return nil // Storage module doesn't have any database migrations
}

// Upload handles file upload through the configured provider
func (m *Module) Upload(ctx context.Context, input *models.UploadInput) (*models.UploadOutput, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.provider.Upload(ctx, input)
}

// Download retrieves file content through the configured provider
func (m *Module) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.provider.Download(ctx, path)
}

// Delete removes a file through the configured provider
func (m *Module) Delete(ctx context.Context, path string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.provider.Delete(ctx, path)
}

// List returns files in a directory through the configured provider
func (m *Module) List(ctx context.Context, prefix string) ([]models.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.provider.List(ctx, prefix)
}

// initializeProvider creates a new storage provider based on configuration
func (m *Module) initializeProvider() (StorageProvider, error) {
	var provider StorageProvider
	var err error

	switch m.config.Provider {
	case "s3":
		if m.config.S3 == nil {
			return nil, fmt.Errorf("S3 configuration is required for S3 provider")
		}
		provider, err = s3.New(*m.config.S3)
		if err != nil {
			return nil, fmt.Errorf("failed to create S3 provider: %w", err)
		}
	case "gcs":
		if m.config.GCS == nil {
			return nil, fmt.Errorf("GCS configuration is required for GCS provider")
		}
		provider, err = gcs.New(*m.config.GCS)
		if err != nil {
			return nil, fmt.Errorf("failed to create GCS provider: %w", err)
		}
	case "local":
		provider, err = local.New(m.config.LocalPath, m.config.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create local provider: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", m.config.Provider)
	}

	return provider, nil
}
