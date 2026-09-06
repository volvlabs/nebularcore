package authorization

import (
	"context"
	"fmt"
	"time"

	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/volvlabs/nebularcore/modules/auth/config"
	"github.com/volvlabs/nebularcore/modules/auth/repositories"
	"gorm.io/gorm"
)

// rbacModelConf is the standard casbin RBAC-with-roles model: a subject
// (user) may act on an object if some role granted to it (via g, the role
// mapping set up by AssignRole) has a matching policy (via p, set up by
// GrantPermission). Embedded rather than loaded from a file on disk — this
// manager previously pointed casbin at a relative path ("auth_model.conf")
// that didn't exist anywhere in the module, meaning NewAuthorizationManager
// would fail outright the moment anything actually constructed one.
const rbacModelConf = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

// AuthorizationManager handles role and permission management
type AuthorizationManager struct {
	enforcer     *casbin.Enforcer
	roleRepo     *repositories.RoleRepository
	resourceRepo *repositories.ResourceRepository
	db           *gorm.DB
}

// NewAuthorizationManager creates a new authorization manager. cfg selects
// where casbin policy/grouping rows live (database via gorm-adapter, or a
// file via casbin's own file-adapter) and, optionally, a custom RBAC model
// file — see config.AuthorizationConfig's doc comment. This is the single
// enforcer shared by AuthorizationManager itself and AuthMiddleware (see
// middleware.NewAuthMiddleware), replacing what used to be two
// independently-constructed enforcers.
func NewAuthorizationManager(db *gorm.DB, cfg config.AuthorizationConfig) (*AuthorizationManager, error) {
	adapter, err := newAdapter(db, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin adapter: %w", err)
	}

	casbinModel, err := newModel(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load RBAC model: %w", err)
	}

	enforcer, err := casbin.NewEnforcer(casbinModel, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load policy: %w", err)
	}

	// Initialize repositories
	roleRepo := repositories.NewRoleRepository(db)
	resourceRepo := repositories.NewResourceRepository(db)

	return &AuthorizationManager{
		enforcer:     enforcer,
		roleRepo:     roleRepo,
		resourceRepo: resourceRepo,
		db:           db,
	}, nil
}

// newAdapter picks the casbin policy store: cfg.Source == "file" reads/
// writes PolicyPath via casbin's built-in file-adapter, anything else
// (including the zero value, for backward compatibility) uses gorm-adapter
// against db — today's only behavior before this config existed.
func newAdapter(db *gorm.DB, cfg config.AuthorizationConfig) (persist.Adapter, error) {
	if cfg.Source == "file" {
		return fileadapter.NewAdapter(cfg.PolicyPath), nil
	}
	return gormadapter.NewAdapterByDB(db)
}

// newModel loads a custom RBAC model file if configured, else falls back
// to the package's embedded default.
func newModel(cfg config.AuthorizationConfig) (casbinmodel.Model, error) {
	if cfg.ModelPath != "" {
		return casbinmodel.NewModelFromFile(cfg.ModelPath)
	}
	return casbinmodel.NewModelFromString(rbacModelConf)
}

// Enforcer returns the shared casbin enforcer, for callers (AuthMiddleware,
// the management HTTP handlers) that need direct Enforce/policy calls
// outside the higher-level role/permission methods below.
func (m *AuthorizationManager) Enforcer() *casbin.Enforcer {
	return m.enforcer
}

// CreateRole creates a new role
func (m *AuthorizationManager) CreateRole(ctx context.Context, name, description string, metadata map[string]interface{}) error {
	_, err := m.roleRepo.CreateRole(ctx, map[string]interface{}{
		"name":        name,
		"description": description,
		"metadata":    metadata,
	})
	return err
}

// AssignRole assigns a role to a user
func (m *AuthorizationManager) AssignRole(ctx context.Context, userID, roleName string, duration *time.Duration) error {
	// Start transaction
	tx := m.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find role
	var role repositories.Role
	if err := tx.Where("name = ?", roleName).First(&role).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Calculate expiry if duration is provided
	var expiresAt *time.Time
	if duration != nil {
		t := time.Now().Add(*duration)
		expiresAt = &t
	}

	// Assign role in database
	if err := m.roleRepo.AssignRole(ctx, userID, role.ID, expiresAt); err != nil {
		tx.Rollback()
		return err
	}

	// Add role to Casbin
	if _, err := m.enforcer.AddRoleForUser(userID, roleName); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// UnassignRole removes a role from a user
func (m *AuthorizationManager) UnassignRole(ctx context.Context, userID, roleName string) error {
	// Start transaction
	tx := m.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find role
	var role repositories.Role
	if err := tx.Where("name = ?", roleName).First(&role).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Remove role assignment from database
	if err := m.roleRepo.UnassignRole(ctx, userID, role.ID); err != nil {
		tx.Rollback()
		return err
	}

	// Remove role from Casbin
	if _, err := m.enforcer.DeleteRoleForUser(userID, roleName); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GrantPermission grants a permission to a role
func (m *AuthorizationManager) GrantPermission(ctx context.Context, roleName, resource, action string) error {
	_, err := m.enforcer.AddPolicy(roleName, resource, action)
	return err
}

// RevokePermission revokes a permission from a role
func (m *AuthorizationManager) RevokePermission(ctx context.Context, roleName, resource, action string) error {
	_, err := m.enforcer.RemovePolicy(roleName, resource, action)
	return err
}

// HasPermission checks if a user has a specific permission
func (m *AuthorizationManager) HasPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	return m.enforcer.Enforce(userID, resource, action)
}

// GetUserRoles gets all roles assigned to a user
func (m *AuthorizationManager) GetUserRoles(ctx context.Context, userID string) ([]*repositories.Role, error) {
	return m.roleRepo.GetUserRoles(ctx, userID)
}

// GetRolePermissions gets all permissions assigned to a role
func (m *AuthorizationManager) GetRolePermissions(ctx context.Context, roleName string) ([][]string, error) {
	permissions, err := m.enforcer.GetPermissionsForUser(roleName)
	return permissions, err
}

// HasRole checks if a user has a specific role
func (m *AuthorizationManager) HasRole(ctx context.Context, userID, roleName string) (bool, error) {
	return m.roleRepo.HasRole(ctx, userID, roleName)
}

// GetRole returns a role's stored record, including Metadata — consuming
// apps read this back to check app-defined conventions like a "delegable"
// flag (nebularcore itself doesn't interpret metadata contents).
func (m *AuthorizationManager) GetRole(ctx context.Context, name string) (*repositories.Role, error) {
	return m.roleRepo.GetRoleByName(ctx, name)
}

// ListRoles returns every role.
func (m *AuthorizationManager) ListRoles(ctx context.Context) ([]*repositories.Role, error) {
	return m.roleRepo.ListRoles(ctx)
}

// UpdateRole updates a role's description/metadata. Name is immutable —
// see RoleRepository.UpdateRole's doc comment for why.
func (m *AuthorizationManager) UpdateRole(ctx context.Context, name, description string, metadata map[string]interface{}) error {
	return m.roleRepo.UpdateRole(ctx, name, map[string]interface{}{
		"description": description,
		"metadata":    metadata,
	})
}

// DeleteRole removes a role and cascades: every user's assignment of this
// role in casbin's grouping policy (g), every permission granted to this
// role in casbin's policy store (p), then the roles/role_assignments rows.
// Order matters — the casbin cascade needs the role name, which is still
// resolvable before the DB row is gone.
func (m *AuthorizationManager) DeleteRole(ctx context.Context, name string) error {
	if _, err := m.enforcer.RemoveFilteredGroupingPolicy(1, name); err != nil {
		return fmt.Errorf("failed to remove role assignments from casbin: %w", err)
	}
	if _, err := m.enforcer.RemoveFilteredPolicy(0, name); err != nil {
		return fmt.Errorf("failed to remove role permissions from casbin: %w", err)
	}
	if err := m.roleRepo.DeleteRole(ctx, name); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

// ListPermissionsForRole is a clarifying alias for GetRolePermissions,
// matching the naming used elsewhere in this file's new List* methods.
func (m *AuthorizationManager) ListPermissionsForRole(ctx context.Context, roleName string) ([][]string, error) {
	return m.GetRolePermissions(ctx, roleName)
}

// ListAllPolicies returns every policy line in the p store — every
// role->resource->action grant across all roles.
func (m *AuthorizationManager) ListAllPolicies(ctx context.Context) ([][]string, error) {
	return m.enforcer.GetPolicy()
}

// CreateResource creates a new resource in the catalog.
func (m *AuthorizationManager) CreateResource(ctx context.Context, name, description string, actions []string) error {
	_, err := m.resourceRepo.CreateResource(ctx, map[string]interface{}{
		"name":        name,
		"description": description,
		"actions":     repositories.ResourceActions(actions),
	})
	return err
}

// GetResource returns a resource's stored record.
func (m *AuthorizationManager) GetResource(ctx context.Context, name string) (*repositories.Resource, error) {
	return m.resourceRepo.GetResourceByName(ctx, name)
}

// ListResources returns every resource in the catalog.
func (m *AuthorizationManager) ListResources(ctx context.Context) ([]*repositories.Resource, error) {
	return m.resourceRepo.ListResources(ctx)
}

// UpdateResource updates a resource's description/actions. Name is
// immutable — see ResourceRepository.UpdateResource's doc comment.
func (m *AuthorizationManager) UpdateResource(ctx context.Context, name, description string, actions []string) error {
	return m.resourceRepo.UpdateResource(ctx, name, map[string]interface{}{
		"description": description,
		"actions":     repositories.ResourceActions(actions),
	})
}

// DeleteResource removes a resource from the catalog. Does not cascade any
// casbin p rows referencing this resource name — permissions granted
// against a resource that's since been deleted from the catalog remain
// enforceable until explicitly revoked; the catalog is metadata for the
// dashboard, not the source of truth for what casbin will enforce.
func (m *AuthorizationManager) DeleteResource(ctx context.Context, name string) error {
	return m.resourceRepo.DeleteResource(ctx, name)
}
