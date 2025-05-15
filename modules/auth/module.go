package auth

import (
	"context"
	"embed"
	"fmt"

	"github.com/gin-gonic/gin"
	migrationRunner "gitlab.com/jideobs/nebularcore/core/migration_runner"
	"gitlab.com/jideobs/nebularcore/core/module"
	"gitlab.com/jideobs/nebularcore/modules/auth/backends"
	"gitlab.com/jideobs/nebularcore/modules/auth/config"
	"gitlab.com/jideobs/nebularcore/modules/auth/factories"
	"gitlab.com/jideobs/nebularcore/modules/auth/handlers"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces"
	"gitlab.com/jideobs/nebularcore/modules/auth/middleware"
	"gitlab.com/jideobs/nebularcore/modules/auth/repositories"
	"gitlab.com/jideobs/nebularcore/modules/auth/state"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Module implements the NebularCore module interface for authentication
type Module struct {
	name            string
	version         string
	config          *config.Config
	authManager     backends.AuthenticationManager
	authHandler     interfaces.AuthHandler
	passwordHandler interfaces.PasswordHandler
	authMiddleware  interfaces.AuthMiddleware
	tokenIssuer     interfaces.TokenIssuer
	userRepository  interfaces.UserRepository
}

// New creates a new authentication module
func New() *Module {
	return &Module{
		name:    "auth",
		version: "1.0.0",
		config:  config.Default(),
	}
}

// Name returns the module name
func (m *Module) Name() string {
	return m.name
}

// Version returns the module version
func (m *Module) Version() string {
	return m.version
}

func (m *Module) MigrationsDir() string {
	return "migrations"
}

// Dependencies returns the module dependencies
func (m *Module) Dependencies() []string {
	return []string{}
}

// Initialize initializes the auth module
func (m *Module) Initialize(
	ctx context.Context,
	db *gorm.DB,
	router *gin.Engine,
) error {
	if err := m.initializeDefaults(db); err != nil {
		return err
	}

	// Initialize and register authentication backends based on configured backends
	for _, name := range m.config.Backends {
		switch name {
		case "local":
			m.authManager.RegisterBackend(
				name,
				backends.NewLocalBackend(m.userRepository, m.tokenIssuer),
			)
		default:
			return fmt.Errorf("unknown authentication backend: %s", name)
		}
	}

	// Register routes
	apiGroup := router.Group("")
	m.authHandler.RegisterRoutes(apiGroup)
	m.passwordHandler.RegisterRoutes(apiGroup)

	return nil
}

func (m *Module) initializeDefaults(db *gorm.DB) error {
	if m.authManager == nil {
		m.authManager = backends.NewAuthenticationManager()
	}
	if m.userRepository == nil {
		m.userRepository = repositories.NewUserRepository(db, factories.NewDefaultUserFactory())
	}
	if m.tokenIssuer == nil {
		m.tokenIssuer = state.NewJWTTokenIssuer(m.config.JWT)
	}
	if m.authHandler == nil {
		m.authHandler = handlers.NewAuthHandler(m.authManager, m.tokenIssuer, m.config)
	}
	if m.authMiddleware == nil {
		var err error
		m.authMiddleware, err = middleware.NewAuthMiddleware(m.authManager, &m.config.Middleware)
		if err != nil {
			return err
		}
	}
	if m.passwordHandler == nil {
		m.passwordHandler = handlers.NewPasswordHandler(m.userRepository, m.authManager, m.authMiddleware, m.config)
	}

	return nil
}

// Namespace implements module.Module.
func (m *Module) Namespace() module.ModuleNamespace {
	return module.PublicNamespace
}

// Configure configures the module
func (m *Module) Configure(cfg any) error {
	if cfg, ok := cfg.(*config.Config); ok {
		m.config = cfg
		return m.config.Validate()
	}
	return fmt.Errorf("invalid config type for auth module")
}

// NewConfig returns a new configuration instance
func (m *Module) NewConfig() any {
	return config.Default()
}

// Shutdown performs cleanup when the module is shutting down
func (m *Module) Shutdown(ctx context.Context) error {
	return nil
}

// RegisterRoutes registers the module's routes
func (m *Module) RegisterRoutes(router *gin.Engine) {
	apiGroup := router.Group("")
	m.authHandler.RegisterRoutes(apiGroup)
	m.passwordHandler.RegisterRoutes(apiGroup)
}

// Middleware returns the module's middleware
func (m *Module) Middleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		m.authMiddleware.Optional(),
	}
}

// GetAuthManager returns the authentication manager
func (m *Module) GetAuthManager() backends.AuthenticationManager {
	return m.authManager
}

// GetTokenIssuer returns the token issuer
func (m *Module) GetTokenIssuer() interfaces.TokenIssuer {
	return m.tokenIssuer
}

// GetAuthMiddleware returns the authentication middleware
func (m *Module) GetAuthMiddleware() interfaces.AuthMiddleware {
	return m.authMiddleware
}

func (m *Module) ProvidesMigrations() bool {
	return true
}

func (m *Module) GetMigrationSources(projectRoot string) []migrationRunner.Source {
	sources := []migrationRunner.Source{}

	if m.config.UserMigrationScriptPath != "" {
		sources = append(sources, migrationRunner.Source{
			Path:     fmt.Sprintf("file://%s", m.config.UserMigrationScriptPath),
			Priority: 100,
			Exclude:  []string{},
		})

		sources = append(sources, migrationRunner.Source{
			FS:       migrations,
			Path:     "migrations",
			Priority: 50,
			Exclude: []string{
				"000001_init_auth.up.sql",
				"000001_init_auth.down.sql",
			},
		})

		return sources
	}

	sources = append(sources, migrationRunner.Source{
		FS:       migrations,
		Path:     "migrations",
		Priority: 50,
		Exclude:  []string{},
	})

	return sources
}

// WithUserRepository sets the user repository
func (m *Module) WithUserRepository(repo interfaces.UserRepository) *Module {
	m.userRepository = repo
	return m
}

// WithTokenIssuer sets the token issuer
func (m *Module) WithTokenIssuer(issuer interfaces.TokenIssuer) *Module {
	m.tokenIssuer = issuer
	return m
}

// WithAuthManager sets the authentication manager
func (m *Module) WithAuthManager(manager backends.AuthenticationManager) *Module {
	m.authManager = manager
	return m
}

// WithConfig sets the module configuration
func (m *Module) WithConfig(cfg *config.Config) *Module {
	m.config = cfg
	return m
}

// WithAuthHandler sets the authentication handler
func (m *Module) WithAuthHandler(handler interfaces.AuthHandler) *Module {
	m.authHandler = handler
	return m
}

// WithAuthMiddleware sets the authentication middleware
func (m *Module) WithAuthMiddleware(middleware interfaces.AuthMiddleware) *Module {
	m.authMiddleware = middleware
	return m
}
