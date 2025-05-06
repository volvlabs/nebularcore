package backends_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/jideobs/nebularcore/modules/auth/backends"
	autherrors "gitlab.com/jideobs/nebularcore/modules/auth/errors"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces/mocks"
)

func TestLocalBackend(t *testing.T) {
	t.Run("Name and Priority", func(t *testing.T) {
		backend := backends.NewLocalBackend(nil, nil)
		assert.Equal(t, "local", backend.Name())
		assert.Equal(t, 1, backend.Priority())
	})

	t.Run("Supports", func(t *testing.T) {
		backend := backends.NewLocalBackend(nil, nil)
		assert.True(t, backend.Supports("username_password"))
		assert.True(t, backend.Supports("email_password"))
		assert.True(t, backend.Supports("phone_password"))
		assert.False(t, backend.Supports("oauth"))
	})

	t.Run("Authenticate", func(t *testing.T) {
		tests := []struct {
			name        string
			credentials map[string]any
			setup       func(*mocks.UserRepository) interfaces.User
			expectError error
		}{
			{
				name: "successful authentication with username",
				credentials: map[string]any{
					"username": "testuser",
					"password": "password123",
				},
				setup: func(repo *mocks.UserRepository) interfaces.User {
					user := mocks.NewUser(t)
					user.On("GetPasswordHash").Return("$2y$10$H514ifu2MkoHdEbnPM4Z1.rNiJ02CcZtSB2CW9dFVNz1fJX/sfQ3C") // bcrypt hash for "password123"
					user.On("IsActive").Return(true)
					repo.On("FindByUsername", mock.Anything, "testuser").Return(user, nil)
					return user
				},
			},
			{
				name: "successful authentication with email",
				credentials: map[string]any{
					"email":    "test@example.com",
					"password": "password123",
				},
				setup: func(repo *mocks.UserRepository) interfaces.User {
					user := mocks.NewUser(t)
					user.On("GetPasswordHash").Return("$2y$10$H514ifu2MkoHdEbnPM4Z1.rNiJ02CcZtSB2CW9dFVNz1fJX/sfQ3C")
					user.On("IsActive").Return(true)
					repo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)
					return user
				},
			},
			{
				name: "successful authentication with phone number",
				credentials: map[string]any{
					"phoneNumber": "+1234567890",
					"password":    "password123",
				},
				setup: func(repo *mocks.UserRepository) interfaces.User {
					user := mocks.NewUser(t)
					user.On("GetPasswordHash").Return("$2y$10$H514ifu2MkoHdEbnPM4Z1.rNiJ02CcZtSB2CW9dFVNz1fJX/sfQ3C")
					user.On("IsActive").Return(true)
					repo.On("FindByPhoneNumber", mock.Anything, "+1234567890").Return(user, nil)
					return user
				},
			},
			{
				name: "invalid credentials - wrong password",
				credentials: map[string]any{
					"username": "testuser",
					"password": "wrongpassword",
				},
				setup: func(repo *mocks.UserRepository) interfaces.User {
					user := mocks.NewUser(t)
					user.On("GetPasswordHash").Return("$2y$10$H514ifu2MkoHdEbnPM4Z1.rNiJ02CcZtSB2CW9dFVNz1fJX/sfQ3C")
					user.On("IsActive").Return(true)
					repo.On("FindByUsername", mock.Anything, "testuser").Return(user, nil)
					return user
				},
				expectError: autherrors.ErrInvalidCredentials,
			},
			{
				name: "user not found",
				credentials: map[string]any{
					"username": "nonexistent",
					"password": "password123",
				},
				setup: func(repo *mocks.UserRepository) interfaces.User {
					repo.On("FindByUsername", mock.Anything, "nonexistent").Return(nil, autherrors.ErrUserNotFound)
					return nil
				},
				expectError: autherrors.ErrInvalidCredentials,
			},
			{
				name: "user disabled",
				credentials: map[string]any{
					"username": "disabled",
					"password": "password123",
				},
				setup: func(repo *mocks.UserRepository) interfaces.User {
					user := mocks.NewUser(t)
					user.On("IsActive").Return(false)
					repo.On("FindByUsername", mock.Anything, "disabled").Return(user, nil)
					return user
				},
				expectError: autherrors.ErrUserDisabled,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo := mocks.NewUserRepository(t)
				user := tt.setup(repo)
				backend := backends.NewLocalBackend(repo, nil)

				authUser, err := backend.Authenticate(context.Background(), tt.credentials)

				if tt.expectError != nil {
					assert.Equal(t, tt.expectError, err)
					assert.Nil(t, authUser)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, user, authUser)
				}

				repo.AssertExpectations(t)
				if user != nil {
					user.(*mocks.User).AssertExpectations(t)
				}
			})
		}
	})

	t.Run("ValidateToken", func(t *testing.T) {
		tests := []struct {
			name        string
			token       string
			setup       func(*mocks.TokenIssuer, *mocks.UserRepository) interfaces.User
			expectError string
		}{
			{
				name:  "valid token",
				token: "valid-token",
				setup: func(issuer *mocks.TokenIssuer, repo *mocks.UserRepository) interfaces.User {
					userID := uuid.New()
					claims := map[string]any{"sub": userID.String()}
					user := mocks.NewUser(t)
					user.On("IsActive").Return(true)

					issuer.On("ValidateToken", "valid-token").Return(claims, nil)
					repo.On("FindByID", mock.Anything, userID).Return(user, nil)

					return user
				},
			},
			{
				name:        "invalid token",
				token:       "invalid-token",
				setup: func(issuer *mocks.TokenIssuer, repo *mocks.UserRepository) interfaces.User {
					issuer.On("ValidateToken", "invalid-token").Return(nil, autherrors.ErrInvalidToken)
					return nil
				},
				expectError: "invalid or expired token",
			},
			{
				name:  "invalid token claims",
				token: "token-invalid-claims",
				setup: func(issuer *mocks.TokenIssuer, repo *mocks.UserRepository) interfaces.User {
					issuer.On("ValidateToken", "token-invalid-claims").Return(map[string]any{"foo": "bar"}, nil)
					return nil
				},
				expectError: "invalid token claims",
			},
			{
				name:  "user not found",
				token: "token-user-not-found",
				setup: func(issuer *mocks.TokenIssuer, repo *mocks.UserRepository) interfaces.User {
					userID := uuid.New()
					claims := map[string]any{"sub": userID.String()}
					issuer.On("ValidateToken", "token-user-not-found").Return(claims, nil)
					repo.On("FindByID", mock.Anything, userID).Return(nil, autherrors.ErrUserNotFound)
					return nil
				},
				expectError: "user not found",
			},
			{
				name:  "user disabled",
				token: "token-user-disabled",
				setup: func(issuer *mocks.TokenIssuer, repo *mocks.UserRepository) interfaces.User {
					userID := uuid.New()
					claims := map[string]any{"sub": userID.String()}
					user := mocks.NewUser(t)
					user.On("IsActive").Return(false)

					issuer.On("ValidateToken", "token-user-disabled").Return(claims, nil)
					repo.On("FindByID", mock.Anything, userID).Return(user, nil)

					return user
				},
				expectError: "user account is disabled",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				issuer := mocks.NewTokenIssuer(t)
				repo := mocks.NewUserRepository(t)
				expectedUser := tt.setup(issuer, repo)

				backend := backends.NewLocalBackend(repo, issuer)
				user, err := backend.ValidateToken(context.Background(), tt.token)

				if tt.expectError != "" {
					assert.EqualError(t, err, tt.expectError)
					assert.Nil(t, user)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, expectedUser, user)
				}

				issuer.AssertExpectations(t)
				repo.AssertExpectations(t)
				if expectedUser != nil {
					expectedUser.(*mocks.User).AssertExpectations(t)
				}
			})
		}
	})
}
