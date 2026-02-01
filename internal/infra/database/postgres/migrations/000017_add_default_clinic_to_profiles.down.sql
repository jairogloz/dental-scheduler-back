-- Remove default_clinic_id from profiles table

BEGIN;

-- Drop the index
DROP INDEX IF EXISTS idx_profiles_default_clinic_id;

-- Remove the column
ALTER TABLE profiles 
DROP COLUMN IF EXISTS default_clinic_id;

COMMIT;
