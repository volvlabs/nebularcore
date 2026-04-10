-- Create custom users table with all fields
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE,
    phone_number VARCHAR(255) UNIQUE,
    username VARCHAR(255) UNIQUE,
    password VARCHAR(255),
    metadata JSONB,
    active BOOLEAN DEFAULT true,
    password_reset_token VARCHAR(255),
    token VARCHAR(255),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    last_login_at TIMESTAMP,
    password_reset_at TIMESTAMP,
    deleted_at TIMESTAMP,
    -- Custom fields
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    date_of_birth TIMESTAMP,
    address TEXT,
    company_name VARCHAR(255),
    department VARCHAR(255),
    role VARCHAR(255)
);
