DROP TRIGGER IF EXISTS update_countries_updated_at ON countries;
DROP FUNCTION IF EXISTS update_countries_updated_at_column();
DROP INDEX IF EXISTS idx_countries_is_active;
DROP TABLE IF EXISTS countries;
