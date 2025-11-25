-- Create inventory table with optimistic locking
CREATE TABLE inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    store_id UUID,
    -- Stock information
    quantity INTEGER NOT NULL DEFAULT 0,
    reserved_quantity INTEGER NOT NULL DEFAULT 0,
    available_quantity INTEGER GENERATED ALWAYS AS (quantity - reserved_quantity) STORED,
    -- Reorder point
    reorder_point INTEGER DEFAULT 0,
    reorder_quantity INTEGER DEFAULT 0,
    -- Pricing
    cost_price DECIMAL(15,2),
    selling_price DECIMAL(15,2),
    -- Version for optimistic locking
    version INTEGER NOT NULL DEFAULT 1,
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE UNIQUE INDEX idx_inventory_product_store ON inventory(product_id, store_id) WHERE store_id IS NOT NULL;
CREATE UNIQUE INDEX idx_inventory_product_global ON inventory(product_id) WHERE store_id IS NULL;
CREATE INDEX idx_inventory_available_quantity ON inventory(available_quantity) WHERE available_quantity < reorder_point;
CREATE INDEX idx_inventory_store_id ON inventory(store_id) WHERE store_id IS NOT NULL;

-- Stock movement history
CREATE TABLE stock_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inventory_id UUID NOT NULL REFERENCES inventory(id) ON DELETE CASCADE,
    movement_type VARCHAR(50) NOT NULL, -- in, out, adjustment, reserved, released
    quantity INTEGER NOT NULL,
    previous_quantity INTEGER NOT NULL,
    new_quantity INTEGER NOT NULL,
    reason TEXT,
    reference_id UUID, -- order_id, adjustment_id, etc.
    reference_type VARCHAR(50),
    user_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stock_movements_inventory_id ON stock_movements(inventory_id);
CREATE INDEX idx_stock_movements_created_at ON stock_movements(created_at DESC);
CREATE INDEX idx_stock_movements_reference ON stock_movements(reference_type, reference_id) WHERE reference_id IS NOT NULL;

-- Update timestamp trigger
CREATE TRIGGER update_inventory_updated_at BEFORE UPDATE ON inventory
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger to increment version for optimistic locking
CREATE OR REPLACE FUNCTION increment_inventory_version()
RETURNS TRIGGER AS $$
BEGIN
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER increment_inventory_version BEFORE UPDATE ON inventory
    FOR EACH ROW EXECUTE FUNCTION increment_inventory_version();

