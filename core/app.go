package core

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/volvlabs/nebularcore/daos"
	"gitlab.com/volvlabs/nebularcore/models"
	"gitlab.com/volvlabs/nebularcore/tools/auth"
	"gitlab.com/volvlabs/nebularcore/tools/validation"
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
}
