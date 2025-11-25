-- Rollback inventory table migration
DROP TRIGGER IF EXISTS increment_inventory_version ON inventory;
DROP TRIGGER IF EXISTS update_inventory_updated_at ON inventory;
DROP FUNCTION IF EXISTS increment_inventory_version();
DROP TABLE IF EXISTS stock_movements;
DROP TABLE IF EXISTS inventory;

