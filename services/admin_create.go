package services

import (
	"gitlab.com/volvlabs/nebularcore/daos"
	"gitlab.com/volvlabs/nebularcore/models"
	"gitlab.com/volvlabs/nebularcore/tools/security"
	"gitlab.com/volvlabs/nebularcore/tools/types"
	"gitlab.com/volvlabs/nebularcore/tools/validation"

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
	dao       *daos.Dao
	validator *validation.Validator
}

func NewAdminCreate(dao *daos.Dao, validator *validation.Validator) *AdminCreate {
	return &AdminCreate{dao, validator}
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

	hashedPassword, err := security.HashPassword(adminCreateRequest.Password)
	if err != nil {
		log.Err(err).Msgf("AdminCreate: could not hash password")
		return nil, err
	}

	admin := &models.Admin{
		FirstName:    adminCreateRequest.FirstName,
		LastName:     adminCreateRequest.LastName,
		Email:        adminCreateRequest.Email,
		Role:         adminCreateRequest.Role,
		PasswordHash: hashedPassword,
	}

	err = a.dao.CreateAdmin(admin)
	if err != nil {
		return nil, err
	}

	return admin, nil
}
