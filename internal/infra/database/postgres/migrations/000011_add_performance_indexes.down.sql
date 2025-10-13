-- Rollback migration 000011: Remove performance indexes
-- This removes all the indexes added in the up migration

-- Remove performance indexes in reverse order
DROP INDEX IF EXISTS idx_patients_name_sort;
DROP INDEX IF EXISTS idx_patients_names;
DROP INDEX IF EXISTS idx_patient_orgs_composite;
DROP INDEX IF EXISTS idx_services_name_org;
DROP INDEX IF EXISTS idx_units_clinic_org;
DROP INDEX IF EXISTS idx_appointments_unit_migration;
DROP INDEX IF EXISTS idx_appointments_migration_source_id_unique;