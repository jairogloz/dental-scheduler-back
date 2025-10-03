-- Reverse the patient name split by combining first_name and last_name back to name
-- Step 1: Drop the view first before modifying table structure
DROP VIEW IF EXISTS upcoming_appointments;

-- Step 2: Add back the original name column
ALTER TABLE patients 
ADD COLUMN name VARCHAR(255);

-- Step 3: Combine first_name and last_name into the name column
UPDATE patients 
SET name = CASE 
    WHEN first_name IS NOT NULL AND last_name IS NOT NULL AND last_name != '' 
    THEN trim(first_name || ' ' || last_name)
    WHEN first_name IS NOT NULL 
    THEN first_name
    ELSE NULL
END;

-- Step 4: Drop the split columns
ALTER TABLE patients DROP COLUMN IF EXISTS first_name;
ALTER TABLE patients DROP COLUMN IF EXISTS last_name;

-- Step 5: Restore the original upcoming_appointments view
CREATE VIEW upcoming_appointments AS
SELECT 
    a.id,
    p.name as patient_name,
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
