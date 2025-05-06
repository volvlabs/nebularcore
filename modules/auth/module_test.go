package auth_test

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	migrationRunner "gitlab.com/jideobs/nebularcore/core/migration_runner"
	"gitlab.com/jideobs/nebularcore/modules/auth"
	"gitlab.com/jideobs/nebularcore/modules/auth/backends"
	"gitlab.com/jideobs/nebularcore/modules/auth/config"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces"
	eventMocks "gitlab.com/jideobs/nebularcore/modules/event/mocks"
	"gorm.io/gorm"
)

// Mock implementations
type mockAuthManager struct {
	mock.Mock
	backends.AuthenticationManager
}

type mockTokenIssuer struct {
	mock.Mock
	interfaces.TokenIssuer
}

type mockAuthHandler struct {
	mock.Mock
	interfaces.AuthHandler
}

type mockAuthMiddleware struct {
	mock.Mock
	interfaces.AuthMiddleware
}

type mockUserRepository struct {
	mock.Mock
	interfaces.UserRepository
}

func (m *mockAuthHandler) RegisterRoutes(group *gin.RouterGroup, h gin.HandlerFunc) {}

func (m *mockAuthHandler) SocialSignup(next gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		next(c)
	}
}

func (m *mockAuthHandler) InitiateSocialLogin(c *gin.Context) {}

func (m *mockAuthMiddleware) Optional() gin.HandlerFunc {
	return func(c *gin.Context) {}
}

func (m *mockAuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {}
}

func TestModule(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *auth.Module
		validate func(t *testing.T, module *auth.Module)
	}{
		{
			name: "New module initialization",
			setup: func() *auth.Module {
				eventBus := eventMocks.NewEventBus(t)
				return auth.New(eventBus)
			},
			validate: func(t *testing.T, module *auth.Module) {
				assert.NotNil(t, module)
				assert.Equal(t, "auth", module.Name())
				assert.Equal(t, "1.0.0", module.Version())
				assert.Equal(t, []string{"event"}, module.Dependencies())
				assert.NotEmpty(t, module.Namespace())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module := tt.setup()
			tt.validate(t, module)
		})
	}
}

func TestModuleConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		config      interface{}
		expectError bool
	}{
		{
			name: "valid configuration",
			config: func() *config.Config {
				cfg := config.Default()
				cfg.JWT.AccessTokenSecret = "test-secret"
				cfg.JWT.RefreshTokenSecret = "test-refresh-secret"
				return cfg
			}(),
			expectError: false,
		},
		{
			name:        "invalid configuration type",
			config:      "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module := auth.New(eventMocks.NewEventBus(t))
			err := module.Configure(tt.config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestModuleInitialization(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (*auth.Module, *gorm.DB, *gin.Engine)
		validate    func(t *testing.T, module *auth.Module, err error)
		expectError bool
	}{
		{
			name: "default configuration",
			setup: func() (*auth.Module, *gorm.DB, *gin.Engine) {
				module := auth.New(eventMocks.NewEventBus(t))
				cfg := config.Default()
				cfg.JWT.AccessTokenSecret = "test-secret"
				cfg.JWT.RefreshTokenSecret = "test-refresh-secret"
				module.WithConfig(cfg)

				// Set up mock dependencies
				mockAuthManager := &mockAuthManager{}
				mockTokenIssuer := &mockTokenIssuer{}
				mockAuthHandler := &mockAuthHandler{}
				mockAuthMiddleware := &mockAuthMiddleware{}
				mockUserRepo := &mockUserRepository{}

				module.WithAuthManager(mockAuthManager)
				module.WithTokenIssuer(mockTokenIssuer)
				module.WithAuthHandler(mockAuthHandler)
				module.WithAuthMiddleware(mockAuthMiddleware)
				module.WithUserRepository(mockUserRepo)

				return module, &gorm.DB{}, gin.New()
			},
			validate: func(t *testing.T, module *auth.Module, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, module.GetAuthManager())
				assert.NotNil(t, module.GetTokenIssuer())
				assert.NotNil(t, module.GetAuthMiddleware())
			},
			expectError: false,
		},
		{
			name: "custom components",
			setup: func() (*auth.Module, *gorm.DB, *gin.Engine) {
				module := auth.New(eventMocks.NewEventBus(t))
				cfg := config.Default()
				cfg.JWT.AccessTokenSecret = "test-secret"
				cfg.JWT.RefreshTokenSecret = "test-refresh-secret"
				module.WithConfig(cfg)

				mockAuthManager := &mockAuthManager{}
				mockTokenIssuer := &mockTokenIssuer{}
				mockAuthHandler := &mockAuthHandler{}
				mockAuthMiddleware := &mockAuthMiddleware{}
				mockUserRepo := &mockUserRepository{}

				module.WithAuthManager(mockAuthManager)
				module.WithTokenIssuer(mockTokenIssuer)
				module.WithAuthHandler(mockAuthHandler)
				module.WithAuthMiddleware(mockAuthMiddleware)
				module.WithUserRepository(mockUserRepo)

				return module, &gorm.DB{}, gin.New()
			},
			validate: func(t *testing.T, module *auth.Module, err error) {
				assert.NoError(t, err)
				_, ok := module.GetAuthManager().(*mockAuthManager)
				assert.True(t, ok)
				_, ok = module.GetTokenIssuer().(*mockTokenIssuer)
				assert.True(t, ok)
				_, ok = module.GetAuthMiddleware().(*mockAuthMiddleware)
				assert.True(t, ok)
			},
			expectError: false,
		},
		{
			name: "unknown backend",
			setup: func() (*auth.Module, *gorm.DB, *gin.Engine) {
				module := auth.New(eventMocks.NewEventBus(t))
				cfg := config.Default()
				cfg.JWT.AccessTokenSecret = "test-secret"
				cfg.JWT.RefreshTokenSecret = "test-refresh-secret"
				cfg.Middleware = config.MiddlewareConfig{
					AuthorizationEnabled: true,
					PermissionModelPath:  "./middleware/test-data/test-model.conf",
					PermissionPolicyPath: "./middleware/test-data/test-policy.csv",
				}
				cfg.Backends = []string{"unknown"}
				module.WithConfig(cfg)
				return module, &gorm.DB{}, gin.New()
			},
			validate: func(t *testing.T, module *auth.Module, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unknown authentication backend")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, db, router := tt.setup()
			err := module.Initialize(context.Background(), db, router)
			tt.validate(t, module, err)
		})
	}
}

func TestModuleMigrations(t *testing.T) {
	tests := []struct {
		name            string
		setup           func() *auth.Module
		projectRoot     string
		expectedSources func(t *testing.T, sources []migrationRunner.Source)
	}{
		{
			name: "default migration sources",
			setup: func() *auth.Module {
				return auth.New(eventMocks.NewEventBus(t))
			},
			projectRoot: "/project/root",
			expectedSources: func(t *testing.T, sources []migrationRunner.Source) {
				assert.Len(t, sources, 1)
				assert.Equal(t, "migrations", sources[0].Path)
				assert.Equal(t, 50, sources[0].Priority)
				assert.Empty(t, sources[0].Exclude)
			},
		},
		{
			name: "custom user migration path",
			setup: func() *auth.Module {
				module := auth.New(eventMocks.NewEventBus(t))
				cfg := config.Default()
				cfg.UserMigrationScriptPath = "/custom/path"
				module.WithConfig(cfg)
				return module
			},
			projectRoot: "/project/root",
			expectedSources: func(t *testing.T, sources []migrationRunner.Source) {
				assert.Len(t, sources, 2)
				assert.Equal(t, "file:///custom/path", sources[0].Path)
				assert.Equal(t, 100, sources[0].Priority)
				assert.Empty(t, sources[0].Exclude)

				assert.Equal(t, "migrations", sources[1].Path)
				assert.Equal(t, 50, sources[1].Priority)
				assert.Contains(t, sources[1].Exclude, "000001_init_auth.up.sql")
				assert.Contains(t, sources[1].Exclude, "000001_init_auth.down.sql")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module := tt.setup()
			assert.True(t, module.ProvidesMigrations())
			sources := module.GetMigrationSources(tt.projectRoot)
			tt.expectedSources(t, sources)
		})
	}
}

func TestModuleMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *auth.Module
		expectedLength int
	}{
		{
			name: "default middleware",
			setup: func() *auth.Module {
				return auth.New(eventMocks.NewEventBus(t)).
					WithAuthMiddleware(&mockAuthMiddleware{})
			},
			expectedLength: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module := tt.setup()
			middleware := module.Middleware()
			assert.Len(t, middleware, tt.expectedLength)
		})
	}
}

func TestModuleShutdown(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *auth.Module
		expectError bool
	}{
		{
			name: "successful shutdown",
			setup: func() *auth.Module {
				return auth.New(eventMocks.NewEventBus(t)).
					WithAuthMiddleware(&mockAuthMiddleware{})
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module := tt.setup()
			err := module.Shutdown(context.Background())
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
