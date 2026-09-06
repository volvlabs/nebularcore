package backends_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/volvlabs/nebularcore/modules/auth/backends"
	autherrors "github.com/volvlabs/nebularcore/modules/auth/errors"
	"github.com/volvlabs/nebularcore/modules/auth/interfaces"
	"github.com/volvlabs/nebularcore/modules/auth/interfaces/mocks"
	"github.com/volvlabs/nebularcore/modules/auth/models"
	"github.com/volvlabs/nebularcore/modules/auth/pkg"
	"github.com/volvlabs/nebularcore/modules/auth/types"
)

func TestGoogleBackend(t *testing.T) {
	t.Run("Name and Priority", func(t *testing.T) {
		backend := backends.NewGoogleBackend(nil, nil, nil, nil)
		assert.Equal(t, "google", backend.Name())
		assert.Equal(t, 2, backend.Priority())
	})

	t.Run("Supports", func(t *testing.T) {
		backend := backends.NewGoogleBackend(nil, nil, nil, nil)
		assert.True(t, backend.Supports("google"))
		assert.False(t, backend.Supports("local"))
	})

	t.Run("Authenticate", func(t *testing.T) {
		scenarios := []struct {
			name          string
			code          string
			idToken       string
			setup         func(*mocks.SocialAccountRepository, *mocks.GoogleSignin) interfaces.User
			expectedError error
		}{
			{
				name: "successful authentication with google signin",
				code: "authcode",
				setup: func(repo *mocks.SocialAccountRepository, signin *mocks.GoogleSignin) interfaces.User {
					user := models.User{
						Active: true,
					}
					socialAcount := &models.SocialAccount{
						User: user,
					}
					signin.On("Exchange", mock.Anything, "authcode").Return(&pkg.GoogleUser{
						ID: "googleid",
					}, nil)
					repo.On("FindByProvider", mock.Anything, types.AuthProviderGoogle, "googleid").
						Return(socialAcount, nil)
					return &user
				},
			},
			{
				name: "failed to exchange code for google user",
				code: "authcode",
				setup: func(repo *mocks.SocialAccountRepository, signin *mocks.GoogleSignin) interfaces.User {
					signin.On("Exchange", mock.Anything, "authcode").Return(nil, autherrors.ErrInvalidToken)
					return nil
				},
				expectedError: autherrors.ErrInvalidToken,
			},
			{
				name: "failed to find user by provider",
				code: "authcode",
				setup: func(repo *mocks.SocialAccountRepository, signin *mocks.GoogleSignin) interfaces.User {
					signin.On("Exchange", mock.Anything, "authcode").Return(&pkg.GoogleUser{
						ID: "googleid",
					}, nil)
					repo.On("FindByProvider", mock.Anything, types.AuthProviderGoogle, "googleid").Return(nil, autherrors.ErrUserNotFound)
					return nil
				},
				expectedError: autherrors.ErrUserNotFound,
			},
			{
				name:    "successful authentication with id token",
				idToken: "idtoken",
				setup: func(repo *mocks.SocialAccountRepository, signin *mocks.GoogleSignin) interfaces.User {
					user := models.User{
						Active: true,
					}
					socialAcount := &models.SocialAccount{
						User: user,
					}
					signin.On("VerifyGoogleIDToken", mock.Anything, "idtoken").Return(&pkg.GoogleUser{
						ID: "googleid",
					}, nil)
					repo.On("FindByProvider", mock.Anything, types.AuthProviderGoogle, "googleid").Return(socialAcount, nil)
					return &user
				},
			},
			{
				name:    "failed to verify id token",
				idToken: "idtoken",
				setup: func(repo *mocks.SocialAccountRepository, signin *mocks.GoogleSignin) interfaces.User {
					signin.On("VerifyGoogleIDToken", mock.Anything, "idtoken").Return(nil, autherrors.ErrInvalidToken)
					return nil
				},
				expectedError: autherrors.ErrInvalidToken,
			},
			{
				name: "missing credentials (no code or idToken provided)",
				setup: func(repo *mocks.SocialAccountRepository, signin *mocks.GoogleSignin) interfaces.User {
					// no expectations should be set on repo/signin because Authenticate should return early
					return nil
				},
				expectedError: autherrors.ErrInvalidCredentials,
			},
		}

		for _, tt := range scenarios {
			t.Run(tt.name, func(t *testing.T) {
				repo := mocks.NewSocialAccountRepository(t)
				userRepo := mocks.NewUserRepository(t)
				signin := mocks.NewGoogleSignin(t)
				backend := backends.NewGoogleBackend(repo, userRepo, nil, signin)
				user := tt.setup(repo, signin)

				authUser, err := backend.Authenticate(
					context.Background(), map[string]any{"code": tt.code, "idToken": tt.idToken})

				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError, err)
					assert.Nil(t, authUser)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, user, authUser)
				}

				repo.AssertExpectations(t)
				signin.AssertExpectations(t)
			})
		}
	})
}
