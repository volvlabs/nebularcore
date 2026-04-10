package authentication_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volvlabs/nebularcore/core"
	"github.com/volvlabs/nebularcore/entities"
	"github.com/volvlabs/nebularcore/models/requests"
	"github.com/volvlabs/nebularcore/services/authentication"
	"github.com/volvlabs/nebularcore/test"
	"github.com/volvlabs/nebularcore/tools/filesystem"
	"github.com/volvlabs/nebularcore/tools/security"
	"github.com/volvlabs/nebularcore/tools/types"
)

func TestInitiateResetPassword(t *testing.T) {
	type args struct {
		payload *requests.InitiateResetPasswordPayload
	}
	scenarios := []struct {
		name          string
		args          args
		beforeFunc    func(*testing.T, core.App)
		afterFunc     func(*testing.T, core.App)
		expectedError error
	}{
		{
			name: "should initiate password reset successfully",
			args: args{
				payload: &requests.InitiateResetPasswordPayload{
					Email: "john.doe@gmail.com",
				},
			},
			beforeFunc: func(t *testing.T, testApp core.App) {
				err := testApp.Dao().Save(&entities.Auth{
					Identity: "john.doe@gmail.com",
				})
				require.NoError(t, err)
			},
			afterFunc: func(t *testing.T, testApp core.App) {
				auth := &entities.Auth{}
				err := testApp.Dao().DB().Model(&entities.Auth{}).
					Where("identity = ?", "john.doe@gmail.com").
					First(auth).Error
				require.NoError(t, err)

				assert.NotEqual(t, "", auth.ResetPasswordToken)
				assert.NotEqual(t, nil, auth.ResetPasswordTokenExpiryDate)
				assert.True(t, time.Now().Before(auth.ResetPasswordTokenExpiryDate.Time()))
			},
			expectedError: nil,
		},
		{
			name: "should return no error even for an inexistent user",
			args: args{
				payload: &requests.InitiateResetPasswordPayload{
					Email: "john.doe@outlook.com",
				},
			},
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
			if scenario.beforeFunc != nil {
				scenario.beforeFunc(t, testApp)
			}

			authService := authentication.New(testApp)

			// Act:
			err := authService.InitiateResetPassword(scenario.args.payload)

			// Assert:
			assert.Equal(t, scenario.expectedError, err)
			if scenario.afterFunc != nil {
				scenario.afterFunc(t, testApp)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	type args struct {
		payload *requests.ValidateRequestPasswordTokenPayload
	}
	scenarios := []struct {
		name        string
		args        args
		beforeFunc  func(*testing.T, core.App)
		expectedErr error
	}{
		{
			name: "should return nil since the token is valid",
			args: args{
				payload: &requests.ValidateRequestPasswordTokenPayload{
					Token: "xxxxxxxxx",
				},
			},
			beforeFunc: func(t *testing.T, app core.App) {
				expiryDateTime, err := types.ParseDateTime(time.Now().Add(1 * time.Hour))
				require.NoError(t, err)

				err = app.Dao().DB().Save(&entities.Auth{
					Identity:                     "john.doe@gmail.com",
					ResetPasswordToken:           "xxxxxxxxx",
					ResetPasswordTokenExpiryDate: expiryDateTime,
				}).Error
				require.NoError(t, err)
			},
		},
		{
			name: "should return error because user with reset token not found",
			args: args{
				payload: &requests.ValidateRequestPasswordTokenPayload{
					Token: "xxxxxx",
				},
			},
			expectedErr: authentication.ErrInvalidPasswordToken,
		},
		{
			name: "should return error because reset token already expired",
			args: args{
				payload: &requests.ValidateRequestPasswordTokenPayload{
					Token: "xxxxxxxxx",
				},
			},
			beforeFunc: func(t *testing.T, app core.App) {
				expiryDateTime, err := types.ParseDateTime(time.Now().Add(-1 * time.Hour))
				require.NoError(t, err)

				err = app.Dao().DB().Save(&entities.Auth{
					Identity:                     "john.doe@gmail.com",
					ResetPasswordToken:           "xxxxxxxxx",
					ResetPasswordTokenExpiryDate: expiryDateTime,
				}).Error
				require.NoError(t, err)
			},
			expectedErr: authentication.ErrTokenExpired,
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Arrange:
			testApp, _ := test.NewTestApp()
			tearDownMigration := test.RunMigration(
				t, filesystem.GetRootDir("../../"), testApp.DataDir())
			defer tearDownMigration(t)
			if scenario.beforeFunc != nil {
				scenario.beforeFunc(t, testApp)
			}

			authService := authentication.New(testApp)

			// Act:
			err := authService.ValidateToken(scenario.args.payload)

			// Assert:
			assert.Equal(t, scenario.expectedErr, err)
		})
	}
}

func TestResetPassword(t *testing.T) {
	type args struct {
		payload *requests.ResetPasswordPayload
	}
	scenarios := []struct {
		name        string
		args        args
		beforeFunc  func(*testing.T, core.App)
		afterFunc   func(*testing.T, core.App)
		expectedErr error
	}{
		{
			name: "should successfully reset the password",
			args: args{
				payload: &requests.ResetPasswordPayload{
					Token:    "xxxxxxxxx",
					Password: "newpassword",
				},
			},
			beforeFunc: func(t *testing.T, app core.App) {
				expiryDateTime, err := types.ParseDateTime(time.Now().Add(1 * time.Hour))
				require.NoError(t, err)

				err = app.Dao().DB().Save(&entities.Auth{
					Identity:                     "john.doe@gmail.com",
					ResetPasswordToken:           "xxxxxxxxx",
					ResetPasswordTokenExpiryDate: expiryDateTime,
				}).Error
				require.NoError(t, err)
			},
			afterFunc: func(t *testing.T, app core.App) {
				auth := &entities.Auth{}
				err := app.Dao().DB().Where("identity = ?", "john.doe@gmail.com").
					First(auth).Error
				require.NoError(t, err)

				assert.True(t, security.ValidatePassword(auth.PasswordHash, "newpassword"))
			},
		},
		{
			name: "should return error because the token is invalid",
			args: args{
				payload: &requests.ResetPasswordPayload{
					Token:    "xxxxxxxxx",
					Password: "newpassword",
				},
			},
			expectedErr: authentication.ErrInvalidPasswordToken,
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Arrange:
			testApp, _ := test.NewTestApp()
			tearDownMigration := test.RunMigration(
				t, filesystem.GetRootDir("../../"), testApp.DataDir())
			defer tearDownMigration(t)
			if scenario.beforeFunc != nil {
				scenario.beforeFunc(t, testApp)
			}

			authService := authentication.New(testApp)

			// Act:
			err := authService.ResetPassword(scenario.args.payload)

			// Assert:
			assert.Equal(t, scenario.expectedErr, err)
			if scenario.afterFunc != nil {
				scenario.afterFunc(t, testApp)
			}
		})
	}
}
