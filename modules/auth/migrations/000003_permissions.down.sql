-- Drop triggers
DROP TRIGGER IF EXISTS update_permissions_updated_at ON permissions;
DROP TRIGGER IF EXISTS update_permission_groups_updated_at ON permission_groups;

-- Drop user group associations
DROP TABLE IF EXISTS user_groups;

-- Drop group permissions
DROP TABLE IF EXISTS group_permissions;

-- Drop permission groups
DROP TABLE IF EXISTS permission_groups;

-- Drop user direct permissions
DROP TABLE IF EXISTS user_permissions;

-- Drop role permissions
DROP TABLE IF EXISTS role_permissions;

-- Drop base permissions table
DROP TABLE IF EXISTS permissions;
