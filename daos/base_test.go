package daos_test

import (
	"gitlab.com/jideobs/nebularcore/entities"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/jideobs/nebularcore/daos"
	"gitlab.com/jideobs/nebularcore/test"
	"gitlab.com/jideobs/nebularcore/tools/filesystem"
)

func TestDao_Save(t *testing.T) {
	app, _ := test.NewTestApp()

	scenarios := []struct {
		name    string
		model   *entities.Admin
		wantErr bool
	}{
		{
			name: "should create new entry in database successfully",
			model: &entities.Admin{
				UserBase: entities.UserBase{
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john.doe@gmail.com",
				},
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

	admin := &entities.Admin{
		UserBase: entities.UserBase{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@gmail.com",
		},
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
		recordToCreate *entities.Admin
		where          *entities.Admin
		wantErr        bool
	}{
		{
			name:           "should return error record not found",
			recordToCreate: nil,
			where: &entities.Admin{
				UserBase: entities.UserBase{
					Email: "test@gmail.com",
				},
			},
			wantErr: true,
		},
		{
			name: "should return record successfully",
			recordToCreate: &entities.Admin{
				UserBase: entities.UserBase{
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john.doe@gmail.com",
				},
			},
			where: &entities.Admin{
				UserBase: entities.UserBase{
					Email: "john.doe@gmail.com",
				},
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

			admin := entities.Admin{}
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

	admin := &entities.Admin{
		UserBase: entities.UserBase{
			Email: "john.doe@gmail.com",
		},
	}
	app.Dao().Save(admin)

	where := &entities.Admin{}
	where.SetId(admin.Id)

	// Act:
	err := app.Dao().Delete(where)

	// Assert:
	assert.Equal(t, nil, err)
	adminGotten := entities.Admin{}
	app.Dao().FindBy(&adminGotten, &entities.Admin{
		UserBase: entities.UserBase{
			Email: admin.Email,
		}})

	assert.Equal(t, true, adminGotten.IsDeleted)
	assert.NotNil(t, adminGotten.DeletedAt)
}

func TestDao_Updates(t *testing.T) {
	// Arrange:
	app, _ := test.NewTestApp()
	tearDownMigration := test.RunMigration(t, filesystem.GetRootDir("../"), app.DataDir())
	defer tearDownMigration(t)

	admin := &entities.Admin{
		UserBase: entities.UserBase{
			FirstName: "John",
			Email:     "john.doe@gmail.com",
		},
	}
	app.Dao().Save(admin)
	assert.Equal(t, "John", admin.FirstName)
	assert.Equal(t, "", admin.LastName)

	where := &entities.Admin{}
	where.SetId(admin.Id)
	updates := &entities.Admin{
		UserBase: entities.UserBase{
			FirstName: "Jane",
			LastName:  "Dawn",
		},
	}

	// Act:
	err := app.Dao().Updates(where, updates)

	// Assert:
	assert.Equal(t, nil, err)

	adminGotten := entities.Admin{}
	app.Dao().FindBy(&adminGotten, &entities.Admin{
		UserBase: entities.UserBase{
			Email: admin.Email,
		},
	})

	assert.Equal(t, "Jane", adminGotten.FirstName)
	assert.Equal(t, "Dawn", adminGotten.LastName)
}
