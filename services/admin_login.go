package services

import (
	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/daos"
	"gitlab.com/jideobs/nebularcore/models"
	"gitlab.com/jideobs/nebularcore/tools/types"
	"gitlab.com/jideobs/nebularcore/tools/validation"

	"github.com/rs/zerolog/log"
)

type AdminLoginRequest struct {
	Identity string `json:"identity" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AdminLogin struct {
	app       core.App
	dao       *daos.Dao
	validator *validation.Validator
}

func NewAdminLogin(app core.App) *AdminLogin {
	return &AdminLogin{app, app.Dao(), app.Validator()}
}

func (a *AdminLogin) validate(adminLoginRequest AdminLoginRequest) error {
	fieldErrs, err := a.validator.Validate(adminLoginRequest)
	if err != nil {
		return &types.RequestBodyError{
			Message: "error validating request body",
			Errors:  fieldErrs,
		}
	}

	isValid, err := validation.ValidateEmail(adminLoginRequest.Identity)
	if err != nil {
		log.Err(err).Msgf("AdminLogin: error occurred validating email %s", adminLoginRequest.Identity)
		return &types.SystemError{Message: err.Error()}
	}

	if !isValid {
		return &types.UserError{Message: "email entered is invalid"}
	}

	return nil
}

func (a *AdminLogin) Submit(adminLoginRequest AdminLoginRequest) (*models.Admin, error) {
	if err := a.validate(adminLoginRequest); err != nil {
		return nil, err
	}

	auth := Auth{a.app, a.dao, a.validator}
	if err := auth.PasswordLogin(adminLoginRequest.Identity, adminLoginRequest.Password); err != nil {
		return nil, err
	}

	admin, err := a.dao.FindAdminByEmail(adminLoginRequest.Identity)
	if err != nil && !types.ErrIsUserError(err) {
		return nil, err
	}

	return admin, nil
}
