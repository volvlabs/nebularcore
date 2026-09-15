-- Create tenants table in public schema
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    schema_name VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create index on commonly queried fields
CREATE INDEX IF NOT EXISTS idx_tenants_code ON tenants(code);
CREATE INDEX IF NOT EXISTS idx_tenants_schema_name ON tenants(schema_name);
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for updated_at
CREATE TRIGGER update_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create tenant_settings table in public schema
CREATE TABLE IF NOT EXISTS tenant_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    key VARCHAR(100) NOT NULL,
    value TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, key)
);

-- Create index on tenant_settings
CREATE INDEX IF NOT EXISTS idx_tenant_settings_tenant ON tenant_settings(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_settings_key ON tenant_settings(key);

-- Create trigger for tenant_settings updated_at
CREATE TRIGGER update_tenant_settings_updated_at
    BEFORE UPDATE ON tenant_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create schema_migrations table for tenant module
CREATE TABLE IF NOT EXISTS schema_migrations_tenant (
    version bigint NOT NULL PRIMARY KEY,
    dirty boolean NOT NULL
);

-- Create function to create tenant schema
CREATE OR REPLACE FUNCTION create_tenant_schema(schema_name TEXT)
RETURNS void AS $$
BEGIN
    EXECUTE format('CREATE SCHEMA IF NOT EXISTS %I', schema_name);
END;
$$ LANGUAGE plpgsql;

-- Create function to drop tenant schema
CREATE OR REPLACE FUNCTION drop_tenant_schema(schema_name TEXT)
RETURNS void AS $$
BEGIN
    EXECUTE format('DROP SCHEMA IF EXISTS %I CASCADE', schema_name);
END;
$$ LANGUAGE plpgsql;
