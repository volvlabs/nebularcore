package backends_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/jideobs/nebularcore/modules/auth/backends"
	backendMocks "gitlab.com/jideobs/nebularcore/modules/auth/backends/mocks"
	autherrors "gitlab.com/jideobs/nebularcore/modules/auth/errors"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces/mocks"
)

func TestAuthenticationManager(t *testing.T) {
	t.Run("RegisterBackend", func(t *testing.T) {
		t.Run("successful registration", func(t *testing.T) {
			manager := backends.NewAuthenticationManager()
			backend := backendMocks.NewAuthenticationBackend(t)

			assert.NotPanics(t, func() {
				manager.RegisterBackend("test", backend)
			})
		})

		t.Run("duplicate registration", func(t *testing.T) {
			manager := backends.NewAuthenticationManager()
			backend := backendMocks.NewAuthenticationBackend(t)

			manager.RegisterBackend("test", backend)
			assert.Panics(t, func() {
				manager.RegisterBackend("test", backend)
			})
		})
	})

	t.Run("Authenticate", func(t *testing.T) {
		tests := []struct {
			name        string
			setup       func() (backends.AuthenticationManager, interfaces.User)
			credentials map[string]any
			expectError string
		}{
			{
				name: "successful authentication with first backend",
				setup: func() (backends.AuthenticationManager, interfaces.User) {
					manager := backends.NewAuthenticationManager()
					backend1 := backendMocks.NewAuthenticationBackend(t)
					backend2 := backendMocks.NewAuthenticationBackend(t)
					user := mocks.NewUser(t)

					creds := map[string]any{"username": "testuser", "password": "password123"}
					backend1.On("Authenticate", mock.Anything, creds).Return(user, nil)

					manager.RegisterBackend("backend1", backend1)
					manager.RegisterBackend("backend2", backend2)

					return manager, user
				},
				credentials: map[string]any{"username": "testuser", "password": "password123"},
			},
			{
				name: "successful authentication with second backend",
				setup: func() (backends.AuthenticationManager, interfaces.User) {
					manager := backends.NewAuthenticationManager()
					backend1 := backendMocks.NewAuthenticationBackend(t)
					backend2 := backendMocks.NewAuthenticationBackend(t)
					user := mocks.NewUser(t)

					creds := map[string]any{"username": "testuser", "password": "password123"}
					backend1.On("Authenticate", mock.Anything, creds).Return(nil, autherrors.ErrInvalidCredentials)
					backend2.On("Authenticate", mock.Anything, creds).Return(user, nil)

					manager.RegisterBackend("backend1", backend1)
					manager.RegisterBackend("backend2", backend2)

					return manager, user
				},
				credentials: map[string]any{"username": "testuser", "password": "password123"},
			},
			{
				name: "no backends registered",
				setup: func() (backends.AuthenticationManager, interfaces.User) {
					manager := backends.NewAuthenticationManager()
					return manager, nil
				},
				credentials: map[string]any{"username": "testuser", "password": "password123"},
				expectError: "no authentication backends registered",
			},
			{
				name: "all backends fail with invalid credentials",
				setup: func() (backends.AuthenticationManager, interfaces.User) {
					manager := backends.NewAuthenticationManager()
					backend1 := backendMocks.NewAuthenticationBackend(t)
					backend2 := backendMocks.NewAuthenticationBackend(t)

					creds := map[string]any{"username": "testuser", "password": "password123"}
					backend1.On("Authenticate", mock.Anything, creds).Return(nil, autherrors.ErrInvalidCredentials)
					backend2.On("Authenticate", mock.Anything, creds).Return(nil, autherrors.ErrInvalidCredentials)

					manager.RegisterBackend("backend1", backend1)
					manager.RegisterBackend("backend2", backend2)

					return manager, nil
				},
				credentials: map[string]any{"username": "testuser", "password": "password123"},
				expectError: "invalid credentials",
			},
			{
				name: "backend returns non-credential error",
				setup: func() (backends.AuthenticationManager, interfaces.User) {
					manager := backends.NewAuthenticationManager()
					backend := backendMocks.NewAuthenticationBackend(t)

					creds := map[string]any{"username": "testuser", "password": "password123"}
					backend.On("Authenticate", mock.Anything, creds).Return(nil, autherrors.ErrUserDisabled)

					manager.RegisterBackend("backend1", backend)

					return manager, nil
				},
				credentials: map[string]any{"username": "testuser", "password": "password123"},
				expectError: "user_disabled",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				manager, expectedUser := tt.setup()

				user, err := manager.Authenticate(context.Background(), tt.credentials)

				if tt.expectError != "" {
					assert.EqualError(t, err, tt.expectError)
					assert.Nil(t, user)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, expectedUser, user)
				}
			})
		}
	})

	t.Run("ValidateToken", func(t *testing.T) {
		tests := []struct {
			name        string
			setup       func() (backends.AuthenticationManager, interfaces.User)
			token       string
			expectError string
		}{
			{
				name: "successful validation with first backend",
				setup: func() (backends.AuthenticationManager, interfaces.User) {
					manager := backends.NewAuthenticationManager()
					backend1 := backendMocks.NewAuthenticationBackend(t)
					backend2 := backendMocks.NewAuthenticationBackend(t)
					user := mocks.NewUser(t)

					backend1.On("ValidateToken", mock.Anything, "valid-token").Return(user, nil)

					manager.RegisterBackend("backend1", backend1)
					manager.RegisterBackend("backend2", backend2)

					return manager, user
				},
				token: "valid-token",
			},
			{
				name: "successful validation with second backend",
				setup: func() (backends.AuthenticationManager, interfaces.User) {
					manager := backends.NewAuthenticationManager()
					backend1 := backendMocks.NewAuthenticationBackend(t)
					backend2 := backendMocks.NewAuthenticationBackend(t)
					user := mocks.NewUser(t)

					backend1.On("ValidateToken", mock.Anything, "valid-token").Return(nil, autherrors.ErrInvalidToken)
					backend2.On("ValidateToken", mock.Anything, "valid-token").Return(user, nil)

					manager.RegisterBackend("backend1", backend1)
					manager.RegisterBackend("backend2", backend2)

					return manager, user
				},
				token: "valid-token",
			},
			{
				name: "no backends registered",
				setup: func() (backends.AuthenticationManager, interfaces.User) {
					manager := backends.NewAuthenticationManager()
					return manager, nil
				},
				token:       "valid-token",
				expectError: "no authentication backends registered",
			},
			{
				name: "all backends fail with invalid token",
				setup: func() (backends.AuthenticationManager, interfaces.User) {
					manager := backends.NewAuthenticationManager()
					backend1 := backendMocks.NewAuthenticationBackend(t)
					backend2 := backendMocks.NewAuthenticationBackend(t)

					backend1.On("ValidateToken", mock.Anything, "invalid-token").Return(nil, autherrors.ErrInvalidToken)
					backend2.On("ValidateToken", mock.Anything, "invalid-token").Return(nil, autherrors.ErrInvalidToken)

					manager.RegisterBackend("backend1", backend1)
					manager.RegisterBackend("backend2", backend2)

					return manager, nil
				},
				token:       "invalid-token",
				expectError: "invalid token",
			},
			{
				name: "backend returns non-token error",
				setup: func() (backends.AuthenticationManager, interfaces.User) {
					manager := backends.NewAuthenticationManager()
					backend := backendMocks.NewAuthenticationManager(t)

					backend.On("ValidateToken", mock.Anything, "token-error").Return(nil, autherrors.ErrUserDisabled)

					manager.RegisterBackend("backend1", backend)

					return manager, nil
				},
				token:       "token-error",
				expectError: "user_disabled",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				manager, expectedUser := tt.setup()

				user, err := manager.ValidateToken(context.Background(), tt.token)

				if tt.expectError != "" {
					assert.EqualError(t, err, tt.expectError)
					assert.Nil(t, user)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, expectedUser, user)
				}
			})
		}
	})
}
