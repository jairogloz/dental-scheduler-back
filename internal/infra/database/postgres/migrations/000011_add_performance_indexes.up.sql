-- Add performance-critical indexes for optimal query performance
-- Migration 000011: Add missing performance indexes

-- Critical for UPDATE operations (used in WHERE clause)
-- Make migration_source_id unique where it's not null
CREATE UNIQUE INDEX idx_appointments_migration_source_id_unique 
ON appointments (migration_source_id) 
WHERE migration_source_id IS NOT NULL;

-- Composite index for the main query optimization
-- Optimizes queries filtering by unit_id and migration_source_id together
CREATE INDEX idx_appointments_unit_migration 
ON appointments (unit_id, migration_source_id) 
WHERE migration_source_id IS NOT NULL;

-- Composite index for better join performance in units table
-- INCLUDE clause adds non-key columns for index-only scans
CREATE INDEX idx_units_clinic_org 
ON units (clinic_id) 
INCLUDE (id);

-- For name lookups with organization filtering
-- Optimizes queries searching services by name within an organization
CREATE INDEX idx_services_name_org 
ON services (organization_id, name);

-- Composite index for optimal join performance in patient_organizations
-- Optimizes the common query pattern of finding patients by organization
CREATE INDEX idx_patient_orgs_composite 
ON patient_organizations (organization_id, patient_id);

-- For name-based searches and sorting on patients
-- Optimizes queries searching patients by first_name and/or last_name
CREATE INDEX idx_patients_names 
ON patients (first_name, last_name);

-- For the LoadPatients ORDER BY clause with ID for tie-breaking
-- Optimizes sorting patients by name with consistent ordering
CREATE INDEX idx_patients_name_sort 
ON patients (first_name, last_name, id);

-- Add comments for documentation
COMMENT ON INDEX idx_appointments_migration_source_id_unique IS 'Ensures migration_source_id uniqueness for data integrity';
COMMENT ON INDEX idx_appointments_unit_migration IS 'Optimizes queries filtering by unit and migration source ID';
COMMENT ON INDEX idx_units_clinic_org IS 'Optimizes unit-clinic joins with index-only scans';
COMMENT ON INDEX idx_services_name_org IS 'Optimizes service searches by name within organization';
COMMENT ON INDEX idx_patient_orgs_composite IS 'Optimizes patient-organization relationship queries';
COMMENT ON INDEX idx_patients_names IS 'Optimizes patient searches by name';
COMMENT ON INDEX idx_patients_name_sort IS 'Optimizes patient sorting with consistent ordering';