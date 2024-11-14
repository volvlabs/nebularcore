package core

import (
	"fmt"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gitlab.com/jideobs/nebularcore/daos"
	"gitlab.com/jideobs/nebularcore/models"
	"gitlab.com/jideobs/nebularcore/models/config"
	"gitlab.com/jideobs/nebularcore/tools/auth"
	"gitlab.com/jideobs/nebularcore/tools/aws"
	"gitlab.com/jideobs/nebularcore/tools/aws/scheduler"
	"gitlab.com/jideobs/nebularcore/tools/eventclient"
	"gitlab.com/jideobs/nebularcore/tools/filesystem"
	"gitlab.com/jideobs/nebularcore/tools/gcloud"
	"gitlab.com/jideobs/nebularcore/tools/security"
	"gitlab.com/jideobs/nebularcore/tools/types"
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

	eventClient eventclient.Client
	scheduler   scheduler.Client

	fs *filesystem.System
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
	if b.fs != nil {
		return b.fs, nil
	}

	settings := b.Settings()
	var err error
	if settings.S3.Enabled {
		b.fs, err = filesystem.NewWithS3(
			settings.S3.Bucket,
			settings.Aws.Region,
			settings.Aws.AccessKeyID,
			settings.Aws.SecretAccessKey,
		)
	} else if settings.Glcoud.Storage.Enabled {
		b.fs, err = filesystem.NewWithGoogleCloudStorage(
			settings.Glcoud.Storage.Bucket,
			settings.Glcoud.Storage.CredfileLocation,
		)
	} else if b.Env == "test" || settings.InMemory.Enabled {
		b.fs, err = filesystem.NewMemory()
	} else {
		b.fs, err = filesystem.NewLocal(filepath.Join(b.DataDir(), LocalStorageDirName))
	}

	return b.fs, err
}

func (b *BaseApp) GetFileURL(key string) string {
	settings := b.Settings()
	if settings.S3.Enabled {
		cloudFrontConfig := settings.CloudFront
		return fmt.Sprintf("%s/%s", cloudFrontConfig.Domain, key)
	}

	return fmt.Sprintf("%s/files?key=%s", settings.Domain, key)
}

func (b *BaseApp) EventClient() eventclient.Client {
	if b.eventClient == nil {
		settings := b.Settings()
		switch settings.EventClient {
		case types.AWSEventBridgeClient:
			eventClient, err := aws.NewEventClient(
				settings.Aws.AccessKeyID,
				settings.Aws.SecretAccessKey,
				settings.Aws.Region,
				settings.EventBridge.EventBus,
			)
			if err != nil {
				log.Err(err).Msgf("failed to initialize event bridge client")
			}
			b.eventClient = eventClient
		case types.GcloudPubSubClient:
			eventClient, err := gcloud.NewEventClient(settings.Glcoud)
			if err != nil {
				log.Err(err).Msgf("failed to initialize gcloud pubsub client")
			}
			b.eventClient = eventClient
		}
	}

	return b.eventClient
}

func (b *BaseApp) Scheduler() scheduler.Client {
	if b.scheduler == nil {
		settings := b.Settings()
		scheduler, err := scheduler.New(
			settings.Aws.AccessKeyID,
			settings.Aws.SecretAccessKey,
			settings.Aws.Region,
		)
		if err != nil {
			log.Err(err).Msgf("failed to initialize scheduler client")
		}
		b.scheduler = scheduler
	}

	return b.scheduler
}
