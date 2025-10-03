-- Remove migration_source_id column from patients table
DROP INDEX IF EXISTS idx_patients_migration_source_id;
ALTER TABLE patients DROP COLUMN IF EXISTS migration_source_id;
