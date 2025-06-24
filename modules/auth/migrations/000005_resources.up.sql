-- Create module_resources table
CREATE TABLE IF NOT EXISTS module_resources (
    id UUID PRIMARY KEY,
    module_name VARCHAR(255) NOT NULL,
    resource VARCHAR(255) NOT NULL,
    actions TEXT[] NOT NULL,
    description TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_module_resources_name ON module_resources(module_name);
CREATE INDEX IF NOT EXISTS idx_module_resources_resource ON module_resources(resource);
CREATE INDEX IF NOT EXISTS idx_module_resources_deleted_at ON module_resources(deleted_at);

-- Add trigger for updated_at
CREATE TRIGGER update_module_resources_updated_at
    BEFORE UPDATE ON module_resources
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
