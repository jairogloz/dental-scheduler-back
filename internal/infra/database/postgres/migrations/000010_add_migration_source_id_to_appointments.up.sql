-- Add migration_source_id column to appointments table
-- This field stores the ID of the appointment in the legacy system
ALTER TABLE appointments ADD COLUMN migration_source_id VARCHAR(255);

-- Add index for faster lookups by migration_source_id
CREATE INDEX idx_appointments_migration_source_id ON appointments(migration_source_id);
