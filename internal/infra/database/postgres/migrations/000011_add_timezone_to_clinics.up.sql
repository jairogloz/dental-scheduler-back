-- Add timezone column to clinics table
-- This stores IANA timezone strings (e.g., 'America/New_York', 'Europe/London', 'UTC')
ALTER TABLE clinics 
ADD COLUMN timezone VARCHAR(100) NOT NULL DEFAULT 'UTC';

-- Add comment for documentation
COMMENT ON COLUMN clinics.timezone IS 'IANA timezone identifier for the clinic location (e.g., America/New_York)';
