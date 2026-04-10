package core

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/volvlabs/nebularcore/daos"
	"github.com/volvlabs/nebularcore/models"
	"github.com/volvlabs/nebularcore/tools/auth"
	"github.com/volvlabs/nebularcore/tools/aws/scheduler"
	"github.com/volvlabs/nebularcore/tools/eventclient"
	"github.com/volvlabs/nebularcore/tools/filesystem"
	"github.com/volvlabs/nebularcore/tools/security"
	"github.com/volvlabs/nebularcore/tools/validation"
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
