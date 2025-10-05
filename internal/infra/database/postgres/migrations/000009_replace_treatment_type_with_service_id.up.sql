-- Drop the upcoming_appointments view before modifying the appointments table
DROP VIEW IF EXISTS upcoming_appointments;

-- Drop treatment_type column and add service_id column to appointments table
ALTER TABLE appointments DROP COLUMN IF EXISTS treatment_type;

ALTER TABLE appointments ADD COLUMN service_id VARCHAR(255);

-- Add foreign key constraint to services table with ON DELETE SET NULL
ALTER TABLE appointments 
    ADD CONSTRAINT fk_appointments_service 
    FOREIGN KEY (service_id) 
    REFERENCES services(id) 
    ON DELETE SET NULL;

-- Create index on service_id for better query performance
CREATE INDEX idx_appointments_service_id ON appointments(service_id);

-- Recreate the upcoming_appointments view with service_id and service_name
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
    a.service_id,
    s.name as service_name,
    a.status
FROM appointments a
JOIN patients p ON a.patient_id = p.id
JOIN doctors d ON a.doctor_id = d.id
JOIN units u ON a.unit_id = u.id
JOIN clinics c ON u.clinic_id = c.id
JOIN organizations o ON c.organization_id = o.id
LEFT JOIN services s ON a.service_id = s.id
WHERE a.status = 'scheduled' AND a.start_time > NOW()
ORDER BY a.start_time;
