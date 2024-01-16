package services

import (
	"gitlab.com/volvlabs/nebularcore/daos"
	"gitlab.com/volvlabs/nebularcore/tools/types"
	"gitlab.com/volvlabs/nebularcore/tools/validation"

	"github.com/google/uuid"
)

type AdminPasswordChangeRequest struct {
	OldPassword     string `json:"oldPassword" validate:"required"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirmPassword" validate:"required"`
}

type AdminPasswordChange struct {
	dao       *daos.Dao
	validator *validation.Validator
}

func NewAdminPasswordChange(dao *daos.Dao, validator *validation.Validator) *AdminPasswordChange {
	return &AdminPasswordChange{
		dao:       dao,
		validator: validator,
	}
}

func (a *AdminPasswordChange) validate(adminPasswordChangeRequest AdminPasswordChangeRequest) error {
	fieldErrs, err := a.validator.Validate(adminPasswordChangeRequest)
	if err != nil {
		return &types.RequestBodyError{
			Message: "error validating request body",
			Errors:  fieldErrs,
		}
	}
	if adminPasswordChangeRequest.Password != adminPasswordChangeRequest.ConfirmPassword {
		return &types.UserError{Message: "password and confirm password don't match"}
	}

	return nil
}

func (a *AdminPasswordChange) ChangePassword(adminId uuid.UUID, adminPasswordChangeRequest AdminPasswordChangeRequest) error {
	if err := a.validate(adminPasswordChangeRequest); err != nil {
		return err
	}

	admin, err := a.dao.FindAdminById(adminId)
	if err != nil {
		return err
	}

	auth := Auth{a.dao}
	return auth.ChangePassword(
		admin.Email,
		adminPasswordChangeRequest.OldPassword,
		adminPasswordChangeRequest.ConfirmPassword,
	)
}
