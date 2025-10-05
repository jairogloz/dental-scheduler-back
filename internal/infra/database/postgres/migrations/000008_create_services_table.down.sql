-- Rollback: Drop services table and related objects

-- Drop indexes first
DROP INDEX IF EXISTS idx_services_name;
DROP INDEX IF EXISTS idx_services_organization_id;

-- Drop the services table (CASCADE will remove foreign key constraints)
DROP TABLE IF EXISTS services CASCADE;
