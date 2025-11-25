-- Create payments table (PCI-DSS compliant)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    user_id UUID NOT NULL,
    -- Payment method (tokenized, no raw card data)
    payment_method_token VARCHAR(255) NOT NULL,
    payment_method_type VARCHAR(50) NOT NULL, -- card, bank_transfer, etc.
    -- Amount
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    -- Status
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed, refunded
    -- Encrypted sensitive data (PCI-DSS compliant)
    encrypted_data BYTEA,
    -- External payment provider
    provider VARCHAR(50), -- stripe, paypal, etc.
    provider_transaction_id VARCHAR(255),
    provider_response JSONB,
    -- 3D Secure
    three_d_secure_enabled BOOLEAN DEFAULT FALSE,
    three_d_secure_status VARCHAR(50),
    -- Fraud detection
    fraud_score DECIMAL(5,2),
    fraud_flagged BOOLEAN DEFAULT FALSE,
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP,
    completed_at TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_provider_transaction_id ON payments(provider_transaction_id) WHERE provider_transaction_id IS NOT NULL;
CREATE INDEX idx_payments_created_at ON payments(created_at DESC);

-- Payment audit trail (PCI-DSS requirement)
CREATE TABLE payment_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    old_status VARCHAR(50),
    new_status VARCHAR(50),
    ip_address INET,
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payment_audit_log_payment_id ON payment_audit_log(payment_id);
CREATE INDEX idx_payment_audit_log_created_at ON payment_audit_log(created_at);

-- Refunds table
CREATE TABLE refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL REFERENCES payments(id),
    amount DECIMAL(15,2) NOT NULL,
    reason TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    provider_refund_id VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP
);

CREATE INDEX idx_refunds_payment_id ON refunds(payment_id);

-- Update timestamp trigger
CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

