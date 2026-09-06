-- Drop triggers first
DROP TRIGGER IF EXISTS update_tenant_settings_updated_at ON tenant_settings;
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;

-- Drop tenant-related tables
DROP TABLE IF EXISTS tenant_settings;
DROP TABLE IF EXISTS tenants;

-- Drop schema_migrations table for tenant module
DROP TABLE IF EXISTS schema_migrations_tenant;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS create_tenant_schema(TEXT);
DROP FUNCTION IF EXISTS drop_tenant_schema(TEXT);
