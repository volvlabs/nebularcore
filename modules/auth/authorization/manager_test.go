package authorization_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volvlabs/nebularcore/modules/auth/authorization"
	"github.com/volvlabs/nebularcore/modules/auth/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewAuthorizationManager previously pointed casbin at a relative file path
// ("auth_model.conf") that didn't exist anywhere in this module — it would
// have failed the instant anything actually constructed one. This test
// exercises the real flow (create role, assign it, grant a permission,
// enforce it) against real Postgres, since the underlying casbin gorm
// adapter and this package's own Role/RoleAssignment models use jsonb
// (Postgres-only), which rules out a SQLite fallback here.
func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	host := envOr("TEST_DATABASE_HOST", "localhost")
	port := envOr("TEST_DATABASE_PORT", "15433")
	dsn := fmt.Sprintf("host=%s user=postgres password=test dbname=auth_test port=%s sslmode=disable", host, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("no reachable test postgres at %s:%s (%v) — skipping authorization manager tests", host, port, err)
	}

	require.NoError(t, db.Exec(`DROP TABLE IF EXISTS role_assignments, roles, resources, casbin_rule, users CASCADE`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE users (id UUID PRIMARY KEY DEFAULT gen_random_uuid())`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE roles (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL UNIQUE,
			description TEXT,
			metadata JSONB,
			created_at TIMESTAMPTZ DEFAULT now(),
			updated_at TIMESTAMPTZ DEFAULT now(),
			deleted_at TIMESTAMPTZ
		)`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE role_assignments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id),
			role_id UUID NOT NULL REFERENCES roles(id),
			created_at TIMESTAMPTZ DEFAULT now(),
			expires_at TIMESTAMPTZ
		)`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE resources (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL UNIQUE,
			description TEXT,
			actions JSONB,
			created_at TIMESTAMPTZ DEFAULT now(),
			updated_at TIMESTAMPTZ DEFAULT now(),
			deleted_at TIMESTAMPTZ
		)`).Error)

	return db
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func TestAuthorizationManager_EndToEnd(t *testing.T) {
	db := setupDB(t)

	mgr, err := authorization.NewAuthorizationManager(db, config.AuthorizationConfig{})
	require.NoError(t, err, "constructing the manager must not fail (previously failed: missing auth_model.conf)")

	ctx := context.Background()
	require.NoError(t, mgr.CreateRole(ctx, "farm_manager", "manages a farm", nil))

	var userID string
	require.NoError(t, db.Raw("INSERT INTO users DEFAULT VALUES RETURNING id").Scan(&userID).Error)

	require.NoError(t, mgr.AssignRole(ctx, userID, "farm_manager", nil))

	hasRole, err := mgr.HasRole(ctx, userID, "farm_manager")
	require.NoError(t, err)
	require.True(t, hasRole)

	allowedBefore, err := mgr.HasPermission(ctx, userID, "zones", "write")
	require.NoError(t, err)
	require.False(t, allowedBefore, "no permission granted yet — must not be allowed")

	require.NoError(t, mgr.GrantPermission(ctx, "farm_manager", "zones", "write"))

	allowedAfter, err := mgr.HasPermission(ctx, userID, "zones", "write")
	require.NoError(t, err)
	require.True(t, allowedAfter, "farm_manager was granted zones:write")

	deniedAction, err := mgr.HasPermission(ctx, userID, "zones", "delete")
	require.NoError(t, err)
	require.False(t, deniedAction, "farm_manager was never granted zones:delete")

	require.NoError(t, mgr.UnassignRole(ctx, userID, "farm_manager"))
	revoked, err := mgr.HasPermission(ctx, userID, "zones", "write")
	require.NoError(t, err)
	require.False(t, revoked, "permission must be gone once the role is unassigned")
}

func TestAuthorizationManager_RoleCRUD(t *testing.T) {
	db := setupDB(t)
	mgr, err := authorization.NewAuthorizationManager(db, config.AuthorizationConfig{})
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, mgr.CreateRole(ctx, "consultant", "external advisor", map[string]interface{}{"delegable": false}))

	role, err := mgr.GetRole(ctx, "consultant")
	require.NoError(t, err)
	require.Equal(t, "external advisor", role.Description)
	require.Equal(t, false, role.Metadata["delegable"])

	roles, err := mgr.ListRoles(ctx)
	require.NoError(t, err)
	require.Len(t, roles, 1)
	require.Equal(t, "consultant", roles[0].Name)

	require.NoError(t, mgr.UpdateRole(ctx, "consultant", "updated description", map[string]interface{}{"delegable": true}))
	updated, err := mgr.GetRole(ctx, "consultant")
	require.NoError(t, err)
	require.Equal(t, "updated description", updated.Description)
	require.Equal(t, true, updated.Metadata["delegable"])

	var userID string
	require.NoError(t, db.Raw("INSERT INTO users DEFAULT VALUES RETURNING id").Scan(&userID).Error)
	require.NoError(t, mgr.AssignRole(ctx, userID, "consultant", nil))
	require.NoError(t, mgr.GrantPermission(ctx, "consultant", "reports", "read"))

	require.NoError(t, mgr.DeleteRole(ctx, "consultant"))

	_, err = mgr.GetRole(ctx, "consultant")
	require.Error(t, err, "role must be gone after DeleteRole")

	hasRole, err := mgr.HasRole(ctx, userID, "consultant")
	require.NoError(t, err)
	require.False(t, hasRole, "casbin role assignment must be cascaded away by DeleteRole")

	allowed, err := mgr.HasPermission(ctx, userID, "reports", "read")
	require.NoError(t, err)
	require.False(t, allowed, "casbin permission grant must be cascaded away by DeleteRole")
}

func TestAuthorizationManager_ResourceCRUD(t *testing.T) {
	db := setupDB(t)
	mgr, err := authorization.NewAuthorizationManager(db, config.AuthorizationConfig{})
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, mgr.CreateResource(ctx, "zones", "farm zones", []string{"create", "read", "delete"}))

	resource, err := mgr.GetResource(ctx, "zones")
	require.NoError(t, err)
	require.Equal(t, "farm zones", resource.Description)
	require.ElementsMatch(t, []string{"create", "read", "delete"}, []string(resource.Actions))

	resources, err := mgr.ListResources(ctx)
	require.NoError(t, err)
	require.Len(t, resources, 1)
	require.Equal(t, "zones", resources[0].Name)

	require.NoError(t, mgr.UpdateResource(ctx, "zones", "updated description", []string{"read"}))
	updated, err := mgr.GetResource(ctx, "zones")
	require.NoError(t, err)
	require.Equal(t, "updated description", updated.Description)
	require.ElementsMatch(t, []string{"read"}, []string(updated.Actions))

	require.NoError(t, mgr.DeleteResource(ctx, "zones"))
	_, err = mgr.GetResource(ctx, "zones")
	require.Error(t, err, "resource must be gone after DeleteResource")
}
