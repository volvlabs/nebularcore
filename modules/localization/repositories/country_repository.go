package repositories

import (
	"context"
	"errors"

	"github.com/volvlabs/nebularcore/modules/localization/models"
	"gorm.io/gorm"
)

var ErrCountryNotFound = errors.New("country not found")

// CountryRepository is the persistence contract for the countries reference
// table.
type CountryRepository interface {
	FindByCode(ctx context.Context, code string) (*models.Country, error)
	ListActive(ctx context.Context) ([]models.Country, error)
	// ListAll returns every seeded country regardless of IsActive, ordered by
	// Name — for an admin surface that needs to activate/deactivate
	// countries, not just read the currently-active set.
	ListAll(ctx context.Context) ([]models.Country, error)
	SetActive(ctx context.Context, code string, active bool) error
}

type countryRepository struct {
	db *gorm.DB
}

func NewCountryRepository(db *gorm.DB) CountryRepository {
	return &countryRepository{db: db}
}

func (r *countryRepository) FindByCode(ctx context.Context, code string) (*models.Country, error) {
	var country models.Country
	if err := r.db.WithContext(ctx).First(&country, "code = ?", code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCountryNotFound
		}
		return nil, err
	}
	return &country, nil
}

func (r *countryRepository) ListActive(ctx context.Context) ([]models.Country, error) {
	var countries []models.Country
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&countries).Error; err != nil {
		return nil, err
	}
	return countries, nil
}

func (r *countryRepository) ListAll(ctx context.Context) ([]models.Country, error) {
	var countries []models.Country
	if err := r.db.WithContext(ctx).Order("name").Find(&countries).Error; err != nil {
		return nil, err
	}
	return countries, nil
}

func (r *countryRepository) SetActive(ctx context.Context, code string, active bool) error {
	res := r.db.WithContext(ctx).Model(&models.Country{}).Where("code = ?", code).Update("is_active", active)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrCountryNotFound
	}
	return nil
}
