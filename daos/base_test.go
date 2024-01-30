package daos_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/jideobs/nebularcore/daos"
	"gitlab.com/jideobs/nebularcore/models"
	"gitlab.com/jideobs/nebularcore/test"
	"gitlab.com/jideobs/nebularcore/tools/filesystem"
)

func TestDao_Save(t *testing.T) {
	app, _ := test.NewTestApp()

	scenarios := []struct {
		name    string
		model   *models.Admin
		wantErr bool
	}{
		{
			name: "should create new entry in database successfully",
			model: &models.Admin{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@gmail.com",
			},
			wantErr: false,
		},
	}
	for _, tt := range scenarios {
		t.Run(tt.name, func(t *testing.T) {
			tearDownMigration := test.RunMigration(t, filesystem.GetRootDir("../"), app.DataDir())
			defer tearDownMigration(t)
			d := daos.New(app.Dao().DB())
			if err := d.Save(tt.model); (err != nil) != tt.wantErr {
				t.Errorf("Dao.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDao_SaveExistingRecordEntry(t *testing.T) {
	app, _ := test.NewTestApp()

	tearDownMigration := test.RunMigration(t, filesystem.GetRootDir("../"), app.DataDir())
	defer tearDownMigration(t)
	d := daos.New(app.Dao().DB())

	admin := &models.Admin{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@gmail.com",
	}
	err := d.Save(admin)
	if err != nil {
		t.Errorf("Dao.Save() error = %v", err)
	}
	admin.FirstName = "Jane"
	err = d.Save(admin)
	if err != nil {
		t.Errorf("Dao.Save() error = %v", err)
	}
}

func TestDao_FindRecord(t *testing.T) {
	app, _ := test.NewTestApp()
	scenarios := []struct {
		name           string
		recordToCreate *models.Admin
		where          *models.Admin
		wantErr        bool
	}{
		{
			name:           "should return error record not found",
			recordToCreate: nil,
			where: &models.Admin{
				Email: "test@gmail.com",
			},
			wantErr: true,
		},
		{
			name: "should return record successfully",
			recordToCreate: &models.Admin{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@gmail.com",
			},
			where: &models.Admin{
				Email: "john.doe@gmail.com",
			},
			wantErr: false,
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			d := daos.New(app.Dao().DB())
			tearDownMigration := test.RunMigration(t, filesystem.GetRootDir("../"), app.DataDir())
			defer tearDownMigration(t)

			if scenario.recordToCreate != nil {
				err := d.Save(scenario.recordToCreate)
				if err != nil {
					t.Fatalf("Dao.Save() error = %v", err)
				}
			}

			admin := models.Admin{}
			err := d.FindBy(&admin, scenario.where)
			if (err != nil) != scenario.wantErr {
				t.Errorf("Dao.FindRecord() error = %v, wantErr %v", err, scenario.wantErr)
			}
		})
	}
}

func TestDao_Delete(t *testing.T) {
	// Arrange:
	app, _ := test.NewTestApp()
	tearDownMigration := test.RunMigration(t, filesystem.GetRootDir("../"), app.DataDir())
	defer tearDownMigration(t)

	admin := &models.Admin{
		Email: "john.doe@gmail.com",
	}
	app.Dao().Save(admin)

	where := &models.Admin{}
	where.SetId(admin.Id)

	// Act:
	err := app.Dao().Delete(where)

	// Assert:
	assert.Equal(t, nil, err)
	adminGotten := models.Admin{}
	app.Dao().FindBy(&adminGotten, &models.Admin{Email: admin.Email})

	assert.Equal(t, true, adminGotten.IsDeleted)
	assert.NotNil(t, adminGotten.DeletedAt)
}

func TestDao_Updates(t *testing.T) {
	// Arrange:
	app, _ := test.NewTestApp()
	tearDownMigration := test.RunMigration(t, filesystem.GetRootDir("../"), app.DataDir())
	defer tearDownMigration(t)

	admin := &models.Admin{
		FirstName: "John",
		Email:     "john.doe@gmail.com",
	}
	app.Dao().Save(admin)
	assert.Equal(t, "John", admin.FirstName)
	assert.Equal(t, "", admin.LastName)

	where := &models.Admin{}
	where.SetId(admin.Id)
	updates := &models.Admin{
		FirstName: "Jane",
		LastName:  "Dawn",
	}

	// Act:
	err := app.Dao().Updates(where, updates)

	// Assert:
	assert.Equal(t, nil, err)

	adminGotten := models.Admin{}
	app.Dao().FindBy(&adminGotten, &models.Admin{Email: admin.Email})

	assert.Equal(t, "Jane", adminGotten.FirstName)
	assert.Equal(t, "Dawn", adminGotten.LastName)
}
