-- Add index to speed up status-based appointment queries
CREATE INDEX IF NOT EXISTS idx_appointments_status ON appointments(status);
