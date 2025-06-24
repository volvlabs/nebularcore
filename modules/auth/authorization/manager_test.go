package authorization

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.com/jideobs/nebularcore/modules/auth/models"
	"gitlab.com/jideobs/nebularcore/modules/auth/models/requests"
	"gitlab.com/jideobs/nebularcore/modules/auth/repositories/mocks"
	"gitlab.com/jideobs/nebularcore/tools/types"
)

func TestCreateRole(t *testing.T) {
	type args struct {
		ctx     context.Context
		payload *requests.CreateRolePayload
	}
	tests := []struct {
		name  string
		setup func(*mocks.RoleRepository)
		args  args
		err   error
	}{
		{
			name: "successful role creation",
			args: args{
				ctx: context.Background(),
				payload: &requests.CreateRolePayload{
					Name:        "admin",
					Description: "admin role",
					Metadata:    nil,
				},
			},
			setup: func(roleRepo *mocks.RoleRepository) {
				role := &models.Role{
					Name:        "admin",
					Description: "admin role",
					Metadata:    nil,
				}
				roleRepo.On("CreateRole", mock.Anything, role).
					Return(&models.Role{}, nil)
			},
			err: nil,
		},
		{
			name: "empty role name",
			args: args{
				ctx: context.Background(),
				payload: &requests.CreateRolePayload{
					Name:        "",
					Description: "admin role",
					Metadata:    nil,
				},
			},
			setup: func(roleRepo *mocks.RoleRepository) {
			},
			err: types.NewValidationError("Validation failed. Please check the provided values and try again.", []types.FieldError{
				{
					Field:   "Name",
					Message: "Name is a required field",
				},
			}),
		},
		{
			name: "database error",
			args: args{
				ctx: context.Background(),
				payload: &requests.CreateRolePayload{
					Name:        "admin",
					Description: "admin role",
					Metadata:    nil,
				},
			},
			setup: func(roleRepo *mocks.RoleRepository) {
				roleRepo.On("CreateRole", mock.Anything, mock.AnythingOfType("*models.Role")).
					Return(nil, errors.New("db error"))
			},
			err: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRoleRepo := mocks.NewRoleRepository(t)
			mockPermissionRepo := mocks.NewPermissionRepository(t)

			if tt.setup != nil {
				tt.setup(mockRoleRepo)
			}

			manager := NewAuthorizationManager(mockRoleRepo, mockPermissionRepo)

			err := manager.CreateRole(context.Background(), tt.args.payload)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestAssignRole(t *testing.T) {
	type args struct {
		ctx      context.Context
		userID   uuid.UUID
		roleID   uuid.UUID
		roleName string
	}

	testUserID := uuid.New()
	testRoleID := uuid.New()
	tests := []struct {
		name  string
		setup func(*mocks.RoleRepository)
		args  args
		err   error
	}{
		{
			name: "successful role assignment",
			args: args{
				ctx:      context.Background(),
				userID:   testUserID,
				roleID:   testRoleID,
				roleName: "admin",
			},
			setup: func(roleRepo *mocks.RoleRepository) {
				roleRepo.On("AssignRole", mock.Anything, testUserID, testRoleID, "admin").
					Return(nil)
			},
			err: nil,
		},
		{
			name: "database error",
			args: args{
				ctx:      context.Background(),
				userID:   testUserID,
				roleID:   testRoleID,
				roleName: "admin",
			},
			setup: func(roleRepo *mocks.RoleRepository) {
				roleRepo.On("AssignRole", mock.Anything, testUserID, testRoleID, "admin").
					Return(errors.New("db error"))
			},
			err: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRoleRepo := mocks.NewRoleRepository(t)
			mockPermissionRepo := mocks.NewPermissionRepository(t)

			if tt.setup != nil {
				tt.setup(mockRoleRepo)
			}

			manager := NewAuthorizationManager(mockRoleRepo, mockPermissionRepo)

			err := manager.AssignRole(context.Background(), tt.args.userID, tt.args.roleID, tt.args.roleName)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestUnassignRole(t *testing.T) {
	type args struct {
		ctx      context.Context
		userID   uuid.UUID
		roleID   uuid.UUID
		roleName string
	}

	testUserID := uuid.New()
	testRoleID := uuid.New()
	tests := []struct {
		name  string
		setup func(*mocks.RoleRepository)
		args  args
		err   error
	}{
		{
			name: "successful role unassignment",
			args: args{
				ctx:      context.Background(),
				userID:   testUserID,
				roleID:   testRoleID,
				roleName: "admin",
			},
			setup: func(roleRepo *mocks.RoleRepository) {
				roleRepo.On("UnassignRole", mock.Anything, testUserID, testRoleID, "admin").
					Return(nil)
			},
			err: nil,
		},
		{
			name: "database error",
			args: args{
				ctx:      context.Background(),
				userID:   testUserID,
				roleID:   testRoleID,
				roleName: "admin",
			},
			setup: func(roleRepo *mocks.RoleRepository) {
				roleRepo.On("UnassignRole", mock.Anything, testUserID, testRoleID, "admin").
					Return(errors.New("db error"))
			},
			err: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRoleRepo := mocks.NewRoleRepository(t)
			mockPermissionRepo := mocks.NewPermissionRepository(t)

			if tt.setup != nil {
				tt.setup(mockRoleRepo)
			}

			manager := NewAuthorizationManager(mockRoleRepo, mockPermissionRepo)

			err := manager.UnassignRole(context.Background(), tt.args.userID, tt.args.roleID, tt.args.roleName)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestGrantRolePermission(t *testing.T) {
	type args struct {
		ctx          context.Context
		roleID       uuid.UUID
		permissionID uuid.UUID
		grantedByID  uuid.UUID
		roleName     string
	}

	testRoleID := uuid.New()
	testPermissionID := uuid.New()
	testGrantedByID := uuid.New()
	tests := []struct {
		name  string
		setup func(*mocks.PermissionRepository)
		args  args
		err   error
	}{
		{
			name: "successful permission grant",
			args: args{
				ctx:          context.Background(),
				roleID:       testRoleID,
				permissionID: testPermissionID,
				grantedByID:  testGrantedByID,
				roleName:     "admin",
			},
			setup: func(permRepo *mocks.PermissionRepository) {
				permRepo.On("AssignPermissionToRole", mock.Anything, testRoleID, testPermissionID, testGrantedByID, "admin").
					Return(nil)
			},
			err: nil,
		},
		{
			name: "database error",
			args: args{
				ctx:          context.Background(),
				roleID:       testRoleID,
				permissionID: testPermissionID,
				grantedByID:  testGrantedByID,
				roleName:     "admin",
			},
			setup: func(permRepo *mocks.PermissionRepository) {
				permRepo.On("AssignPermissionToRole", mock.Anything, testRoleID, testPermissionID, testGrantedByID, "admin").
					Return(errors.New("db error"))
			},
			err: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRoleRepo := mocks.NewRoleRepository(t)
			mockPermissionRepo := mocks.NewPermissionRepository(t)

			if tt.setup != nil {
				tt.setup(mockPermissionRepo)
			}

			manager := NewAuthorizationManager(mockRoleRepo, mockPermissionRepo)

			err := manager.GrantRolePermission(context.Background(), tt.args.roleID, tt.args.permissionID, tt.args.grantedByID, tt.args.roleName)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestRevokeRolePermission(t *testing.T) {
	type args struct {
		ctx          context.Context
		roleID       uuid.UUID
		permissionID uuid.UUID
		roleName     string
	}

	testRoleID := uuid.New()
	testPermissionID := uuid.New()
	tests := []struct {
		name  string
		setup func(*mocks.PermissionRepository)
		args  args
		err   error
	}{
		{
			name: "successful permission revoke",
			args: args{
				ctx:          context.Background(),
				roleID:       testRoleID,
				permissionID: testPermissionID,
				roleName:     "admin",
			},
			setup: func(permRepo *mocks.PermissionRepository) {
				permRepo.On("RevokePermissionToRole", mock.Anything, testRoleID, testPermissionID, "admin").
					Return(nil)
			},
			err: nil,
		},
		{
			name: "database error",
			args: args{
				ctx:          context.Background(),
				roleID:       testRoleID,
				permissionID: testPermissionID,
				roleName:     "admin",
			},
			setup: func(permRepo *mocks.PermissionRepository) {
				permRepo.On("RevokePermissionToRole", mock.Anything, testRoleID, testPermissionID, "admin").
					Return(errors.New("db error"))
			},
			err: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRoleRepo := mocks.NewRoleRepository(t)
			mockPermissionRepo := mocks.NewPermissionRepository(t)

			if tt.setup != nil {
				tt.setup(mockPermissionRepo)
			}

			manager := NewAuthorizationManager(mockRoleRepo, mockPermissionRepo)

			err := manager.RevokeRolePermission(context.Background(), tt.args.roleID, tt.args.permissionID, tt.args.roleName)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestCreatePermission(t *testing.T) {
	type args struct {
		ctx     context.Context
		payload *requests.CreatePermissionPayload
	}

	testResourceID := uuid.New()
	tests := []struct {
		name  string
		setup func(*mocks.PermissionRepository)
		args  args
		want  *models.Permission
		err   error
	}{
		{
			name: "successful permission creation",
			args: args{
				ctx: context.Background(),
				payload: &requests.CreatePermissionPayload{
					Name:        "read:users",
					ResourceID:  testResourceID,
					Action:      "read",
					Description: "Read users permission",
					Metadata:    nil,
				},
			},
			setup: func(permRepo *mocks.PermissionRepository) {
				permission := &models.Permission{
					Name:        "read:users",
					ResourceID:  testResourceID,
					Action:      "read",
					Description: "Read users permission",
					Metadata:    nil,
				}
				permRepo.On("CreatePermission", mock.Anything, permission).
					Return(permission, nil)
			},
			want: &models.Permission{
				Name:        "read:users",
				ResourceID:  testResourceID,
				Action:      "read",
				Description: "Read users permission",
				Metadata:    nil,
			},
			err: nil,
		},
		{
			name: "database error",
			args: args{
				ctx: context.Background(),
				payload: &requests.CreatePermissionPayload{
					Name:        "read:users",
					ResourceID:  testResourceID,
					Action:      "read",
					Description: "Read users permission",
					Metadata:    nil,
				},
			},
			setup: func(permRepo *mocks.PermissionRepository) {
				permRepo.On("CreatePermission", mock.Anything, mock.AnythingOfType("*models.Permission")).
					Return(nil, errors.New("db error"))
			},
			want: nil,
			err:  errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRoleRepo := mocks.NewRoleRepository(t)
			mockPermissionRepo := mocks.NewPermissionRepository(t)

			if tt.setup != nil {
				tt.setup(mockPermissionRepo)
			}

			manager := NewAuthorizationManager(mockRoleRepo, mockPermissionRepo)

			got, err := manager.CreatePermission(context.Background(), tt.args.payload)
			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGrantUserPermission(t *testing.T) {
	type args struct {
		ctx          context.Context
		userID       uuid.UUID
		permissionID uuid.UUID
		grantedByID  uuid.UUID
	}

	testUserID := uuid.New()
	testPermissionID := uuid.New()
	testGrantedByID := uuid.New()
	tests := []struct {
		name  string
		setup func(*mocks.PermissionRepository)
		args  args
		err   error
	}{
		{
			name: "successful permission grant",
			args: args{
				ctx:          context.Background(),
				userID:       testUserID,
				permissionID: testPermissionID,
				grantedByID:  testGrantedByID,
			},
			setup: func(permRepo *mocks.PermissionRepository) {
				permRepo.On("AssignPermissionToUser", mock.Anything, testUserID, testPermissionID, testGrantedByID).
					Return(nil)
			},
			err: nil,
		},
		{
			name: "database error",
			args: args{
				ctx:          context.Background(),
				userID:       testUserID,
				permissionID: testPermissionID,
				grantedByID:  testGrantedByID,
			},
			setup: func(permRepo *mocks.PermissionRepository) {
				permRepo.On("AssignPermissionToUser", mock.Anything, testUserID, testPermissionID, testGrantedByID).
					Return(errors.New("db error"))
			},
			err: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRoleRepo := mocks.NewRoleRepository(t)
			mockPermissionRepo := mocks.NewPermissionRepository(t)

			if tt.setup != nil {
				tt.setup(mockPermissionRepo)
			}

			manager := NewAuthorizationManager(mockRoleRepo, mockPermissionRepo)

			err := manager.GrantUserPermission(context.Background(), tt.args.userID, tt.args.permissionID, tt.args.grantedByID)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestRevokeUserPermission(t *testing.T) {
	type args struct {
		ctx          context.Context
		userID       uuid.UUID
		permissionID uuid.UUID
	}

	testUserID := uuid.New()
	testPermissionID := uuid.New()
	tests := []struct {
		name  string
		setup func(*mocks.PermissionRepository)
		args  args
		err   error
	}{
		{
			name: "successful permission revoke",
			args: args{
				ctx:          context.Background(),
				userID:       testUserID,
				permissionID: testPermissionID,
			},
			setup: func(permRepo *mocks.PermissionRepository) {
				permRepo.On("RevokePermissionToUser", mock.Anything, testUserID, testPermissionID).
					Return(nil)
			},
			err: nil,
		},
		{
			name: "database error",
			args: args{
				ctx:          context.Background(),
				userID:       testUserID,
				permissionID: testPermissionID,
			},
			setup: func(permRepo *mocks.PermissionRepository) {
				permRepo.On("RevokePermissionToUser", mock.Anything, testUserID, testPermissionID).
					Return(errors.New("db error"))
			},
			err: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRoleRepo := mocks.NewRoleRepository(t)
			mockPermissionRepo := mocks.NewPermissionRepository(t)

			if tt.setup != nil {
				tt.setup(mockPermissionRepo)
			}

			manager := NewAuthorizationManager(mockRoleRepo, mockPermissionRepo)

			err := manager.RevokeUserPermission(context.Background(), tt.args.userID, tt.args.permissionID)
			assert.Equal(t, tt.err, err)
		})
	}
}
