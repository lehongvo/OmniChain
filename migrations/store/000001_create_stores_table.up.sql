-- Create stores table
CREATE TABLE stores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    -- Location
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100),
    postal_code VARCHAR(20),
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    -- Contact
    phone VARCHAR(20),
    email VARCHAR(255),
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    -- Metadata
    settings JSONB DEFAULT '{}',
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_stores_code ON stores(code) WHERE deleted_at IS NULL;
CREATE INDEX idx_stores_is_active ON stores(is_active) WHERE deleted_at IS NULL AND is_active = TRUE;
CREATE INDEX idx_stores_location ON stores(latitude, longitude) WHERE deleted_at IS NULL;

-- Geospatial index for location queries (requires PostGIS extension)
-- CREATE INDEX idx_stores_location_gist ON stores USING GIST(ST_MakePoint(longitude, latitude));

-- Update timestamp trigger
CREATE TRIGGER update_stores_updated_at BEFORE UPDATE ON stores
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

