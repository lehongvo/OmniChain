-- Create orders table with encryption support
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    store_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    -- Encrypted payment information
    encrypted_payment_data BYTEA,
    -- Order metadata
    items JSONB NOT NULL,
    shipping_address JSONB,
    billing_address JSONB,
    notes TEXT,
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    cancelled_at TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_orders_user_id ON orders(user_id) WHERE cancelled_at IS NULL;
CREATE INDEX idx_orders_store_id ON orders(store_id) WHERE cancelled_at IS NULL;
CREATE INDEX idx_orders_status ON orders(status) WHERE cancelled_at IS NULL;
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX idx_orders_user_status ON orders(user_id, status) WHERE cancelled_at IS NULL;

-- GIN index for JSONB queries
CREATE INDEX idx_orders_items_gin ON orders USING GIN(items);

-- Audit trail table
CREATE TABLE order_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    old_status VARCHAR(50),
    new_status VARCHAR(50),
    user_id UUID,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_audit_log_order_id ON order_audit_log(order_id);
CREATE INDEX idx_order_audit_log_created_at ON order_audit_log(created_at);

-- Update timestamp trigger
CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

