package core

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gitlab.com/jideobs/nebularcore/daos"
	"gitlab.com/jideobs/nebularcore/models"
	"gitlab.com/jideobs/nebularcore/models/config"
	"gitlab.com/jideobs/nebularcore/tools/auth"
	"gitlab.com/jideobs/nebularcore/tools/filesystem"
	"gitlab.com/jideobs/nebularcore/tools/security"
	"gitlab.com/jideobs/nebularcore/tools/validation"

	"gorm.io/gorm"
)

const LocalStorageDirName string = "storage"

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

	otp *security.Otp
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
	if b.onTerminateHandler != nil {
		return b.onTerminateHandler()
	}

	return nil
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

func (b *BaseApp) Otp() *security.Otp {
	if b.otp == nil {
		b.otp = security.NewOtp(security.OtpOptions{
			Secret: b.Settings().OtpGenerationSecret,
			Period: b.Settings().OtpPeriod,
		})
	}
	return b.otp
}

func (b *BaseApp) NewFileSystem() (*filesystem.System, error) {
	settings := b.Settings()
	if settings.S3.Enabled {
		return filesystem.NewWithS3(
			settings.S3.Bucket,
			settings.Aws.Region,
			settings.Aws.AccessKeyID,
			settings.Aws.SecretAccessKey,
		)
	}

	return filesystem.NewLocal(filepath.Join(b.DataDir(), LocalStorageDirName))
}
