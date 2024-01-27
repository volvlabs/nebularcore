package services_test

import (
	"testing"

	"github.com/google/uuid"
	"gitlab.com/jideobs/nebularcore/services"
	"gitlab.com/jideobs/nebularcore/test"
	"gitlab.com/jideobs/nebularcore/tools/filesystem"
	"gitlab.com/jideobs/nebularcore/tools/types"
)

func TestCreateAdmin(t *testing.T) {
	scenarios := []struct {
		name    string
		req     services.AdminCreateRequest
		wantErr error
	}{
		{
			name: "should successfully create admin",
			req: services.AdminCreateRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Role:      "operator",
				Password:  "password123",
			},
			wantErr: nil,
		},
		{
			name: "should fail to create admin with invalid email",
			req: services.AdminCreateRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@turgish.com",
				Role:      "operator",
				Password:  "password123",
			},
			wantErr: &types.UserError{Message: "email entered is invalid"},
		},
		{
			name: "should fail to create admin with already existing email",
			req: services.AdminCreateRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "jogn.doe@gmail.com",
				Role:      "operator",
				Password:  "password123",
			},
			wantErr: &types.UserError{Message: "email already exists"},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			app, _ := test.NewTestApp()
			tearDownMigration := test.RunMigration(
				t, filesystem.GetRootDir("../../"), app.DataDir())
			defer tearDownMigration(t)

			adminCreate := services.NewAdminCreate(app)
			admin, err := adminCreate.Create(scenario.req)
			if err != nil && err.Error() != scenario.wantErr.Error() {
				t.Errorf("got error %v, want %v", err, scenario.wantErr)
			}
			if admin != nil {
				if admin.Id == uuid.Nil {
					t.Errorf("got admin.Id nil, want an actual UUID")
				}
				if admin.Email != scenario.req.Email {
					t.Errorf("got admin.Email %s, want %s", admin.Email, scenario.req.Email)
				}
				if admin.FirstName != scenario.req.FirstName {
					t.Errorf("got admin.FirstName %s, want %s", admin.FirstName, scenario.req.FirstName)
				}
				if admin.LastName != scenario.req.LastName {
					t.Errorf("got admin.LastName %s, want %s", admin.LastName, scenario.req.LastName)
				}
			}
		})
	}
}
