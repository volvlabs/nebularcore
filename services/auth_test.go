package services_test

import (
	"testing"

	"gitlab.com/volvlabs/nebularcore/services"
	"gitlab.com/volvlabs/nebularcore/test"
	"gitlab.com/volvlabs/nebularcore/tools/filesystem"
	"gitlab.com/volvlabs/nebularcore/tools/types"
)

func TestCreate(t *testing.T) {
	app, _ := test.NewTestApp()

	scenarios := []struct {
		name         string
		identity     string
		passwordHash string
		wantErr      error
	}{
		{
			name:         "should successfully create auth information",
			identity:     "john.doe@gmail.com",
			passwordHash: "password",
			wantErr:      nil,
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			tearDownMigration := test.RunMigration(
				t, filesystem.GetRootDir("../../"), app.DataDir())
			defer tearDownMigration(t)

			auth := services.NewAuth(app.Dao())
			err := auth.Create(scenario.identity, scenario.passwordHash)
			if err != nil && err != scenario.wantErr {
				t.Errorf("got %v, want %v", err, scenario.wantErr)
			}
		})
	}
}

func TestChangePassword(t *testing.T) {
	app, _ := test.NewTestApp()
	scenarios := []struct {
		name                string
		identity            string
		password            string
		oldPassword         string
		userEnteredPassword string
		wantErr             error
	}{
		{
			name:                "should successfully change password",
			identity:            "john.doe@gmail.com",
			password:            "password",
			oldPassword:         "password123",
			userEnteredPassword: "password123",
			wantErr:             nil,
		},
		{
			name:                "should return error because current password is incorrect",
			identity:            "john.doe@gmail.com",
			password:            "password",
			oldPassword:         "password123",
			userEnteredPassword: "password12",
			wantErr:             &types.UserError{Message: "current password is incorrect"},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			tearDownMigration := test.RunMigration(
				t, filesystem.GetRootDir("../../"), app.DataDir())
			defer tearDownMigration(t)

			auth := services.NewAuth(app.Dao())
			err := auth.Create(scenario.identity, scenario.oldPassword)
			if err != nil {
				t.Fatalf("Tyring to create auth, got %v", err)
			}

			err = auth.ChangePassword(
				scenario.identity,
				scenario.userEnteredPassword,
				scenario.password)
			if err != nil && err.Error() != scenario.wantErr.Error() {
				t.Errorf("got error %v, want %v", err, scenario.wantErr)
			}
		})
	}
}

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
			name:                "should return error because auth was not found with the entered identity",
			identity:            "XXXXXXXXXXXXXXXXXX",
			password:            "XXXXXXXX",
			userEnteredIdentity: "inexistent-identity",
			userEnteredPassword: "wrongPassword",
			wantErr:             &types.UserError{Message: "invalid login credentials"},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			tearDownMigration := test.RunMigration(
				t, filesystem.GetRootDir("../../"), app.DataDir())
			defer tearDownMigration(t)

			auth := services.NewAuth(app.Dao())
			err := auth.Create(scenario.identity, scenario.password)
			if err != nil {
				t.Fatalf("Error creating auth, got %v", err)
			}

			err = auth.PasswordLogin(
				scenario.userEnteredIdentity, scenario.userEnteredPassword)
			if err != nil && err.Error() != scenario.wantErr.Error() {
				t.Errorf("got error %v, want %v", err, scenario.wantErr)
			}
		})
	}
}
