package authentication_test

import (
	"errors"
	"github.com/google/uuid"
	"github.com/volvlabs/nebularcore/services/authentication"
	"github.com/volvlabs/nebularcore/tools/types"
	"testing"

	"github.com/volvlabs/nebularcore/test"
	"github.com/volvlabs/nebularcore/tools/filesystem"
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
