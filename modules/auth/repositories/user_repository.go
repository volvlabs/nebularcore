package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	autherrors "github.com/volvlabs/nebularcore/modules/auth/errors"
	"github.com/volvlabs/nebularcore/modules/auth/factories"
	"github.com/volvlabs/nebularcore/modules/auth/interfaces"
	"github.com/volvlabs/nebularcore/modules/auth/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserRepository handles user-related database operations
type UserRepository struct {
	db      *gorm.DB
	factory interfaces.UserRepositoryFactory
}

// NewUserRepository creates a new user repository
func NewUserRepository(
	db *gorm.DB,
	factory interfaces.UserRepositoryFactory,
) *UserRepository {
	if factory == nil {
		factory = factories.NewDefaultUserFactory()
	}
	return &UserRepository{
		db:      db,
		factory: factory,
	}
}

// Create creates a new user. Model(user).Create(map) alone leaves the
// returned user with a zero-value ID: GORM only writes DB-generated columns
// (the UUID primary key, defaulted timestamps) back into the value actually
// passed to Create — here that's the map, not the model, since they're
// separate arguments. RETURNING the generated id into the map (verified:
// it does land there) and then re-fetching the full row is what actually
// gets the caller a correctly populated user, rather than one whose
// GetID() is the zero UUID. This affects every caller, including the real
// signup path (backends.NewLocalBackend), not just a bootstrap edge case.
func (r *UserRepository) Create(ctx context.Context, data map[string]any) (interfaces.User, error) {
	user := r.factory.NewUser()
	if err := r.db.WithContext(ctx).Model(user).Clauses(clause.Returning{}).Create(data).Error; err != nil {
		return nil, err
	}

	id, err := coerceUUID(data["id"])
	if err != nil {
		return nil, fmt.Errorf("reading generated user id: %w", err)
	}

	return r.FindByID(ctx, id)
}

// coerceUUID handles the possible concrete types a driver hands back for a
// uuid column scanned into a map[string]any — observed as a string with
// lib/pq/pgx, but handled defensively for []byte and uuid.UUID too.
func coerceUUID(v any) (uuid.UUID, error) {
	switch val := v.(type) {
	case uuid.UUID:
		return val, nil
	case string:
		return uuid.Parse(val)
	case []byte:
		return uuid.ParseBytes(val)
	default:
		return uuid.Nil, fmt.Errorf("unsupported id type %T", v)
	}
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (interfaces.User, error) {
	user := r.factory.NewUser()
	err := r.db.WithContext(ctx).Where("id = ?", id).First(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, autherrors.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (interfaces.User, error) {
	user := r.factory.NewUser()
	err := r.db.WithContext(ctx).Where("email = ?", email).First(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, autherrors.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// FindByPhoneNumber finds a user by phone number
func (r *UserRepository) FindByPhoneNumber(ctx context.Context, phoneNumber string) (interfaces.User, error) {
	user := r.factory.NewUser()
	err := r.db.WithContext(ctx).Where("phone_number = ?", phoneNumber).First(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, autherrors.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// FindByUsername finds a user by username
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (interfaces.User, error) {
	user := r.factory.NewUser()
	err := r.db.WithContext(ctx).Where("username = ?", username).First(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, autherrors.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// FindByPhone finds a user by phone
func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (interfaces.User, error) {
	user := r.factory.NewUser()
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, autherrors.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// FindByIdentifier finds a user by email, username, or phone number
func (r *UserRepository) FindByIdentifier(ctx context.Context, identifier string) (interfaces.User, error) {
	user := r.factory.NewUser()
	err := r.db.WithContext(ctx).Where("email = ? OR username = ? OR phone = ?", identifier, identifier, identifier).First(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, autherrors.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// FindByCredentials finds a user by their credentials
func (r *UserRepository) FindByCredentials(ctx context.Context, username, password string) (interfaces.User, error) {
	user := r.factory.NewUser()
	err := r.db.WithContext(ctx).Where("(username = ? OR email = ? OR phone = ?) AND password = ?", username, username, username, password).First(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, autherrors.ErrInvalidCredentials
		}
		return nil, err
	}
	return user, nil
}

// ValidateCredentials validates a user's credentials
func (r *UserRepository) ValidateCredentials(ctx context.Context, user interfaces.User, password string) error {
	if user.GetPasswordHash() != password {
		return autherrors.ErrInvalidCredentials
	}
	return nil
}

// SetPassword sets a user's password
func (r *UserRepository) SetPassword(ctx context.Context, user interfaces.User, password string) error {
	return r.db.WithContext(ctx).Model(user).Update("password", password).Error
}

// FindByToken finds a user by their token
func (r *UserRepository) FindByToken(ctx context.Context, token string) (interfaces.User, error) {
	user := r.factory.NewUser()
	err := r.db.WithContext(ctx).Where("token = ?", token).First(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, autherrors.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// UpdateToken updates a user's token
func (r *UserRepository) UpdateToken(ctx context.Context, user interfaces.User, token string) error {
	return r.db.WithContext(ctx).Model(user).Update("token", token).Error
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user interfaces.User, data map[string]any) error {
	return r.db.WithContext(ctx).Model(user).Updates(data).Error
}

// Delete soft deletes a user
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(r.factory.NewUser(), "id = ?", id).Error
}

// List lists users with pagination
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]interfaces.User, error) {
	var users []interfaces.User
	var modelUsers []*models.User
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&modelUsers).Error
	users = make([]interfaces.User, len(modelUsers))
	for i, u := range modelUsers {
		users[i] = u
	}
	return users, err
}

// Count counts total number of users
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(r.factory.NewUser()).Count(&count).Error
	return count, err
}

// UpdateLastLogin updates the user's last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, user interfaces.User) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(user).Update("last_login_at", now).Error
}

// ChangePassword updates a user's password hash
func (r *UserRepository) ChangePassword(ctx context.Context, user interfaces.User, newPassword string) error {
	return r.db.WithContext(ctx).Model(user).Update("password", newPassword).Error
}

// ResetPassword resets a user's password
func (r *UserRepository) ResetPassword(ctx context.Context, user interfaces.User, newPassword string) error {
	return r.db.WithContext(ctx).Model(user).Updates(map[string]interface{}{
		"password":             newPassword,
		"password_reset_token": nil,
		"password_reset_at":    nil,
	}).Error
}

// SetPasswordResetToken sets a password reset token for a user
func (r *UserRepository) SetPasswordResetToken(ctx context.Context, user interfaces.User, token string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(user).Updates(map[string]interface{}{
		"password_reset_token": token,
		"password_reset_at":    now,
	}).Error
}

// FindByResetToken finds a user by their password reset token
func (r *UserRepository) FindByResetToken(ctx context.Context, token string) (interfaces.User, error) {
	user := r.factory.NewUser()
	err := r.db.WithContext(ctx).Where("password_reset_token = ?", token).First(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, autherrors.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}
