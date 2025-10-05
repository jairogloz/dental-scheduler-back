-- Create services table for clinic services catalog
CREATE TABLE IF NOT EXISTS services (
    id VARCHAR(255) PRIMARY KEY,  -- Custom ID format (e.g., 'srv_brackets_monthly')
    name VARCHAR(255) NOT NULL,
    base_price DECIMAL(10, 2),  -- Nullable to allow services without pricing initially
    organization_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Foreign key constraint
    CONSTRAINT fk_services_organization 
        FOREIGN KEY (organization_id) 
        REFERENCES organizations(id) 
        ON DELETE CASCADE
);

-- Create index on organization_id for fast lookups
CREATE INDEX idx_services_organization_id ON services(organization_id);

-- Create index on name for search/filtering
CREATE INDEX idx_services_name ON services(name);

-- Add comments for documentation
COMMENT ON TABLE services IS 'Catalog of dental services offered by clinics within an organization';
COMMENT ON COLUMN services.id IS 'Custom service identifier with prefix (e.g., srv_brackets_monthly)';
COMMENT ON COLUMN services.name IS 'Display name of the service';
COMMENT ON COLUMN services.base_price IS 'Base price for the service in organization currency';
COMMENT ON COLUMN services.organization_id IS 'Organization that owns this service';
