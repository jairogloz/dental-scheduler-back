-- Create patient_organizations junction table to support many-to-many relationship
-- between patients and organizations
CREATE TABLE patient_organizations (
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (patient_id, organization_id)
);

-- Indexes for performance
CREATE INDEX idx_patient_organizations_patient_id ON patient_organizations(patient_id);
CREATE INDEX idx_patient_organizations_organization_id ON patient_organizations(organization_id);

-- Add comments for documentation
COMMENT ON TABLE patient_organizations IS 'Junction table linking patients to organizations they can book appointments with';
COMMENT ON COLUMN patient_organizations.patient_id IS 'Foreign key to patients table';
COMMENT ON COLUMN patient_organizations.organization_id IS 'Foreign key to organizations table';
COMMENT ON COLUMN patient_organizations.created_at IS 'When the patient-organization relationship was created';
COMMENT ON COLUMN patient_organizations.updated_at IS 'When the patient-organization relationship was last updated';
