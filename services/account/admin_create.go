package account

import (
	"gitlab.com/jideobs/nebularcore/entities"
	"gitlab.com/jideobs/nebularcore/services/authentication"
	"gitlab.com/jideobs/nebularcore/tools/types"
	"gitlab.com/jideobs/nebularcore/tools/validation"

	"github.com/rs/zerolog/log"
)

type AdminCreateRequest struct {
	FirstName string     `json:"firstName" validate:"required"`
	LastName  string     `json:"lastName" validate:"required"`
	Email     string     `json:"email" validate:"required,email"`
	Role      types.Role `json:"role" validate:"required"`
	Password  string     `json:"password" validate:"required"`
}

func (a *Service) validate(adminCreateRequest AdminCreateRequest) error {
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

func (a *Service) CreateAdmin(adminCreateRequest AdminCreateRequest) (*entities.Admin, error) {
	if err := a.validate(adminCreateRequest); err != nil {
		return nil, err
	}

	admin := &entities.Admin{
		UserBase: entities.UserBase{
			FirstName: adminCreateRequest.FirstName,
			LastName:  adminCreateRequest.LastName,
			Email:     adminCreateRequest.Email,
			Role:      adminCreateRequest.Role,
			IsActive:  true,
		},
	}

	err := a.dao.CreateAdmin(admin)
	if err != nil {
		return nil, err
	}
	authService := authentication.New(a.app)
	err = authService.Create(
		adminCreateRequest.Email,
		adminCreateRequest.Password,
		"admins",
		types.Admin,
		admin.Id,
	)
	if err != nil {
		return nil, err
	}

	return admin, nil
}
