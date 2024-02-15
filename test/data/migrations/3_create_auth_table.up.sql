CREATE TABLE IF NOT EXISTS auths (
    id TEXT PRIMARY KEY,
    created TIMESTAMP,
    updated TIMESTAMP,
    identity TEXT,
    role TEXT,
    password_hash TEXT
);