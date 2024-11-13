CREATE TABLE IF NOT EXISTS auths (
    id TEXT PRIMARY KEY,
    created TIMESTAMP,
    updated TIMESTAMP,
    identity TEXT,
    user_table_name TEXT,
    user_id TEXT UNIQUE,
    role TEXT,
    password_hash TEXT
);