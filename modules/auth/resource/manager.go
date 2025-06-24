package resource

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/morkid/paginate"
	"gitlab.com/jideobs/nebularcore/modules/auth/models"
	"gitlab.com/jideobs/nebularcore/modules/auth/models/requests"
	"gitlab.com/jideobs/nebularcore/modules/auth/models/responses"
	"gitlab.com/jideobs/nebularcore/modules/auth/repositories"
	"gitlab.com/jideobs/nebularcore/tools/validation"
)

type Manager interface {
	AddResource(ctx context.Context, payload *requests.AddResourcePayload) (*models.Resource, error)
	GetResource(ctx context.Context, id uuid.UUID) (*models.Resource, error)
	ListResources(ctx context.Context, req *http.Request) paginate.Page
	FilterResource(ctx context.Context, filter string) ([]responses.FilterResourcePayload, error)
	UpdateResource(ctx context.Context, id uuid.UUID, payload *requests.UpdateResourcePayload) (*models.Resource, error)
	DeleteResource(ctx context.Context, id uuid.UUID) error
}

type manager struct {
	resourceRepository repositories.ResourceRepository
	validator          validation.Validator
}

func NewManager(resourceRepository repositories.ResourceRepository) Manager {
	return &manager{
		resourceRepository: resourceRepository,
		validator:          *validation.New(),
	}
}

// AddResource implements Manager.
func (m *manager) AddResource(
	ctx context.Context,
	payload *requests.AddResourcePayload,
) (*models.Resource, error) {
	if err := m.validator.Validate(payload); err != nil {
		return nil, err
	}

	resource := &models.Resource{
		Resource:    payload.Resource,
		Actions:     payload.Actions,
		Description: payload.Description,
		Metadata:    payload.Metadata,
	}

	if err := m.resourceRepository.Create(ctx, resource); err != nil {
		return nil, err
	}
	return resource, nil
}

// DeleteResource implements Manager.
func (m *manager) DeleteResource(ctx context.Context, id uuid.UUID) error {
	return m.resourceRepository.Delete(ctx, id)
}

// GetResource implements Manager.
func (m *manager) GetResource(ctx context.Context, id uuid.UUID) (*models.Resource, error) {
	return m.resourceRepository.FindByID(ctx, id)
}

// ListResources implements Manager.
func (m *manager) ListResources(ctx context.Context, req *http.Request) paginate.Page {
	return m.resourceRepository.List(ctx, req)
}

func (m *manager) FilterResource(ctx context.Context, filter string) ([]responses.FilterResourcePayload, error) {
	return m.resourceRepository.Filter(ctx, filter)
}

// UpdateResource implements Manager.
func (m *manager) UpdateResource(ctx context.Context, id uuid.UUID, payload *requests.UpdateResourcePayload) (*models.Resource, error) {
	if err := m.validator.Validate(payload); err != nil {
		return nil, err
	}

	resource := &models.Resource{
		Resource:    payload.Resource,
		Actions:     payload.Actions,
		Description: payload.Description,
		Metadata:    payload.Metadata,
	}

	if err := m.resourceRepository.Update(ctx, id, resource); err != nil {
		return nil, err
	}
	return resource, nil
}
