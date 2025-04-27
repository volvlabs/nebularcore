package core

import (
	"context"

	"github.com/gin-gonic/gin"
	"gitlab.com/jideobs/nebularcore/daos"
	"gitlab.com/jideobs/nebularcore/models"
	"gitlab.com/jideobs/nebularcore/tools/auth"
	"gitlab.com/jideobs/nebularcore/tools/aws/scheduler"
	"gitlab.com/jideobs/nebularcore/tools/eventclient"
	"gitlab.com/jideobs/nebularcore/tools/filesystem"
	"gitlab.com/jideobs/nebularcore/tools/security"
	"gitlab.com/jideobs/nebularcore/tools/validation"
	"gorm.io/gorm"
)

type TerminateHandler func() error

type App interface {
	IsDev() bool
	IsACLEnforced() bool
	Bootstrap() error
	OnTerminate(TerminateHandler)
	DataDir() string
	MigrationsDir() string
	Terminate() error
	Acm() *auth.AccessControlManager
	Dao() *daos.Dao
	Settings() *models.Settings
	Validator() *validation.Validator
	Router() *gin.Engine
	Otp() *security.Otp
	NewFileSystem() (*filesystem.System, error)
	GetFileURL(key string) string
	EventClient() eventclient.Client
	Scheduler() scheduler.Client
	SchemaName(tenantId string) string
	DBSessionFromContext(ctx context.Context) *gorm.DB
	RegisterEventClient(eventClinet eventclient.Client)
}
