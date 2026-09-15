ALTER TABLE social_accounts 
ADD COLUMN IF NOT EXISTS email VARCHAR(80),
ADD COLUMN IF NOT EXISTS name VARCHAR(255),
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

CREATE UNIQUE INDEX IF NOT EXISTS idx_social_accounts_provider_email ON social_accounts(provider, email);
CREATE INDEX IF NOT EXISTS idx_social_accounts_deleted_at ON social_accounts(deleted_at);
