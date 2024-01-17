package services_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jideobs/nebularcore/models"
	"gitlab.com/jideobs/nebularcore/services"
	"gitlab.com/jideobs/nebularcore/test"
	"gitlab.com/jideobs/nebularcore/tools/filesystem"
	"gitlab.com/jideobs/nebularcore/tools/security"
	"gitlab.com/jideobs/nebularcore/tools/types"
)

func TestAdminPasswordChange(t *testing.T) {
	app, _ := test.NewTestApp()
	hashedPassword, _ := security.HashPassword("password123")
	admin := &models.Admin{
		FirstName:    "John",
		LastName:     "Doe",
		Email:        "john.doe@gmail.com",
		PasswordHash: hashedPassword,
	}
	scenarios := []struct {
		name    string
		admin   *models.Admin
		adminId uuid.UUID
		req     services.AdminPasswordChangeRequest
		wantErr error
	}{
		{
			name:    "should successfully change admin password",
			admin:   admin,
			adminId: uuid.Nil,
			req: services.AdminPasswordChangeRequest{
				Password:        "XXXXXXXXXXX",
				ConfirmPassword: "XXXXXXXXXXX",
				OldPassword:     "password123",
			},
			wantErr: nil,
		},
		{
			name:    "should fail to change password because old password is incorrect",
			admin:   admin,
			adminId: uuid.Nil,
			req: services.AdminPasswordChangeRequest{
				Password:        "XXXXXXXXXXX",
				ConfirmPassword: "XXXXXXXXXXX",
				OldPassword:     "incorrectPassword",
			},
			wantErr: &types.UserError{Message: "current password is incorrect"},
		},
		{
			name:    "should fail to change password because passwords do not match",
			admin:   admin,
			adminId: uuid.Nil,
			req: services.AdminPasswordChangeRequest{
				Password:        "XXXXXXXXXXX",
				ConfirmPassword: "NotSamePassword",
				OldPassword:     "password123",
			},
			wantErr: &types.UserError{Message: "password and confirm password don't match"},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			tearDownMigration := test.RunMigration(
				t, filesystem.GetRootDir("../../"), app.DataDir())
			defer tearDownMigration(t)

			app.Dao().CreateAdmin(scenario.admin)
			app.Dao().CreateAuth(&models.Auth{
				Identity:     scenario.admin.Email,
				PasswordHash: hashedPassword,
			})
			adminId := scenario.adminId
			if adminId == uuid.Nil {
				adminId = scenario.admin.Id
			}

			adminPasswordChange := services.NewAdminPasswordChange(app.Dao(), app.Validator())
			err := adminPasswordChange.ChangePassword(adminId, scenario.req)

			assert.Equal(t, scenario.wantErr, err)
		})
	}
}
