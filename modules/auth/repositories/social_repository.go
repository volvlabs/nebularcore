package repositories

import (
	"context"

	"gorm.io/gorm"
	"gitlab.com/jideobs/nebularcore/modules/auth/models"
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
func (r *SocialAccountRepository) Create(ctx context.Context, data map[string]interface{}) (*models.SocialAccount, error) {
	account := &models.SocialAccount{}
	if err := r.db.WithContext(ctx).Model(account).Create(data).Error; err != nil {
		return nil, err
	}
	return account, nil
}

// FindByProvider finds a social account by provider and provider ID
func (r *SocialAccountRepository) FindByProvider(ctx context.Context, provider, providerID string) (*models.SocialAccount, error) {
	account := &models.SocialAccount{}
	if err := r.db.WithContext(ctx).Where("provider = ? AND provider_id = ?", provider, providerID).First(account).Error; err != nil {
		return nil, err
	}
	return account, nil
}

// FindByUserID finds all social accounts for a user
func (r *SocialAccountRepository) FindByUserID(ctx context.Context, userID uint) ([]*models.SocialAccount, error) {
	var accounts []*models.SocialAccount
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}
