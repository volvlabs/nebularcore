package authentication_test

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jideobs/nebularcore/entities"
	"gitlab.com/jideobs/nebularcore/models/requests"
	"gitlab.com/jideobs/nebularcore/services/authentication"
	"gitlab.com/jideobs/nebularcore/test"
	"gitlab.com/jideobs/nebularcore/tools/filesystem"
	"gitlab.com/jideobs/nebularcore/tools/types"
	"testing"
)

func TestResetPassword(t *testing.T) {
	scenarios := []struct {
		name                 string
		resetPasswordRequest requests.ResetPasswordRequest
		createUser           bool
		expectedError        error
	}{
		{
			name: "should start password reset successfully",
			resetPasswordRequest: requests.ResetPasswordRequest{
				Email: "john.doe@gmail.com",
			},
			createUser:    true,
			expectedError: nil,
		},
		{
			name:                 "should return error because of empty request data",
			resetPasswordRequest: requests.ResetPasswordRequest{},
			createUser:           true,
			expectedError: &types.RequestBodyError{
				Message: "error validating request body",
				Errors: []types.FieldError{
					{
						Field:   "Email",
						Message: "Email is a required field",
					},
				},
			},
		},
		{
			name: "should return no error even when the user record was not found",
			resetPasswordRequest: requests.ResetPasswordRequest{
				Email: "john.doe@gmail.com",
			},
			createUser:    false,
			expectedError: nil,
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Arrange:
			testApp, _ := test.NewTestApp()
			tearDownMigration := test.RunMigration(
				t, filesystem.GetRootDir("../../"), testApp.DataDir())
			defer tearDownMigration(t)

			authService := authentication.New(testApp)

			err := testApp.Dao().Save(&entities.Auth{
				Identity: scenario.resetPasswordRequest.Email})
			if err != nil {
				t.Fatal(err)
			}

			// Act:
			err = authService.ResetPassword(scenario.resetPasswordRequest)

			// Assert:
			assert.Equal(t, scenario.expectedError, err)
		})
	}
}
