package authorization_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volvlabs/nebularcore/modules/auth/authorization"
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

	require.NoError(t, db.Exec(`DROP TABLE IF EXISTS role_assignments, roles, casbin_rule, users CASCADE`).Error)
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

	mgr, err := authorization.NewAuthorizationManager(db)
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
