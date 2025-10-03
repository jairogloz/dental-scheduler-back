-- Split patient name into first_name and last_name columns
-- Step 1: Add the new last_name column
ALTER TABLE patients 
ADD COLUMN last_name VARCHAR(255);

-- Step 2: Rename the existing name column to first_name
ALTER TABLE patients 
RENAME COLUMN name TO first_name;

-- Step 3: Update existing data by splitting the first_name into first and last names
-- This attempts to split on the last space in the name
UPDATE patients 
SET 
    last_name = CASE 
        WHEN position(' ' in first_name) > 0 
        THEN trim(substring(first_name from position(' ' in reverse(first_name)) + 1))
        ELSE ''
    END,
    first_name = CASE 
        WHEN position(' ' in first_name) > 0 
        THEN trim(substring(first_name from 1 for length(first_name) - position(' ' in reverse(first_name))))
        ELSE first_name
    END
WHERE first_name IS NOT NULL;

-- Step 4: Update the upcoming_appointments view to use the new column structure
DROP VIEW IF EXISTS upcoming_appointments;
CREATE VIEW upcoming_appointments AS
SELECT 
    a.id,
    CASE 
        WHEN p.last_name IS NOT NULL AND p.last_name != '' 
        THEN trim(p.first_name || ' ' || p.last_name)
        ELSE p.first_name
    END as patient_name,
    d.name as doctor_name,
    u.name as unit_name,
    c.name as clinic_name,
    o.name as organization_name,
    a.start_time,
    a.end_time,
    a.treatment_type,
    a.status
FROM appointments a
JOIN patients p ON a.patient_id = p.id
JOIN doctors d ON a.doctor_id = d.id
JOIN units u ON a.unit_id = u.id
JOIN clinics c ON u.clinic_id = c.id
JOIN organizations o ON c.organization_id = o.id
WHERE a.start_time > NOW()
AND a.status = 'scheduled'
ORDER BY a.start_time;

-- Add comments for documentation
COMMENT ON COLUMN patients.first_name IS 'Patient first name';
COMMENT ON COLUMN patients.last_name IS 'Patient last name';
