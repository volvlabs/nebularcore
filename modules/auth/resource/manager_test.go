package resource_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/morkid/paginate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.com/jideobs/nebularcore/modules/auth/models"
	"gitlab.com/jideobs/nebularcore/modules/auth/models/requests"
	"gitlab.com/jideobs/nebularcore/modules/auth/repositories/mocks"
	"gitlab.com/jideobs/nebularcore/modules/auth/resource"
	"gitlab.com/jideobs/nebularcore/tools/types"
)

func TestAddResource(t *testing.T) {
	type args struct {
		ctx     context.Context
		payload *requests.AddResourcePayload
	}

	tests := []struct {
		name      string
		args      args
		setupRepo func(*mocks.ResourceRepository)
		want      *models.Resource
		wantErr   error
	}{
		{
			name: "successful_resource_creation",
			args: args{
				ctx: context.Background(),
				payload: &requests.AddResourcePayload{
					Module:      "auth",
					Resource:    "users",
					Actions:     []string{"read", "write"},
					Description: "User resource",
					Metadata:    map[string]interface{}{"key": "value"},
				},
			},
			setupRepo: func(m *mocks.ResourceRepository) {
				m.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			want: &models.Resource{
				Resource:    "users",
				Actions:     []string{"read", "write"},
				Description: "User resource",
				Metadata:    map[string]interface{}{"key": "value"},
			},
		},
		{
			name: "validation_error",
			args: args{
				ctx:     context.Background(),
				payload: &requests.AddResourcePayload{},
			},
			wantErr: &types.AppError{
				Type:    types.ErrorTypeValidation,
				Message: "Validation failed. Please check the provided values and try again.",
				Errors: []types.FieldError{
					{
						Field:   "Module",
						Message: "Module is a required field",
					},
					{
						Field:   "Resource",
						Message: "Resource is a required field",
					},
					{
						Field:   "Actions",
						Message: "Actions is a required field",
					},
				},
			},
		},
		{
			name: "repository_error",
			args: args{
				ctx: context.Background(),
				payload: &requests.AddResourcePayload{
					Module:   "auth",
					Resource: "users",
					Actions:  []string{"read", "write"},
				},
			},
			setupRepo: func(m *mocks.ResourceRepository) {
				m.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewResourceRepository(t)

			if tt.setupRepo != nil {
				tt.setupRepo(mockRepo)
			}

			manager := resource.NewManager(mockRepo)

			gotResource, err := manager.AddResource(tt.args.ctx, tt.args.payload)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, gotResource)
		})
	}
}

func TestGetResource(t *testing.T) {
	type args struct {
		ctx context.Context
		id  uuid.UUID
	}

	testID := uuid.New()
	testResource := &models.Resource{
		Resource:    "users",
		Actions:     []string{"read"},
		Description: "User resource",
	}

	tests := []struct {
		name      string
		args      args
		setupRepo func(*mocks.ResourceRepository)
		want      *models.Resource
		wantErr   error
	}{
		{
			name: "successful get",
			args: args{
				ctx: context.Background(),
				id:  testID,
			},
			setupRepo: func(repo *mocks.ResourceRepository) {
				repo.On("FindByID", mock.Anything, testID).Return(testResource, nil)
			},
			want:    testResource,
			wantErr: nil,
		},
		{
			name: "not found",
			args: args{
				ctx: context.Background(),
				id:  testID,
			},
			setupRepo: func(repo *mocks.ResourceRepository) {
				repo.On("FindByID", mock.Anything, testID).Return(nil, errors.New("not found"))
			},
			want:    nil,
			wantErr: errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewResourceRepository(t)

			if tt.setupRepo != nil {
				tt.setupRepo(mockRepo)
			}

			manager := resource.NewManager(mockRepo)

			gotResource, err := manager.GetResource(tt.args.ctx, tt.args.id)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, gotResource)
		})
	}
}

func TestListResources(t *testing.T) {
	type args struct {
		ctx context.Context
		req *http.Request
	}

	testReq, _ := http.NewRequest("GET", "http://example.com/resources?page=1", nil)
	testPage := paginate.Page{
		Items: []interface{}{&models.Resource{
			Resource: "users",
			Actions:  []string{"read"},
		}},
		Total: 1,
	}

	tests := []struct {
		name      string
		args      args
		setupRepo func(*mocks.ResourceRepository)
		wantPage  paginate.Page
	}{
		{
			name: "successful list",
			args: args{
				ctx: context.Background(),
				req: testReq,
			},
			setupRepo: func(repo *mocks.ResourceRepository) {
				repo.On("List", mock.Anything, testReq).Return(testPage)
			},
			wantPage: testPage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewResourceRepository(t)

			if tt.setupRepo != nil {
				tt.setupRepo(mockRepo)
			}

			manager := resource.NewManager(mockRepo)

			gotPage := manager.ListResources(tt.args.ctx, tt.args.req)
			assert.Equal(t, tt.wantPage, gotPage)
		})
	}
}

func TestUpdateResource(t *testing.T) {
	type args struct {
		ctx     context.Context
		id      uuid.UUID
		payload *requests.UpdateResourcePayload
	}

	testID := uuid.New()
	tests := []struct {
		name      string
		args      args
		setupRepo func(*mocks.ResourceRepository)
		want      *models.Resource
		wantErr   error
	}{
		{
			name: "successful_update",
			args: args{
				ctx: context.Background(),
				id:  testID,
				payload: &requests.UpdateResourcePayload{
					Resource:    "users",
					Actions:     []string{"read", "write"},
					Description: "User resource",
					Metadata:    map[string]interface{}{"key": "value"},
				},
			},
			setupRepo: func(m *mocks.ResourceRepository) {
				m.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			want: &models.Resource{
				Resource:    "users",
				Actions:     []string{"read", "write"},
				Description: "User resource",
				Metadata:    map[string]interface{}{"key": "value"},
			},
		},
		{
			name: "validation_error",
			args: args{
				ctx:     context.Background(),
				id:      uuid.New(),
				payload: &requests.UpdateResourcePayload{},
			},
			wantErr: &types.AppError{
				Type:    types.ErrorTypeValidation,
				Message: "Validation failed. Please check the provided values and try again.",
				Errors: []types.FieldError{
					{
						Field:   "Resource",
						Message: "Resource is a required field",
					},
					{
						Field:   "Actions",
						Message: "Actions is a required field",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewResourceRepository(t)

			if tt.setupRepo != nil {
				tt.setupRepo(mockRepo)
			}

			manager := resource.NewManager(mockRepo)

			gotResource, err := manager.UpdateResource(tt.args.ctx, tt.args.id, tt.args.payload)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, gotResource)
		})
	}
}

func TestDeleteResource(t *testing.T) {
	testID := uuid.New()

	type args struct {
		ctx context.Context
		id  uuid.UUID
	}

	tests := []struct {
		name      string
		args      args
		setupRepo func(*mocks.ResourceRepository)
		wantErr   error
	}{
		{
			name: "successful deletion",
			args: args{
				ctx: context.Background(),
				id:  testID,
			},
			setupRepo: func(repo *mocks.ResourceRepository) {
				repo.On("Delete", mock.Anything, testID).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "not found",
			args: args{
				ctx: context.Background(),
				id:  testID,
			},
			setupRepo: func(repo *mocks.ResourceRepository) {
				repo.On("Delete", mock.Anything, testID).Return(errors.New("not found"))
			},
			wantErr: errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewResourceRepository(t)

			if tt.setupRepo != nil {
				tt.setupRepo(mockRepo)
			}

			manager := resource.NewManager(mockRepo)

			err := manager.DeleteResource(tt.args.ctx, tt.args.id)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
