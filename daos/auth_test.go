package daos_test

import (
	"github.com/google/uuid"
	"github.com/volvlabs/nebularcore/entities"
	"testing"

	"github.com/volvlabs/nebularcore/test"
	"github.com/volvlabs/nebularcore/tools/filesystem"
	"github.com/volvlabs/nebularcore/tools/types"
)

func TestCreateAuth(t *testing.T) {
	app, _ := test.NewTestApp()

	tearDownMigration := test.RunMigration(t, filesystem.GetRootDir("../"), app.DataDir())
	defer tearDownMigration(t)
	d := app.Dao()

	scenarios := []struct {
		name string
		auth *entities.Auth
		want error
	}{
		{
			name: "should enter authentication info into db",
			auth: &entities.Auth{
				Identity:     "jideobs@gmail.com",
				UserId:       uuid.New(),
				PasswordHash: "password",
			},
			want: nil,
		},
		{
			name: "should fail to create admin because email already exists",
			auth: &entities.Auth{
				Identity:     "jideobs@gmail.com",
				UserId:       uuid.New(),
				PasswordHash: "password",
			},
			want: &types.UserError{Message: "identity already created"},
		},
	}
	for _, scenario := range scenarios {
		err := d.CreateAuth(scenario.auth)

		if err != nil && err != scenario.want {
			t.Errorf("got %v, want %v", err, scenario.want)
		}
	}
}

func TestUpdatePassword(t *testing.T) {
	app, _ := test.NewTestApp()

	tearDownMigration := test.RunMigration(t, filesystem.GetRootDir("../"), app.DataDir())
	defer tearDownMigration(t)
	d := app.Dao()

	scenarios := []struct {
		name            string
		identity        string
		newPasswordHash string
		want            error
	}{
		{
			name:            "should enter authentication info into db",
			identity:        "john.doe@gmail.com",
			newPasswordHash: "password",
			want:            nil,
		},
	}
	for _, scenario := range scenarios {
		err := d.UpdatePassword(scenario.identity, scenario.newPasswordHash)

		if err != nil && err != scenario.want {
			t.Errorf("got %v, want %v", err, scenario.want)
		}
	}
}

func TestFindAuthByIdentity(t *testing.T) {
	app, _ := test.NewTestApp()

	scenarios := []struct {
		name         string
		identity     string
		authToCreate *entities.Auth
		want         error
	}{
		{
			name:     "should successfully return authentication information",
			identity: "john.doe@gmail.com",
			authToCreate: &entities.Auth{
				Identity:     "john.doe@gmail.com",
				UserId:       uuid.New(),
				PasswordHash: "password",
			},
			want: nil,
		},
		{
			name:     "should return error because the identity does not found",
			identity: "john.doe@gmail.com",
			authToCreate: &entities.Auth{
				Identity:     "john.doe@gmail.com",
				UserId:       uuid.New(),
				PasswordHash: "password",
			},
			want: &types.UserError{Message: "authentication not found"},
		},
	}
	for _, scenario := range scenarios {
		tearDownMigration := test.RunMigration(
			t,
			filesystem.GetRootDir("../"),
			app.DataDir())
		defer tearDownMigration(t)
		d := app.Dao()

		if scenario.authToCreate != nil {
			err := d.CreateAuth(scenario.authToCreate)
			if err != nil {
				t.Fatalf("FindAuthByIdentity: got %v, want %v", err, scenario.want)
			}
		}

		auth, err := d.FindAuthByIdentity(scenario.identity)

		if err != nil && err != scenario.want {
			t.Errorf("got %v, want %v", err, scenario.want)
		}

		if auth != nil && auth.Identity != scenario.authToCreate.Identity &&
			auth.PasswordHash != scenario.authToCreate.PasswordHash {
			t.Errorf("got %v, want %v", auth, scenario.authToCreate)
		}
	}
}
