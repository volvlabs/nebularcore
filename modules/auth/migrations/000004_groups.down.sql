-- Drop indexes
DROP INDEX IF EXISTS idx_group_permissions_tenant;
DROP INDEX IF EXISTS idx_group_permissions_permission;
DROP INDEX IF EXISTS idx_group_permissions_group;
DROP INDEX IF EXISTS idx_user_groups_tenant;
DROP INDEX IF EXISTS idx_user_groups_group;
DROP INDEX IF EXISTS idx_user_groups_user;
DROP INDEX IF EXISTS idx_permission_groups_tenant;

-- Drop tables
DROP TABLE IF EXISTS group_permissions;
DROP TABLE IF EXISTS user_groups;
DROP TABLE IF EXISTS permission_groups;
