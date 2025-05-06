package password

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/jideobs/nebularcore/modules/auth/emitter/mocks"
	authMocks "gitlab.com/jideobs/nebularcore/modules/auth/interfaces/mocks"
	"gitlab.com/jideobs/nebularcore/modules/auth/models/requests"
	"gitlab.com/jideobs/nebularcore/tools/types"
	"golang.org/x/crypto/bcrypt"
)

func TestManager_RequestPasswordReset(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(eventEmitter *mocks.EventEmitter, userRepo *authMocks.UserRepository)
		payload   requests.PasswordResetPayload
		wantError bool
	}{
		{
			name: "successful password reset request",
			setup: func(eventEmitter *mocks.EventEmitter, userRepo *authMocks.UserRepository) {
				mockUser := authMocks.NewUser(t)
				mockUser.On("SetPasswordResetToken", mock.AnythingOfType("*string")).Return(nil)
				mockUser.On("SetPasswordResetAt", mock.AnythingOfType("*time.Time")).Return(nil)

				userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(mockUser, nil)
				userRepo.On("Update", mock.Anything, mockUser, mock.MatchedBy(func(updates map[string]interface{}) bool {
					_, hasToken := updates["password_reset_token"]
					_, hasTime := updates["password_reset_at"]
					return hasToken && hasTime
				})).Return(nil)

				eventEmitter.On("EmitPasswordResetInitiatedEvent", mock.Anything, mockUser).Return(nil)
			},
			payload: requests.PasswordResetPayload{
				Email: "test@example.com",
			},
			wantError: false,
		},
		{
			name: "user not found - should not return error",
			setup: func(eventEmitter *mocks.EventEmitter, userRepo *authMocks.UserRepository) {
				userRepo.On("FindByEmail", mock.Anything, "nonexistent@example.com").Return(nil, nil)
			},
			payload: requests.PasswordResetPayload{
				Email: "nonexistent@example.com",
			},
			wantError: false,
		},
		{
			name: "database error finding user",
			setup: func(eventEmitter *mocks.EventEmitter, userRepo *authMocks.UserRepository) {
				userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, assert.AnError)
			},
			payload: requests.PasswordResetPayload{
				Email: "test@example.com",
			},
			wantError: false,
		},
		{
			name: "error updating user",
			setup: func(eventEmitter *mocks.EventEmitter, userRepo *authMocks.UserRepository) {
				mockUser := authMocks.NewUser(t)
				mockUser.On("SetPasswordResetToken", mock.AnythingOfType("*string")).Return(nil)
				mockUser.On("SetPasswordResetAt", mock.AnythingOfType("*time.Time")).Return(nil)

				userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(mockUser, nil)
				userRepo.On("Update", mock.Anything, mockUser, mock.AnythingOfType("map[string]interface {}")).Return(assert.AnError)
			},
			payload: requests.PasswordResetPayload{
				Email: "test@example.com",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventEmitter := mocks.NewEventEmitter(t)
			userRepo := authMocks.NewUserRepository(t)

			if tt.setup != nil {
				tt.setup(eventEmitter, userRepo)
			}

			manager := NewManager(eventEmitter, userRepo)
			err := manager.RequestPasswordReset(context.Background(), tt.payload)

			if tt.wantError {
				assert.Error(t, err)
				appErr, ok := err.(*types.AppError)
				assert.True(t, ok)
				assert.Equal(t, types.ErrorTypeSystem, appErr.Type)
			} else {
				assert.NoError(t, err)
			}

			eventEmitter.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}

func TestManager_VerifyPasswordReset(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(eventEmitter *mocks.EventEmitter, userRepo *authMocks.UserRepository)
		payload   requests.PasswordResetVerifyPayload
		wantError bool
		errorType types.ErrorType
	}{
		{
			name: "successful password reset verification",
			setup: func(eventEmitter *mocks.EventEmitter, userRepo *authMocks.UserRepository) {
				mockUser := authMocks.NewUser(t)
				resetTime := time.Now().Add(-1 * time.Hour)
				mockUser.On("GetPasswordResetAt").Return(&resetTime)
				mockUser.On("SetPassword", mock.AnythingOfType("string")).Return(nil)
				mockUser.On("SetPasswordResetToken", mock.AnythingOfType("*string")).Return(nil)
				mockUser.On("SetPasswordResetAt", mock.AnythingOfType("*time.Time")).Return(nil)
				resetTime = time.Now()
				mockUser.On("GetPasswordResetAt").Return(&resetTime)

				userRepo.On("FindByResetToken", mock.Anything, "valid_token").Return(mockUser, nil)
				userRepo.On("Update", mock.Anything, mockUser, mock.MatchedBy(func(updates map[string]interface{}) bool {
					_, hasPassword := updates["password"]
					_, hasToken := updates["password_reset_token"]
					_, hasTime := updates["password_reset_at"]
					return hasPassword && hasToken && hasTime
				})).Return(nil)

				eventEmitter.On("EmitPasswordChangedEvent", mock.Anything, mockUser).Return(nil)
			},
			payload: requests.PasswordResetVerifyPayload{
				Token:    "valid_token",
				Password: "new_password123",
			},
			wantError: false,
		},
		{
			name: "invalid reset token",
			setup: func(eventEmitter *mocks.EventEmitter, userRepo *authMocks.UserRepository) {
				userRepo.On("FindByResetToken", mock.Anything, "invalid_token").Return(nil, nil)
			},
			payload: requests.PasswordResetVerifyPayload{
				Token:    "invalid_token",
				Password: "new_password123",
			},
			wantError: true,
			errorType: types.ErrorTypeUser,
		},
		{
			name: "expired reset token",
			setup: func(eventEmitter *mocks.EventEmitter, userRepo *authMocks.UserRepository) {
				mockUser := authMocks.NewUser(t)
				resetTime := time.Now().Add(-25 * time.Hour) // Reset token created 25 hours ago
				mockUser.On("GetPasswordResetAt").Return(&resetTime)

				userRepo.On("FindByResetToken", mock.Anything, "expired_token").Return(mockUser, nil)
			},
			payload: requests.PasswordResetVerifyPayload{
				Token:    "expired_token",
				Password: "new_password123",
			},
			wantError: true,
			errorType: types.ErrorTypeUser,
		},
		{
			name: "database error finding user",
			setup: func(eventEmitter *mocks.EventEmitter, userRepo *authMocks.UserRepository) {
				userRepo.On("FindByResetToken", mock.Anything, "valid_token").Return(nil, assert.AnError)
			},
			payload: requests.PasswordResetVerifyPayload{
				Token:    "valid_token",
				Password: "new_password123",
			},
			wantError: true,
			errorType: types.ErrorTypeUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventEmitter := mocks.NewEventEmitter(t)
			userRepo := authMocks.NewUserRepository(t)

			if tt.setup != nil {
				tt.setup(eventEmitter, userRepo)
			}

			manager := NewManager(eventEmitter, userRepo)
			err := manager.VerifyPasswordReset(context.Background(), tt.payload)

			if tt.wantError {
				assert.Error(t, err)
				appErr, ok := err.(*types.AppError)
				assert.True(t, ok)
				if tt.errorType != types.ErrorType(0) {
					assert.Equal(t, tt.errorType, appErr.Type)
				}
			} else {
				assert.NoError(t, err)
			}

			eventEmitter.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}

func TestManager_ChangePassword(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(eventEmitter *mocks.EventEmitter, userRepo *authMocks.UserRepository, user *authMocks.User)
		payload   requests.PasswordChangePayload
		wantError bool
		errorType types.ErrorType
	}{
		{
			name: "successful password change",
			setup: func(eventEmitter *mocks.EventEmitter, userRepo *authMocks.UserRepository, user *authMocks.User) {
				mockUser := authMocks.NewUser(t)
				userID := uuid.New()
				user.On("GetID").Return(userID)

				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("current_password"), bcrypt.DefaultCost)
				mockUser.On("GetPasswordHash").Return(string(hashedPassword))
				mockUser.On("SetPassword", mock.AnythingOfType("string")).Return(nil)

				userRepo.On("FindByID", mock.Anything, userID).Return(mockUser, nil)
				userRepo.On("Update", mock.Anything, mockUser, mock.MatchedBy(func(updates map[string]any) bool {
					_, hasPassword := updates["password"]
					return hasPassword
				})).Return(nil)

				eventEmitter.On("EmitPasswordChangedEvent", mock.Anything, mockUser).Return(nil)
			},
			payload: requests.PasswordChangePayload{
				CurrentPassword: "current_password",
				NewPassword:     "new_password123",
			},
			wantError: false,
		},
		{
			name: "incorrect current password",
			setup: func(eventEmitter *mocks.EventEmitter, userRepo *authMocks.UserRepository, user *authMocks.User) {
				mockUser := authMocks.NewUser(t)
				mockUser.On("GetPasswordHash").Return("$2a$10$abcdefghijklmnopqrstuvwxyz012345") // bcrypt hash for "different_password"

				userID := uuid.New()
				user.On("GetID").Return(userID)

				userRepo.On("FindByID", mock.Anything, userID).Return(mockUser, nil)
			},
			payload: requests.PasswordChangePayload{
				CurrentPassword: "wrong_password",
				NewPassword:     "new_password123",
			},
			wantError: true,
			errorType: types.ErrorTypeUser,
		},
		{
			name: "database error finding user",
			setup: func(eventEmitter *mocks.EventEmitter, userRepo *authMocks.UserRepository, user *authMocks.User) {
				userID := uuid.New()
				user.On("GetID").Return(userID)

				userRepo.On("FindByID", mock.Anything, userID).Return(nil, assert.AnError)
			},
			payload: requests.PasswordChangePayload{
				CurrentPassword: "current_password",
				NewPassword:     "new_password123",
			},
			wantError: true,
			errorType: types.ErrorTypeSystem,
		},
		{
			name: "database error updating password",
			setup: func(eventEmitter *mocks.EventEmitter, userRepo *authMocks.UserRepository, user *authMocks.User) {
				mockUser := authMocks.NewUser(t)
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("current_password"), bcrypt.DefaultCost)
				mockUser.On("GetPasswordHash").Return(string(hashedPassword))
				mockUser.On("SetPassword", mock.AnythingOfType("string")).Return(nil)

				userID := uuid.New()
				user.On("GetID").Return(userID)

				userRepo.On("FindByID", mock.Anything, userID).Return(mockUser, nil)
				userRepo.On("Update", mock.Anything, mockUser, mock.MatchedBy(
					func(updates map[string]any) bool {
						_, hasPassword := updates["password"]
						return hasPassword
					})).Return(assert.AnError)
			},
			payload: requests.PasswordChangePayload{
				CurrentPassword: "current_password",
				NewPassword:     "new_password123",
			},
			wantError: true,
			errorType: types.ErrorTypeSystem,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventEmitter := mocks.NewEventEmitter(t)
			userRepo := authMocks.NewUserRepository(t)
			user := authMocks.NewUser(t)

			if tt.setup != nil {
				tt.setup(eventEmitter, userRepo, user)
			}

			manager := NewManager(eventEmitter, userRepo)
			err := manager.ChangePassword(context.Background(), user, tt.payload)

			if tt.wantError {
				assert.Error(t, err)
				appErr, ok := err.(*types.AppError)
				assert.True(t, ok)
				if tt.errorType != types.ErrorType(0) {
					assert.Equal(t, tt.errorType, appErr.Type)
				}
			} else {
				assert.NoError(t, err)
			}

			eventEmitter.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			user.AssertExpectations(t)
		})
	}
}
