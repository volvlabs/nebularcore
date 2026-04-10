package authentication_test

import (
	"github.com/google/uuid"
	"github.com/volvlabs/nebularcore/services/authentication"
	"testing"

	"github.com/volvlabs/nebularcore/test"
	"github.com/volvlabs/nebularcore/tools/filesystem"
	"github.com/volvlabs/nebularcore/tools/types"
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
				t, filesystem.GetRootDir("../../"), app.DataDir())
			defer tearDownMigration(t)

			authService := authentication.New(app)
			err := authService.Create(
				scenario.identity, scenario.oldPassword, "admins", types.Admin, uuid.New())
			if err != nil {
				t.Fatalf("Tyring to create authentication, got %v", err)
			}

			err = authService.ChangePassword(
				scenario.identity,
				scenario.userEnteredPassword,
				scenario.password)
			if err != nil && err.Error() != scenario.wantErr.Error() {
				t.Errorf("got error %v, want %v", err, scenario.wantErr)
			}
		})
	}
}
