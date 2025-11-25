-- Rollback payments table migration
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
DROP TABLE IF EXISTS refunds;
DROP TABLE IF EXISTS payment_audit_log;
DROP TABLE IF EXISTS payments;

