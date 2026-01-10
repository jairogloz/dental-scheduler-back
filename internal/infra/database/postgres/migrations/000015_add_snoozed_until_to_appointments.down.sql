-- Rollback: Remove snoozed_until field and related objects

-- Drop index first
DROP INDEX IF EXISTS idx_appointments_snoozed_until;

-- Remove column
ALTER TABLE appointments
DROP COLUMN IF EXISTS snoozed_until;
