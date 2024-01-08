package daos_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/volvlabs/nebularcore/daos"
	"gitlab.com/volvlabs/nebularcore/models"
	"gitlab.com/volvlabs/nebularcore/test"
	"gitlab.com/volvlabs/nebularcore/tools/filesystem"
	"gitlab.com/volvlabs/nebularcore/tools/types"
)

func TestCreateAdmin(t *testing.T) {
	app, _ := test.NewTestApp()

	tearDownMigration := test.RunMigration(t, filesystem.GetRootDir("../../"), app.DataDir())
	defer tearDownMigration(t)
	d := app.Dao()

	scenarios := []struct {
		name  string
		admin *models.Admin
		want  error
	}{
		{
			name: "should successfully create admin",
			admin: &models.Admin{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
			},
			want: nil,
		},
		{
			name: "should fail to create admin because email already exists",
			admin: &models.Admin{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
			},
			want: &types.UserError{Message: "email already registered"},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			err := d.CreateAdmin(scenario.admin)

			if err != nil && err.Error() != scenario.want.Error() {
				t.Errorf("got %v, want %v", err, scenario.want)
			}
		})
	}
}

func TestSaveAdmin(t *testing.T) {
	app, _ := test.NewTestApp()

	tearDownMigration := test.RunMigration(t, filesystem.GetRootDir("../../"), app.DataDir())
	defer tearDownMigration(t)
	d := daos.New(app.Dao().DB())

	admin := &models.Admin{}

	err := d.SaveAdmin(admin)
	assert.Error(t, err)

	err = d.CreateAdmin(admin)
	assert.NoError(t, err)

	admin.FirstName = "Jide"

	err = d.SaveAdmin(admin)
	assert.NoError(t, err)
}

func TestFindAdminByEmail(t *testing.T) {
	app, _ := test.NewTestApp()

	scenarios := []struct {
		name          string
		email         string
		adminToCreate *models.Admin
		want          error
	}{
		{
			name:  "should successfully find admin by email",
			email: "john.doe@example.com",
			adminToCreate: &models.Admin{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
			},
			want: nil,
		},
		{
			name:          "should return admin not found error",
			email:         "john.doe@example.com",
			adminToCreate: nil,
			want:          &types.UserError{Message: "admin not found"},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			tearDownMigration := test.RunMigration(t, filesystem.GetRootDir("../../"), app.DataDir())
			defer tearDownMigration(t)
			d := app.Dao()

			if scenario.adminToCreate != nil {
				err := d.CreateAdmin(scenario.adminToCreate)
				if err != nil {
					t.Fatalf("CreateAdmin() had an error: %v", err)
				}
			}

			admin, err := d.FindAdminByEmail(scenario.email)
			if err != nil && err.Error() != scenario.want.Error() {
				t.Errorf("got %v, want %v", err, scenario.want)
			}
			if admin != nil && admin.Email != scenario.adminToCreate.Email {
				t.Errorf("got %v, want %v", admin, scenario.adminToCreate)
			}
		})
	}
}

func TestFindAdminById(t *testing.T) {
	app, _ := test.NewTestApp()

	scenarios := []struct {
		name          string
		id            uuid.UUID
		adminToCreate *models.Admin
		want          error
	}{
		{
			name: "should successfully find admin by ID",
			id:   uuid.New(),
			adminToCreate: &models.Admin{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
			},
			want: nil,
		},
		{
			name:          "should return admin not found error",
			id:            uuid.New(),
			adminToCreate: nil,
			want:          &types.UserError{Message: "admin not found"},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			tearDownMigration := test.RunMigration(t, filesystem.GetRootDir("../../"), app.DataDir())
			defer tearDownMigration(t)
			d := app.Dao()

			filterById := scenario.id
			if scenario.adminToCreate != nil {
				err := d.CreateAdmin(scenario.adminToCreate)
				if err != nil {
					t.Fatalf("CreateAdmin() had an error: %v", err)
				}
				filterById = scenario.adminToCreate.Id
			}

			admin, err := d.FindAdminById(filterById)
			if err != nil && err.Error() != scenario.want.Error() {
				t.Errorf("got %v, want %v", err, scenario.want)
			}
			if admin != nil && admin.Email != scenario.adminToCreate.Email {
				t.Errorf("got %v, want %v", admin, scenario.adminToCreate)
			}
		})
	}
}
