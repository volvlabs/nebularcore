package services_test

import (
	"testing"

	"gitlab.com/jideobs/nebularcore/services"
	"gitlab.com/jideobs/nebularcore/test"
	"gitlab.com/jideobs/nebularcore/tools/filesystem"
	"gitlab.com/jideobs/nebularcore/tools/types"
)

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
				t, filesystem.GetRootDir("../"), app.DataDir())
			defer tearDownMigration(t)

			auth := services.NewAuth(app)
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
