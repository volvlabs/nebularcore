-- Create roles table
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create role assignments table
CREATE TABLE role_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX idx_role_assignments_user ON role_assignments(user_id);
CREATE INDEX idx_role_assignments_role ON role_assignments(role_id);
CREATE INDEX idx_role_assignments_expires ON role_assignments(expires_at);

-- Create Casbin rules table (if not using gorm-adapter's auto-migration)
CREATE TABLE casbin_rule (
    id BIGSERIAL PRIMARY KEY,
    ptype VARCHAR(255),
    v0 VARCHAR(255),
    v1 VARCHAR(255),
    v2 VARCHAR(255),
    v3 VARCHAR(255),
    v4 VARCHAR(255),
    v5 VARCHAR(255)
);

-- Create indexes for Casbin rules
CREATE INDEX idx_casbin_rule_ptype ON casbin_rule(ptype);
CREATE INDEX idx_casbin_rule_v0 ON casbin_rule(v0);
CREATE INDEX idx_casbin_rule_v1 ON casbin_rule(v1);

-- Create trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_roles_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default roles
INSERT INTO roles (name, description)
VALUES 
    ('admin', 'System administrator with full access'),
    ('user', 'Regular user with basic access'),
    ('guest', 'Guest user with limited access')
ON CONFLICT DO NOTHING;

-- Insert default Casbin rules for admin role
INSERT INTO casbin_rule (ptype, v0, v1, v2)
VALUES 
    ('p', 'admin', '*', '*'),              -- Admin can do anything
    ('p', 'user', 'users', 'read'),        -- Users can read user data
    ('p', 'user', 'profile', '*'),         -- Users can manage their own profile
    ('p', 'guest', 'public', 'read')       -- Guests can read public data
ON CONFLICT DO NOTHING;
