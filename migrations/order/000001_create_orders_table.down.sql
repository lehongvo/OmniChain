-- Rollback orders table migration
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
DROP TABLE IF EXISTS order_audit_log;
DROP TABLE IF EXISTS orders;

