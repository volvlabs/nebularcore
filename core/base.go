package core

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gitlab.com/volvlabs/nebularcore/daos"
	"gitlab.com/volvlabs/nebularcore/models"
	"gitlab.com/volvlabs/nebularcore/models/config"
	"gitlab.com/volvlabs/nebularcore/tools/auth"
	"gitlab.com/volvlabs/nebularcore/tools/validation"

	"gorm.io/gorm"
)

type BaseApp struct {
	Env           string
	isDev         bool
	enforceAcl    bool
	dataDir       string
	migrationsDir string

	onTerminateHandler TerminateHandler
	databaseConfig     config.DatabaseConfig

	aclManager *auth.AccessControlManager
	validator  *validation.Validator
	dao        *daos.Dao

	settings *models.Settings
	router   *gin.Engine
}

type BaseAppConfig struct {
	Env            string
	IsDev          bool
	EnforceAcl     bool
	DataDir        string
	MigrationsDir  string
	DatabaseConfig config.DatabaseConfig
}

func NewBaseApp(config BaseAppConfig) *BaseApp {
	baseApp := &BaseApp{
		Env:            config.Env,
		isDev:          config.IsDev,
		enforceAcl:     config.EnforceAcl,
		dataDir:        config.DataDir,
		migrationsDir:  config.MigrationsDir,
		databaseConfig: config.DatabaseConfig,
		settings:       models.NewSettings(),
		validator:      validation.New(),
		router:         gin.Default(),
	}

	if config.EnforceAcl {
		baseApp.aclManager = auth.NewAccessControlManager()
	}

	return baseApp
}

func (app *BaseApp) IsDev() bool {
	return app.isDev
}

func (app *BaseApp) initDataDB() error {
	var dbConn *gorm.DB
	var err error
	if app.Env == "test" {
		dbConn, err = connectSqliteDB(filepath.Join(app.DataDir(), "data.db"))
	} else {
		dbConn, err = connectPostgresDB(app.databaseConfig)
	}

	if err != nil {
		return err
	}

	if app.IsDev() {
		dbConn = dbConn.Debug()
	}

	app.dao = daos.New(dbConn)

	return nil
}

func (app *BaseApp) Bootstrap() error {
	if err := app.initDataDB(); err != nil {
		return err
	}

	return nil
}

func (b *BaseApp) OnTerminate(handler TerminateHandler) {
	b.onTerminateHandler = handler
}

func (b *BaseApp) Terminate() error {
	return b.onTerminateHandler()
}

func (b *BaseApp) IsACLEnforced() bool {
	return b.enforceAcl
}

func (b *BaseApp) Acm() *auth.AccessControlManager {
	return b.aclManager
}

func (b *BaseApp) Settings() *models.Settings {
	return b.settings
}

func (b *BaseApp) DataDir() string {
	return b.dataDir
}

func (b *BaseApp) MigrationsDir() string {
	return b.migrationsDir
}

func (b *BaseApp) Dao() *daos.Dao {
	return b.dao
}

func (b *BaseApp) Validator() *validation.Validator {
	return b.validator
}

func (b *BaseApp) Router() *gin.Engine {
	return b.router
}
