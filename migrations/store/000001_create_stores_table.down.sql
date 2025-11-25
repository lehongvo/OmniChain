-- Rollback stores table migration
DROP TRIGGER IF EXISTS update_stores_updated_at ON stores;
DROP TABLE IF EXISTS stores;

