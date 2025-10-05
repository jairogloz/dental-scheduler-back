-- Drop the upcoming_appointments view before modifying the appointments table
DROP VIEW IF EXISTS upcoming_appointments;

-- Reverse: drop service_id column and recreate treatment_type column
DROP INDEX IF EXISTS idx_appointments_service_id;

ALTER TABLE appointments DROP CONSTRAINT IF EXISTS fk_appointments_service;

ALTER TABLE appointments DROP COLUMN IF EXISTS service_id;

ALTER TABLE appointments ADD COLUMN treatment_type TEXT;

-- Recreate the original upcoming_appointments view with treatment_type
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
WHERE a.status = 'scheduled' AND a.start_time > NOW()
ORDER BY a.start_time;
