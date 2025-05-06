package repositories

import (
	"context"
	"errors"
	"time"

	autherrors "gitlab.com/jideobs/nebularcore/modules/auth/errors"
	"gitlab.com/jideobs/nebularcore/modules/auth/models"
	"gorm.io/gorm"
)

// APIKeyRepository implements auth.APIKeyRepository with tenant awareness
type APIKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{
		db: db,
	}
}

// Create creates new API credentials
func (r *APIKeyRepository) Create(ctx context.Context, data map[string]interface{}) (models.APICredentials, error) {
	var creds models.APICredentials
	err := r.db.WithContext(ctx).Create(&creds).Error
	return creds, err
}

// FindByKey finds API credentials by API key
func (r *APIKeyRepository) FindByKey(ctx context.Context, apiKey string) (models.APICredentials, error) {
	var creds models.APICredentials
	err := r.db.WithContext(ctx).Where("api_key = ?", apiKey).First(&creds).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return creds, autherrors.ErrInvalidCredentials
		}
		return creds, err
	}
	return creds, nil
}

// FindByUser finds all API credentials for a user
func (r *APIKeyRepository) FindByUser(ctx context.Context, userID string) ([]models.APICredentials, error) {
	var creds []models.APICredentials
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&creds).Error
	return creds, err
}

// UpdateLastUsed updates the last used timestamp for API credentials
func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, creds models.APICredentials) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&creds).Update("last_used_at", &now).Error
}

// Revoke revokes (soft deletes) API credentials
func (r *APIKeyRepository) Revoke(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&models.APICredentials{}).
		Where("id = ?", id).
		Update("status", "revoked").Error
}
