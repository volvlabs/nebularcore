-- Drop triggers
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS role_assignments;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS casbin_rule;
