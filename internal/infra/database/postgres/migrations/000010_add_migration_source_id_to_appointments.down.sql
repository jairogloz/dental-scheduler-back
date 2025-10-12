-- Remove index for migration_source_id
DROP INDEX IF EXISTS idx_appointments_migration_source_id;

-- Remove migration_source_id column from appointments table
ALTER TABLE appointments DROP COLUMN IF EXISTS migration_source_id;
