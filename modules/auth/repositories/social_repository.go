package repositories

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/volvlabs/nebularcore/modules/auth/errors"
	"github.com/volvlabs/nebularcore/modules/auth/models"
	"github.com/volvlabs/nebularcore/modules/auth/types"
	"gorm.io/gorm"
)

// SocialAccountRepository handles social account operations
type SocialAccountRepository struct {
	db *gorm.DB
}

// NewSocialAccountRepository creates a new social account repository
func NewSocialAccountRepository(db *gorm.DB) *SocialAccountRepository {
	return &SocialAccountRepository{
		db: db,
	}
}

// Create creates a new social account
func (r *SocialAccountRepository) Create(
	ctx context.Context,
	data *models.SocialAccount,
) (*models.SocialAccount, error) {
	account := &models.SocialAccount{}
	if err := r.db.WithContext(ctx).
		Model(account).
		Create(data).Error; err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" { // unique_violation
			switch {
			case strings.Contains(pgErr.ConstraintName, "idx_social_accounts_provider_id"):
			case strings.Contains(pgErr.ConstraintName, "social_accounts_provider_provider_user_id_key"):
				return nil, errors.ErrProviderUserIDExists
			case strings.Contains(pgErr.ConstraintName, "user_id"):
				return nil, errors.ErrUserIDExists
			case strings.Contains(pgErr.ConstraintName, "idx_social_accounts_provider_email"):
				return nil, errors.ErrSocialEmailExists
			}
		}
		return nil, err
	}
	return account, nil
}

func (r *SocialAccountRepository) DeleteByUserID(
	ctx context.Context,
	userID string,
) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.SocialAccount{}).Error; err != nil {
		return err
	}
	return nil
}

// FindByProvider finds a social account by provider and provider ID
func (r *SocialAccountRepository) FindByProvider(
	ctx context.Context,
	provider types.AuthProvider,
	providerID string,
) (*models.SocialAccount, error) {
	account := &models.SocialAccount{}
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("provider = ? AND provider_user_id = ?", provider, providerID).
		First(account).Error; err != nil {
		return nil, err
	}
	return account, nil
}

// FindByUserID finds all social accounts for a user
func (r *SocialAccountRepository) FindByUserID(
	ctx context.Context,
	userID string,
) (*models.SocialAccount, error) {
	account := &models.SocialAccount{}
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("user_id = ?", userID).
		First(account).Error; err != nil {
		return nil, err
	}
	return account, nil
}
