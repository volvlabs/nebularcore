package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	autherrors "gitlab.com/jideobs/nebularcore/modules/auth/errors"
	"gitlab.com/jideobs/nebularcore/modules/auth/factories"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces"
	"gitlab.com/jideobs/nebularcore/modules/auth/models"
	"gorm.io/gorm"
)

// UserRepository handles user-related database operations
type UserRepository struct {
	db      *gorm.DB
	factory interfaces.UserRepositoryFactory
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB, factory interfaces.UserRepositoryFactory) *UserRepository {
	if factory == nil {
		factory = factories.NewDefaultUserFactory()
	}
	return &UserRepository{
		db:      db,
		factory: factory,
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, data map[string]any) (interfaces.User, error) {
	user := r.factory.NewUser()
	if err := r.db.WithContext(ctx).Model(user).Create(data).Error; err != nil {
		return nil, err
	}
	return user, nil
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
