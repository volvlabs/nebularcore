package repositories_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	migrationRunner "github.com/volvlabs/nebularcore/core/migration_runner"
	"github.com/volvlabs/nebularcore/modules/auth"
	"github.com/volvlabs/nebularcore/modules/auth/factories"
	"github.com/volvlabs/nebularcore/modules/auth/repositories"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	host := envOr("TEST_DATABASE_HOST", "localhost")
	port := envOr("TEST_DATABASE_PORT", "15433")
	dsn := fmt.Sprintf("host=%s user=postgres password=test dbname=auth_test port=%s sslmode=disable", host, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("no reachable test postgres at %s:%s (%v) — skipping user repository tests", host, port, err)
	}

	// Other test packages share this database and may have left a
	// differently-shaped users table (e.g. authorization's manager_test.go
	// creates a minimal stand-in) — CREATE TABLE IF NOT EXISTS in the real
	// migrations would then silently skip past it. Force a clean slate.
	require.NoError(t, db.Exec(`DROP TABLE IF EXISTS
		user_groups, group_permissions, permission_groups,
		user_permissions, role_permissions, permissions,
		role_assignments, roles, casbin_rule,
		refresh_tokens, api_credentials, social_accounts, users
		CASCADE`).Error)
	require.NoError(t, db.Exec(`DROP TABLE IF EXISTS schema_migrations_auth`).Error)

	sources := auth.New(nil).GetMigrationSources("")
	connString := fmt.Sprintf("postgres://postgres:test@%s:%s/auth_test?sslmode=disable", host, port)
	runner, err := migrationRunner.New(sources, connString, "schema_migrations_auth")
	require.NoError(t, err)
	require.NoError(t, runner.Up())
	runner.Close()

	return db
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// This is the actual signup path's repository (backends.NewLocalBackend
// creates users through the same Create call), not just a bootstrap-script
// edge case — a caller getting back a zero-UUID user after Create() would
// silently break anything downstream that needs the real ID (e.g. issuing
// a JWT with sub="00000000-0000-0000-0000-000000000000", or immediately
// assigning a role by that ID).
func TestUserRepository_Create_ReturnsTheRealGeneratedID(t *testing.T) {
	db := setupDB(t)
	repo := repositories.NewUserRepository(db, factories.NewDefaultUserFactory())

	user, err := repo.Create(context.Background(), map[string]any{
		"email":    "created@example.com",
		"password": "irrelevant-for-this-test",
		"active":   true,
	})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, user.GetID(), "Create must return the DB-generated ID, not the zero value")

	// Confirm it's not just non-zero by coincidence — it must be the row
	// that's actually in the database.
	fetched, err := repo.FindByID(context.Background(), user.GetID())
	require.NoError(t, err)
	require.Equal(t, "created@example.com", fetched.GetEmail())
}
