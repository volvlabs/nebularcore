package repositories

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/morkid/paginate"
	"gitlab.com/jideobs/nebularcore/modules/auth/models"
	"gitlab.com/jideobs/nebularcore/modules/auth/models/responses"
	"gorm.io/gorm"
)

type ResourceRepository interface {
	Create(ctx context.Context, resource *models.Resource) error
	Update(ctx context.Context, id uuid.UUID, resource *models.Resource) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Resource, error)
	FindByModule(ctx context.Context, moduleName string) ([]*models.Resource, error)
	FindByResource(ctx context.Context, resource string) (*models.Resource, error)
	List(ctx context.Context, r *http.Request) paginate.Page
	Filter(ctx context.Context, filter string) ([]responses.FilterResourcePayload, error)
}

// resourceRepository handles module resource-related database operations
type resourceRepository struct {
	db *gorm.DB
}

// NewResourceRepository creates a new module resource repository
func NewResourceRepository(db *gorm.DB) ResourceRepository {
	return &resourceRepository{
		db: db,
	}
}

// Create creates a new module resource
func (r *resourceRepository) Create(ctx context.Context, resource *models.Resource) error {
	return r.db.WithContext(ctx).Create(resource).Error
}

// Update updates an existing module resource
func (r *resourceRepository) Update(ctx context.Context, id uuid.UUID, resource *models.Resource) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Updates(resource).Error
}

// Delete soft-deletes a module resource
func (r *resourceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Resource{}, id).Error
}

// FindByID finds a module resource by ID
func (r *resourceRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Resource, error) {
	var resource models.Resource
	err := r.db.WithContext(ctx).First(&resource, id).Error
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// FindByModule finds all resources for a module
func (r *resourceRepository) FindByModule(ctx context.Context, moduleName string) ([]*models.Resource, error) {
	var resources []*models.Resource
	err := r.db.WithContext(ctx).Where("module_name = ?", moduleName).Find(&resources).Error
	if err != nil {
		return nil, err
	}
	return resources, nil
}

// FindByResource finds a module resource by resource name
func (r *resourceRepository) FindByResource(ctx context.Context, resource string) (*models.Resource, error) {
	var moduleResource models.Resource
	err := r.db.WithContext(ctx).Where("resource = ?", resource).First(&moduleResource).Error
	if err != nil {
		return nil, err
	}
	return &moduleResource, nil
}

// List returns all module resources
func (r *resourceRepository) List(ctx context.Context, req *http.Request) paginate.Page {
	pg := paginate.New()
	db := r.db.WithContext(ctx).Model(&models.Resource{})
	return pg.With(db).Request(req).Response(&[]models.Resource{})
}

func (r *resourceRepository) Filter(ctx context.Context, filter string) ([]responses.FilterResourcePayload, error) {
	resourcesFound := []responses.FilterResourcePayload{}
	err := r.db.WithContext(ctx).Model(&models.Resource{}).
		Where("resource ILIKE '%\\%s%'", filter).
		Find(&resourcesFound).Error
	if err != nil {
		return nil, err
	}

	return resourcesFound, nil
}