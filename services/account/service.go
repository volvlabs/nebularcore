package account

import (
	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/daos"
	"gitlab.com/jideobs/nebularcore/tools/validation"
)

type Service struct {
	app       core.App
	dao       *daos.Dao
	validator *validation.Validator
}

func New(app core.App) *Service {
	return &Service{app, app.Dao(), app.Validator()}
}
