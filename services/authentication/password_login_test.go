package authentication_test

import (
	"github.com/volvlabs/nebularcore/entities"
	"github.com/volvlabs/nebularcore/models/requests"
	"github.com/volvlabs/nebularcore/services/authentication"
	"testing"

	"github.com/volvlabs/nebularcore/test"
	"github.com/volvlabs/nebularcore/tools/filesystem"
	"github.com/volvlabs/nebularcore/tools/types"
)

func TestPasswordLogin(t *testing.T) {
	app, _ := test.NewTestApp()
	scenarios := []struct {
		name                string
		identity            string
		password            string
		userEnteredIdentity string
		userEnteredPassword string
		wantErr             error
	}{
		{
			name:                "should successfully login",
			identity:            "XXXXXXXXXXXXXXXXXX",
			password:            "XXXXXXXX",
			userEnteredIdentity: "XXXXXXXXXXXXXXXXXX",
			userEnteredPassword: "XXXXXXXX",
			wantErr:             nil,
		},
		{
			name:                "should return error because password is incorrect",
			identity:            "XXXXXXXXXXXXXXXXXX",
			password:            "XXXXXXXX",
			userEnteredIdentity: "XXXXXXXXXXXXXXXXXX",
			userEnteredPassword: "wrongPassword",
			wantErr:             &types.UserError{Message: "invalid login credentials"},
		},
		{
			name:                "should return error because authentication was not found with the entered identity",
			identity:            "XXXXXXXXXXXXXXXXXX",
			password:            "XXXXXXXX",
			userEnteredIdentity: "inexistent-identity",
			userEnteredPassword: "wrongPassword",
			wantErr:             &types.UserError{Message: "invalid login credentials"},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Arrange:
			tearDownMigration := test.RunMigration(
				t, filesystem.GetRootDir("../../"), app.DataDir())
			defer tearDownMigration(t)

			admin := &entities.Admin{
				UserBase: entities.UserBase{
					FirstName:   "John",
					LastName:    "Doe",
					Email:       "john.doe@example.com",
					Role:        types.Admin,
					PhoneNumber: "2348091607291",
					IsActive:    true,
				},
			}
			err := app.Dao().CreateAdmin(admin)
			if err != nil {
				t.Fatal(err)
			}

			authService := authentication.New(app)
			err = authService.Create(scenario.identity, scenario.password, "admins", types.Admin, admin.Id)
			if err != nil {
				t.Fatalf("Error creating authentication, got %v", err)
			}

			// Act:
			_, err = authService.PasswordLogin(requests.LoginRequest{
				scenario.userEnteredIdentity, scenario.userEnteredPassword})
			if err != nil && err.Error() != scenario.wantErr.Error() {
				t.Errorf("got error %v, want %v", err, scenario.wantErr)
			}
		})
	}
}
