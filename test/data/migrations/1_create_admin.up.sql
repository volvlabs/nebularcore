CREATE TABLE IF NOT EXISTS admins (
    id TEXT PRIMARY KEY,
    created TIMESTAMP,
    updated TIMESTAMP,
    avatar TEXT,
    first_name TEXT,
    last_name TEXT,
    email TEXT UNIQUE,
    role TEXT,
    token_key TEXT,
    password_hash TEXT,
    is_deleted BOOLEAN DEFAULT false,
    deleted_at TIMESTAMP
);
