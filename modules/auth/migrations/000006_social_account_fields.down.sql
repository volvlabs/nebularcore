ALTER TABLE social_accounts
DROP COLUMN IF EXISTS email,
DROP COLUMN IF EXISTS name,
DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_social_accounts_provider_email;
DROP INDEX IF EXISTS idx_social_accounts_deleted_at;
