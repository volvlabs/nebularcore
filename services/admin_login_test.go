package services_test

import (
	"errors"
	"testing"

	"gitlab.com/volvlabs/nebularcore/models"
	"gitlab.com/volvlabs/nebularcore/services"
	"gitlab.com/volvlabs/nebularcore/test"
	"gitlab.com/volvlabs/nebularcore/tools/filesystem"
	"gitlab.com/volvlabs/nebularcore/tools/security"
)

func TestAdminLogin(t *testing.T) {
	app, _ := test.NewTestApp()
	hashedPassword, _ := security.HashPassword("XXXXXXXXXXX")
	admin := &models.Admin{
		FirstName:    "John",
		LastName:     "Doe",
		Email:        "john.doe@gmail.com",
		Role:         "operator",
		PasswordHash: hashedPassword,
	}
	scenarios := []struct {
		name    string
		req     services.AdminLoginRequest
		admin   *models.Admin
		wantErr error
	}{
		{
			name: "should login admin successfully",
			req: services.AdminLoginRequest{
				Identity: "john.doe@gmail.com",
				Password: "XXXXXXXXXXX",
			},
			admin:   admin,
			wantErr: nil,
		},
		{
			name: "should return error because admin was not found by identity",
			req: services.AdminLoginRequest{
				Identity: "john.doe@gmail.com",
				Password: "XXXXXXXXXXX",
			},
			admin:   nil,
			wantErr: errors.New("invalid login credentials"),
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			tearDownMigration := test.RunMigration(
				t, filesystem.GetRootDir("../../"), app.DataDir())
			defer tearDownMigration(t)
			if scenario.admin != nil {
				app.Dao().CreateAdmin(scenario.admin)
				app.Dao().CreateAuth(&models.Auth{
					Identity:     scenario.admin.Email,
					PasswordHash: hashedPassword,
				})
			}

			adminLogin := services.NewAdminLogin(app.Dao(), app.Validator())
			_, err := adminLogin.Submit(scenario.req)
			if err != nil && err.Error() != scenario.wantErr.Error() {
				t.Errorf("AdminService.Login() error = %v, wantErr %v", err, scenario.wantErr)
				return
			}
		})
	}
}
