package password

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/volvlabs/nebularcore/modules/auth/emitter"
	"github.com/volvlabs/nebularcore/modules/auth/interfaces"
	"github.com/volvlabs/nebularcore/modules/auth/models/requests"
	"github.com/volvlabs/nebularcore/tools/types"
	"golang.org/x/crypto/bcrypt"
)

type Manager interface {
	RequestPasswordReset(ctx context.Context, payload requests.PasswordResetPayload) error
	VerifyPasswordReset(ctx context.Context, payload requests.PasswordResetVerifyPayload) error
	ChangePassword(ctx context.Context, user interfaces.User, payload requests.PasswordChangePayload) error
}

type manager struct {
	eventEmitter emitter.EventEmitter
	userRepo     interfaces.UserRepository
}

func NewManager(eventEmitter emitter.EventEmitter, userRepo interfaces.UserRepository) Manager {
	return &manager{
		eventEmitter: eventEmitter,
		userRepo:     userRepo,
	}
}

func (h *manager) RequestPasswordReset(
	ctx context.Context,
	payload requests.PasswordResetPayload,
) error {
	user, err := h.userRepo.FindByEmail(ctx, payload.Email)
	if err != nil {
		log.Error().Err(err).Msgf("Error finding user by email with email: %s", payload.Email)
		return nil
	}

	if user == nil {
		log.Info().Msgf("User with email: %s does not exist", payload.Email)
		return nil
	}

	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		log.Error().Err(err).Msg("Error generating reset token")
		return types.NewSystemError("Failed to process request")
	}
	resetToken := hex.EncodeToString(token)

	now := time.Now()
	user.SetPasswordResetToken(&resetToken)
	user.SetPasswordResetAt(&now)

	updates := map[string]any{
		"password_reset_token": resetToken,
		"password_reset_at":    now,
	}

	if err := h.userRepo.Update(ctx, user, updates); err != nil {
		log.Error().Err(err).Msg("Error updating user with reset token")
		return types.NewSystemError("Failed to process request")
	}

	if err := h.eventEmitter.EmitPasswordResetInitiatedEvent(ctx, user, resetToken); err != nil {
		log.Error().Err(err).Msg("failed to emit password reset initiated event")
	}
	return nil
}

func (h *manager) VerifyPasswordReset(ctx context.Context, payload requests.PasswordResetVerifyPayload) error {
	user, err := h.userRepo.FindByResetToken(ctx, payload.Token)
	if err != nil {
		log.Error().Err(err).Msg("Error finding user by reset token")
		return types.NewUserError("Invalid or expired reset token")
	}

	if user == nil {
		return types.NewUserError("Invalid or expired reset token")
	}

	resetAt := user.GetPasswordResetAt()
	if resetAt == nil || time.Since(*resetAt) > 24*time.Hour {
		return types.NewUserError("Reset token has expired")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Error hashing password")
		return types.NewSystemError("Failed to process request")
	}

	user.SetPassword(string(hashedPassword))
	user.SetPasswordResetToken(nil)
	user.SetPasswordResetAt(nil)

	updates := map[string]any{
		"password":             string(hashedPassword),
		"password_reset_token": nil,
		"password_reset_at":    nil,
	}

	if err := h.userRepo.Update(ctx, user, updates); err != nil {
		log.Error().Err(err).Msg("Error updating user password")
		return types.NewSystemError("Failed to process request")
	}

	if err := h.eventEmitter.EmitPasswordChangedEvent(ctx, user); err != nil {
		log.Error().Err(err).Msg("failed to emit password changed event")
	}
	return nil
}

func (h *manager) ChangePassword(ctx context.Context, userInfo interfaces.User, payload requests.PasswordChangePayload) error {
	userID := userInfo.GetID()

	user, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("Error finding user")
		return types.NewSystemError("Failed to process request")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.GetPasswordHash()), []byte(payload.CurrentPassword)); err != nil {
		return types.NewUserError("Current password is incorrect")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Error hashing password")
		return types.NewSystemError("Failed to process request")
	}

	user.SetPassword(string(hashedPassword))
	updates := map[string]any{
		"password": string(hashedPassword),
	}

	if err := h.userRepo.Update(ctx, user, updates); err != nil {
		log.Error().Err(err).Msg("Error updating user password")
		return types.NewSystemError("Failed to process request")
	}

	h.eventEmitter.EmitPasswordChangedEvent(ctx, user)
	return nil
}
