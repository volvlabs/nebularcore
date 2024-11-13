package authentication_test

import (
	"errors"
	"github.com/google/uuid"
	"gitlab.com/jideobs/nebularcore/services/authentication"
	"gitlab.com/jideobs/nebularcore/tools/types"
	"testing"

	"gitlab.com/jideobs/nebularcore/test"
	"gitlab.com/jideobs/nebularcore/tools/filesystem"
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
			name:         "should successfully create authentication information",
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

			authService := authentication.New(app)
			err := authService.Create(
				scenario.identity, scenario.passwordHash, "admins", types.Admin, uuid.New())
			if err != nil && !errors.Is(err, scenario.wantErr) {
				t.Errorf("got %v, want %v", err, scenario.wantErr)
			}
		})
	}
}
