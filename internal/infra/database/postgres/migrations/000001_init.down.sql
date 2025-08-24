-- Drop all triggers
DROP TRIGGER IF EXISTS update_doctor_availability_updated_at ON doctor_availability;
DROP TRIGGER IF EXISTS update_appointments_updated_at ON appointments;
DROP TRIGGER IF EXISTS update_patients_updated_at ON patients;
DROP TRIGGER IF EXISTS update_doctors_updated_at ON doctors;
DROP TRIGGER IF EXISTS update_units_updated_at ON units;
DROP TRIGGER IF EXISTS update_clinics_updated_at ON clinics;
DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop views
DROP VIEW IF EXISTS upcoming_appointments;

-- Drop policies
DROP POLICY IF EXISTS "Appointments are viewable by authenticated users only" ON appointments;

-- Disable RLS
ALTER TABLE appointments DISABLE ROW LEVEL SECURITY;

-- Drop constraints
ALTER TABLE appointments DROP CONSTRAINT IF EXISTS check_appointment_times;

-- Drop indexes
DROP INDEX IF EXISTS idx_doctor_availability_doctor_id;
DROP INDEX IF EXISTS idx_appointments_unit_id;
DROP INDEX IF EXISTS idx_appointments_doctor_id;
DROP INDEX IF EXISTS idx_appointments_start_time;
DROP INDEX IF EXISTS idx_doctors_default_unit_id;
DROP INDEX IF EXISTS idx_doctors_organization_id;
DROP INDEX IF EXISTS idx_units_clinic_id;
DROP INDEX IF EXISTS idx_clinics_organization_id;

-- Drop tables (in reverse order due to foreign key constraints)
DROP TABLE IF EXISTS doctor_availability;
DROP TABLE IF EXISTS appointments;
DROP TABLE IF EXISTS patients;
DROP TABLE IF EXISTS doctors;
DROP TABLE IF EXISTS units;
DROP TABLE IF EXISTS clinics;
DROP TABLE IF EXISTS organizations;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";
