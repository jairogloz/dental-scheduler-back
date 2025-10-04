-- Rollback: Remove first_appointment_id column and related objects from patients table

-- Drop the foreign key constraint first
ALTER TABLE patients DROP CONSTRAINT IF EXISTS fk_patients_first_appointment;

-- Drop the index
DROP INDEX IF EXISTS idx_patients_first_appointment_id;

-- Drop the column
ALTER TABLE patients DROP COLUMN IF EXISTS first_appointment_id;
