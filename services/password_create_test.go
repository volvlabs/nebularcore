package services_test

import (
	"testing"

	"gitlab.com/jideobs/nebularcore/services"
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
			name:         "should successfully create auth information",
			identity:     "john.doe@gmail.com",
			passwordHash: "password",
			wantErr:      nil,
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			tearDownMigration := test.RunMigration(
				t, filesystem.GetRootDir("../"), app.DataDir())
			defer tearDownMigration(t)

			auth := services.NewAuth(app)
			err := auth.Create(scenario.identity, scenario.passwordHash)
			if err != nil && err != scenario.wantErr {
				t.Errorf("got %v, want %v", err, scenario.wantErr)
			}
		})
	}
}
