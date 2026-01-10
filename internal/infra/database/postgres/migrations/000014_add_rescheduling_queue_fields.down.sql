-- Drop index first
DROP INDEX IF EXISTS idx_appointments_rescheduling_queue;

-- Drop columns in reverse order
ALTER TABLE appointments DROP COLUMN IF EXISTS cancellation_reason;
ALTER TABLE appointments DROP COLUMN IF EXISTS rescheduled_to_appointment_id;
ALTER TABLE appointments DROP COLUMN IF EXISTS moved_to_needs_rescheduling_at;
