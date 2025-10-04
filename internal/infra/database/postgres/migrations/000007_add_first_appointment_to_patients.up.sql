-- Add first_appointment_id to patients to mark the first appointment for the patient
-- This helps identify when a patient is visiting for the first time

ALTER TABLE patients
  ADD COLUMN IF NOT EXISTS first_appointment_id uuid NULL;

-- Add foreign key constraint to ensure referential integrity
ALTER TABLE patients
  ADD CONSTRAINT fk_patients_first_appointment 
  FOREIGN KEY (first_appointment_id) 
  REFERENCES appointments(id) 
  ON DELETE SET NULL;

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_patients_first_appointment_id ON patients(first_appointment_id);

-- Add comment for documentation
COMMENT ON COLUMN patients.first_appointment_id IS 'ID of the first appointment created for this patient, used to identify first-time visits';
