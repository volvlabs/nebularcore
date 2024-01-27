package services_test

import (
	"testing"

	"gitlab.com/jideobs/nebularcore/services"
	"gitlab.com/jideobs/nebularcore/test"
	"gitlab.com/jideobs/nebularcore/tools/filesystem"
	"gitlab.com/jideobs/nebularcore/tools/types"
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

			auth := services.NewAuth(app)
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
