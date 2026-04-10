package health

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	migrationRunner "github.com/volvlabs/nebularcore/core/migration_runner"
	"github.com/volvlabs/nebularcore/core/module"
	"github.com/volvlabs/nebularcore/modules/health/config"
	"gorm.io/gorm"
)

const (
	ModuleName    = "health"
	ModuleVersion = "1.0.0"
)

// Module implements the health check functionality
type Module struct {
	name      string
	version   string
	router    *gin.Engine
	config    *config.Config
	startTime time.Time
	checks    map[string]*Check
	results   map[string]*CheckResult
	mu        sync.RWMutex
	shutdown  chan struct{}
}

// MigrationsDir implements module.Module.
func (m *Module) MigrationsDir() string {
	return ""
}

func (m *Module) GetMigrationSources(projectRoot string) []migrationRunner.Source {
	return []migrationRunner.Source{}
}

// Namespace implements module.Module.
func (m *Module) Namespace() module.ModuleNamespace {
	return module.PublicNamespace
}

// Configure implements module.Module.
func (m *Module) Configure(cfg any) error {
	if cfg, ok := cfg.(*config.Config); ok {
		m.config = cfg
		return m.config.Validate()
	}
	return fmt.Errorf("invalid config type for auth module")
}

// NewConfig returns a new configuration instance
func (m *Module) NewConfig() any {
	return &config.Config{}
}

// Dependencies implements module.Module.
func (m *Module) Dependencies() []string {
	// Health module has no dependencies
	return nil
}

// Initialize implements module.Module.
func (m *Module) Initialize(ctx context.Context, db *gorm.DB, router *gin.Engine) error {
	m.router = router
	m.startTime = time.Now()
	m.checks = make(map[string]*Check)
	m.results = make(map[string]*CheckResult)
	m.shutdown = make(chan struct{})

	// Add default checks
	if err := m.AddCheck("system", m.systemCheck()); err != nil {
		return fmt.Errorf("failed to add system check: %w", err)
	}
	if db != nil {
		if err := m.AddCheck("database", m.databaseCheck(db)); err != nil {
			return fmt.Errorf("failed to add database check: %w", err)
		}
	}

	// Start check scheduler
	go m.runChecks()

	m.registerRoutes(router)
	return nil
}

// Name implements module.Module.
func (m *Module) Name() string {
	return m.name
}

// Shutdown implements module.Module.
func (m *Module) Shutdown(ctx context.Context) error {
	close(m.shutdown)
	return nil
}

// Version implements module.Module.
func (m *Module) Version() string {
	return m.version
}

// ProvidesMigrations implements module.Module.
func (m *Module) ProvidesMigrations() bool {
	return false
}

// New creates a new health check module
func New() *Module {
	m := &Module{
		name:      ModuleName,
		version:   ModuleVersion,
		startTime: time.Now(),
		checks:    make(map[string]*Check),
		results:   make(map[string]*CheckResult),
		shutdown:  make(chan struct{}),
	}
	return m
}

// registerRoutes sets up the health check endpoints on the provided router
func (m *Module) registerRoutes(router *gin.Engine) {
	basePath := m.config.Path
	health := router.Group(basePath)
	{
		health.GET("/live", m.handleLiveness)
		health.GET("/ready", m.handleReadiness)
		health.GET("/status", m.handleStatus)
	}
}

// HealthStatus represents the health check response
type HealthStatus struct {
	Status    string                  `json:"status"`
	Uptime    string                  `json:"uptime"`
	Timestamp time.Time               `json:"timestamp"`
	Version   string                  `json:"version"`
	Latency   time.Duration           `json:"latency,omitempty"`
	Checks    map[string]*CheckResult `json:"checks,omitempty"`
}

// handleLiveness handles liveness probe requests
func (m *Module) handleLiveness(c *gin.Context) {
	status := HealthStatus{
		Status:    "UP",
		Uptime:    time.Since(m.startTime).String(),
		Timestamp: time.Now(),
		Version:   m.version,
	}

	c.JSON(http.StatusOK, status)
}

// handleReadiness handles readiness probe requests
func (m *Module) handleReadiness(c *gin.Context) {
	start := time.Now()

	m.mu.RLock()
	checks := make(map[string]*CheckResult)
	for name, result := range m.results {
		checks[name] = result
	}
	m.mu.RUnlock()

	status := HealthStatus{
		Status:    m.aggregateStatus(checks),
		Uptime:    time.Since(m.startTime).String(),
		Timestamp: time.Now(),
		Version:   m.version,
		Latency:   time.Since(start),
		Checks:    checks,
	}

	c.JSON(http.StatusOK, status)
}
