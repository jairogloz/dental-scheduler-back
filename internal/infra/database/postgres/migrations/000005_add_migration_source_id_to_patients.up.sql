-- Add migration_source_id column to patients table to track original patient IDs from legacy systems
ALTER TABLE patients 
ADD COLUMN migration_source_id VARCHAR(255);

-- Add index for efficient lookups when matching legacy system data
CREATE INDEX idx_patients_migration_source_id ON patients(migration_source_id);

-- Add comment for documentation
COMMENT ON COLUMN patients.migration_source_id IS 'Original patient ID from the legacy system this record was migrated from';
