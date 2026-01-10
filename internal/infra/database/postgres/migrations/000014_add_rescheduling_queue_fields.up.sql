-- Add columns for rescheduling queue tracking
ALTER TABLE appointments ADD COLUMN moved_to_needs_rescheduling_at TIMESTAMPTZ NULL;
ALTER TABLE appointments ADD COLUMN rescheduled_to_appointment_id UUID NULL REFERENCES appointments(id);
ALTER TABLE appointments ADD COLUMN cancellation_reason TEXT NULL;

-- Add composite index for efficient queue queries
CREATE INDEX idx_appointments_rescheduling_queue 
ON appointments (status, moved_to_needs_rescheduling_at)
WHERE status = 'needs-rescheduling';

-- Add comments for clarity
COMMENT ON COLUMN appointments.moved_to_needs_rescheduling_at IS 'Timestamp when appointment status was changed to needs-rescheduling';
COMMENT ON COLUMN appointments.rescheduled_to_appointment_id IS 'Links to new appointment when this one is rescheduled';
COMMENT ON COLUMN appointments.cancellation_reason IS 'Reason provided when appointment is cancelled';
