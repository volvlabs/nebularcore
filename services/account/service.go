package account

import (
	"github.com/volvlabs/nebularcore/core"
	"github.com/volvlabs/nebularcore/daos"
	"github.com/volvlabs/nebularcore/tools/validation"
)

type Service struct {
	app       core.App
	dao       *daos.Dao
	validator *validation.Validator
}

func New(app core.App) *Service {
	return &Service{app, app.Dao(), app.Validator()}
}
