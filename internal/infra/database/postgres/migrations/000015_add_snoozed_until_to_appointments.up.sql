-- Add snoozed_until field to appointments table
-- This allows temporarily hiding appointments from the rescheduling queue
-- until the snooze period expires

ALTER TABLE appointments
ADD COLUMN snoozed_until TIMESTAMPTZ NULL;

-- Create index for efficient queue filtering
-- This index supports the query that filters out snoozed appointments
CREATE INDEX idx_appointments_snoozed_until 
ON appointments(snoozed_until)
WHERE snoozed_until IS NOT NULL;

-- Add comment for documentation
COMMENT ON COLUMN appointments.snoozed_until IS 'Timestamp until which this appointment is snoozed (hidden from rescheduling queue). NULL means not snoozed.';
