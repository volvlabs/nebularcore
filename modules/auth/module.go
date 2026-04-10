package auth

import (
	"context"
	"embed"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	migrationRunner "github.com/volvlabs/nebularcore/core/migration_runner"
	"github.com/volvlabs/nebularcore/core/module"
	"github.com/volvlabs/nebularcore/modules/auth/backends"
	"github.com/volvlabs/nebularcore/modules/auth/config"
	authEmitter "github.com/volvlabs/nebularcore/modules/auth/emitter"
	"github.com/volvlabs/nebularcore/modules/auth/factories"
	"github.com/volvlabs/nebularcore/modules/auth/handlers"
	"github.com/volvlabs/nebularcore/modules/auth/interfaces"
	"github.com/volvlabs/nebularcore/modules/auth/middleware"
	"github.com/volvlabs/nebularcore/modules/auth/password"
	"github.com/volvlabs/nebularcore/modules/auth/pkg"
	"github.com/volvlabs/nebularcore/modules/auth/repositories"
	"github.com/volvlabs/nebularcore/modules/auth/state"
	"github.com/volvlabs/nebularcore/modules/event"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Module implements the NebularCore module interface for authentication
type Module struct {
	name             string
	version          string
	config           *config.Config
	authManager      backends.AuthenticationManager
	authHandler      interfaces.AuthHandler
	passwordHandler  interfaces.PasswordHandler
	authMiddleware   interfaces.AuthMiddleware
	tokenIssuer      interfaces.TokenIssuer
	userRepository   interfaces.UserRepository
	socialRepository interfaces.SocialAccountRepository
	googleSignin     interfaces.GoogleSignin
	eventBus         event.Bus

	routerGroup         *gin.RouterGroup
	socialSignupHandler gin.HandlerFunc
}

// New creates a new authentication module
func New(eventBus event.Bus) *Module {
	return &Module{
		name:     "auth",
		version:  "1.0.0",
		config:   config.Default(),
		eventBus: eventBus,
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
	return []string{"event"}
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

	if err := m.registerConfiguredBackends(); err != nil {
		return err
	}

	m.authHandler = handlers.NewAuthHandler(
		m.authManager,
		m.tokenIssuer,
		m.googleSignin,
		m.config,
	)

	// Register routes
	m.routerGroup = router.Group("")
	m.authHandler.RegisterRoutes(m.routerGroup, m.socialSignupHandler)
	m.passwordHandler.RegisterRoutes(m.routerGroup)

	return nil
}

func (m *Module) registerConfiguredBackends() error {
	for _, name := range m.config.Backends {
		switch name {
		case "local":
			m.authManager.RegisterBackend(
				name,
				backends.NewLocalBackend(m.userRepository, m.tokenIssuer),
			)
			log.Debug().Msg("Local backend initialized")
		case "google":
			if !m.config.Social.Enabled {
				log.Warn().Msg("Google signin not initialized because social is disabled")
				continue
			}

			googleConfig, ok := m.config.Social.Providers["google"]
			if !ok {
				log.Warn().Msg("Google signin not initialized because config not found")
				continue
			}
			m.googleSignin = pkg.NewGoogleSignin(
				googleConfig.ClientID,
				googleConfig.ClientSecret,
				googleConfig.RedirectURL,
				googleConfig.Scopes,
			)
			m.authManager.RegisterBackend(
				name,
				backends.NewGoogleBackend(m.socialRepository, m.userRepository, m.tokenIssuer, m.googleSignin),
			)
			log.Debug().Msg("Google backend initialized")
		default:
			return fmt.Errorf("unknown authentication backend: %s", name)
		}
	}

	return nil
}

func (m *Module) initializeDefaults(db *gorm.DB) error {
	eventEmitter := authEmitter.NewEventEmitter(m.eventBus)
	if m.authManager == nil {
		m.authManager = backends.NewAuthenticationManager(eventEmitter)
	}
	if m.userRepository == nil {
		m.userRepository = repositories.NewUserRepository(db, factories.NewDefaultUserFactory())
	}
	if m.socialRepository == nil {
		m.socialRepository = repositories.NewSocialAccountRepository(db)
	}
	if m.tokenIssuer == nil {
		m.tokenIssuer = state.NewJWTTokenIssuer(m.config.JWT)
	}
	if m.authMiddleware == nil {
		var err error
		m.authMiddleware, err = middleware.NewAuthMiddleware(m.authManager, &m.config.Middleware)
		if err != nil {
			return err
		}
	}
	if m.passwordHandler == nil {
		passwordManager := password.NewManager(eventEmitter, m.userRepository)
		m.passwordHandler = handlers.NewPasswordHandler(passwordManager, m.authMiddleware, m.config)
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

func (m *Module) GoogleSignin() interfaces.GoogleSignin {
	return m.googleSignin
}

func (m *Module) WithSocialSignupHandler(handler gin.HandlerFunc) {
	if m.routerGroup == nil {
		panic("cannot setup social signup handler before module is initialized")
	}

	m.authHandler.RegisterSocialSignupRoutes(m.routerGroup, handler)
}
