package core

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/jideobs/nebularcore/core/config"
	"gitlab.com/jideobs/nebularcore/core/module"
	"gitlab.com/jideobs/nebularcore/core/utils"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// baseApp provides the default implementation of the App interface
type baseApp[T config.Settings] struct {
	opts     Options[T]
	ctx      context.Context
	cancel   context.CancelFunc
	config   *config.CoreConfig
	settings T
	registry *module.Registry
	router   *gin.Engine
	db       *gorm.DB

	bootstraped bool

	loader *config.ConfigLoader[T]
}

// Options for creating a new application
type Options[T config.Settings] struct {
	ConfigPath string
	EnvPrefix  string
}

// New creates a new application instance
func New[T config.Settings](opts Options[T]) (*baseApp[T], error) {
	loader := config.NewConfigLoader[T]()
	if err := loader.LoadFromFile(opts.ConfigPath); err != nil {
		return nil, err
	}

	if errors := loader.ValidateAll(); len(errors) > 0 {
		return nil, fmt.Errorf("configuration validation failed: %v", errors)
	}

	projectRoot, err := utils.GetProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to determine project root: %w", err)
	}

	app := &baseApp[T]{
		opts:     opts,
		config:   loader.GetCore(),
		settings: loader.GetProject(),
		registry: module.NewRegistry(),
		loader:   loader,
	}

	app.config.ProjectRoot = projectRoot

	return app, nil
}

// Bootstrap initializes the application and all registered modules
func (a *baseApp[T]) Bootstrap(ctx context.Context) error {
	if a.bootstraped {
		return nil
	}
	a.ctx, a.cancel = context.WithCancel(ctx)
	if err := a.configureModules(); err != nil {
		return err
	}

	if err := a.initDB(); err != nil {
		return fmt.Errorf("initializing database: %w", err)
	}
	for _, name := range a.registry.InitOrder() {
		module, _ := a.registry.Get(name)
		if err := module.Initialize(a.ctx, a.db, a.Router()); err != nil {
			return fmt.Errorf("initializing module %s: %w", name, err)
		}
	}

	a.bootstraped = true
	return nil
}

// Shutdown gracefully shuts down the application
func (a *baseApp[T]) Shutdown(ctx context.Context) error {
	order := a.registry.InitOrder()
	for i := len(order) - 1; i >= 0; i-- {
		module, _ := a.registry.Get(order[i])
		if err := module.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutting down module %s: %w", order[i], err)
		}
	}

	if a.db != nil {
		db, err := a.db.DB()
		if err != nil {
			return fmt.Errorf("getting database instance: %w", err)
		}
		if err := db.Close(); err != nil {
			return fmt.Errorf("closing database connection: %w", err)
		}
	}

	return nil
}

func (a *baseApp[T]) configureModules() error {
	for _, name := range a.registry.InitOrder() {
		log.Debug().Msgf("configuring module %s", name)
		module, _ := a.registry.Get(name)
		moduleConfig := module.NewConfig()
		if moduleConfig == nil {
			continue
		}

		err := a.loader.GetModuleConfig(name, moduleConfig)
		if err != nil {
			log.Err(err).Msgf("error getting module %s config", name)
			return fmt.Errorf("error getting module %s config: %w", name, err)
		}

		if err := module.Configure(moduleConfig); err != nil {
			log.Err(err).Msgf("error configuring module %s", name)
			return fmt.Errorf("error configuring module %s: %w", name, err)
		}
	}
	return nil
}

// Run starts the application
func (a *baseApp[T]) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", a.config.Server.Host, a.config.Server.Port),
		Handler:      a.Router(),
		ReadTimeout:  a.config.Server.ReadTimeout,
		WriteTimeout: a.config.Server.WriteTimeout,
	}

	errChan := make(chan error, 1)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errChan:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.config.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("error shutting down server: %w", err)
	}

	if err := a.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("error during application shutdown: %w", err)
	}

	return nil
}

// RegisterModule registers a new module with the application
func (a *baseApp[T]) RegisterModule(m module.Module) error {
	return a.registry.Register(m)
}

// GetModules retrieves all registered modules
func (a *baseApp[T]) GetModules() map[string]module.Module {
	return a.registry.GetModules()
}

func (a *baseApp[T]) GetModulesByNamespace(namespace module.ModuleNamespace) map[string]module.Module {
	return a.registry.GetModulesByNamespace(namespace)
}

// GetModule retrieves a registered module by name
func (a *baseApp[T]) GetModule(name string) (module.Module, bool) {
	return a.registry.Get(name)
}

// Config returns the core configuration
func (a *baseApp[T]) Config() *config.CoreConfig {
	return a.config
}

// Settings returns the project settings
func (a *baseApp[T]) Settings() T {
	return a.settings
}

// Router returns the Gin router instance
func (a *baseApp[T]) Router() *gin.Engine {
	if a.router == nil {
		a.router = gin.Default()
	}
	return a.router
}

// DB returns the database instance
func (a *baseApp[T]) DB() *gorm.DB {
	return a.db
}

// initDB initializes the database connection
func (a *baseApp[T]) initDB() error {
	var err error
	switch a.config.Database.Driver {
	case "postgres":
		a.db, err = a.initPostgresDB()
	case "sqlite":
		a.db, err = a.initSqliteDB()
	case "cloudsqlpostgres":
		a.db, err = a.initGoogleCloudSQL(a.config.Database)
	default:
		return fmt.Errorf("unsupported database type: %s", a.config.Database.Driver)
	}
	return err
}

func (a *baseApp[T]) initPostgresDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		a.config.Database.Host,
		a.config.Database.Username,
		a.config.Database.Password,
		a.config.Database.Name,
		a.config.Database.Port,
		a.config.Database.SSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}

	return db.Debug(), nil
}

func (a *baseApp[T]) initSqliteDB() (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(a.config.Database.Name), &gorm.Config{
		SkipDefaultTransaction: true,
	})
}

func (a *baseApp[T]) initGoogleCloudSQL(config config.DatabaseConfig) (*gorm.DB, error) {
	return gorm.Open(postgres.New(postgres.Config{
		DriverName: "cloudsqlpostgres",
		DSN: fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=%s TimeZone=UTC",
			config.Host, config.Username, config.Name, config.Password, config.SSLMode),
	}))
}
