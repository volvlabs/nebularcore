package services

import (
	"gitlab.com/jideobs/nebularcore/core"
	"gitlab.com/jideobs/nebularcore/daos"
	"gitlab.com/jideobs/nebularcore/tools/types"
	"gitlab.com/jideobs/nebularcore/tools/validation"

	"github.com/google/uuid"
)

type AdminPasswordChangeRequest struct {
	OldPassword     string `json:"oldPassword" validate:"required"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirmPassword" validate:"required"`
}

type AdminPasswordChange struct {
	app       core.App
	dao       *daos.Dao
	validator *validation.Validator
}

func NewAdminPasswordChange(app core.App) *AdminPasswordChange {
	return &AdminPasswordChange{
		app:       app,
		dao:       app.Dao(),
		validator: app.Validator(),
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

	auth := Auth{a.app, a.dao, a.validator}
	return auth.ChangePassword(
		admin.Email,
		adminPasswordChangeRequest.OldPassword,
		adminPasswordChangeRequest.ConfirmPassword,
	)
}
