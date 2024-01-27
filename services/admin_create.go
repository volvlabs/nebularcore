package services

import (
	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/daos"
	"gitlab.com/jideobs/nebularcore/models"
	"gitlab.com/jideobs/nebularcore/tools/types"
	"gitlab.com/jideobs/nebularcore/tools/validation"

	"github.com/rs/zerolog/log"
)

type AdminCreateRequest struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Role      string `json:"role" validate:"required"`
	Password  string `json:"password" validate:"required"`
}

type AdminCreate struct {
	app       core.App
	dao       *daos.Dao
	validator *validation.Validator
}

func NewAdminCreate(app core.App) *AdminCreate {
	return &AdminCreate{app, app.Dao(), app.Validator()}
}

func (a *AdminCreate) validate(adminCreateRequest AdminCreateRequest) error {
	fieldErrs, err := a.validator.Validate(adminCreateRequest)
	if err != nil {
		return &types.RequestBodyError{
			Message: "error validating request body",
			Errors:  fieldErrs,
		}
	}

	isValid, err := validation.ValidateEmail(adminCreateRequest.Email)
	if err != nil {
		log.Err(err).Msgf("AdminCreate: error occurred validating email %s", adminCreateRequest.Email)
		return &types.SystemError{Message: err.Error()}
	}

	if !isValid {
		return &types.UserError{Message: "email entered is invalid"}
	}

	return nil
}

func (a *AdminCreate) Create(adminCreateRequest AdminCreateRequest) (*models.Admin, error) {
	if err := a.validate(adminCreateRequest); err != nil {
		return nil, err
	}

	admin := &models.Admin{
		FirstName: adminCreateRequest.FirstName,
		LastName:  adminCreateRequest.LastName,
		Email:     adminCreateRequest.Email,
		Role:      adminCreateRequest.Role,
	}

	err := a.dao.CreateAdmin(admin)
	if err != nil {
		return nil, err
	}
	auth := Auth{a.app, a.dao, a.validator}
	err = auth.Create(
		adminCreateRequest.Email,
		adminCreateRequest.Password,
	)
	if err != nil {
		return nil, err
	}

	return admin, nil
}
