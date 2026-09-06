-- Drop indexes first
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_phone;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_social_accounts_user;

-- Drop tables in reverse order (to handle foreign key dependencies)
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS user_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS social_accounts;
DROP TABLE IF EXISTS users;
