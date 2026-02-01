-- Add default_clinic_id to profiles table
-- This allows users to have a default clinic for operations like cash sessions and payments

BEGIN;

-- Add default_clinic_id column
ALTER TABLE profiles 
ADD COLUMN default_clinic_id UUID REFERENCES clinics(id) ON DELETE SET NULL;

-- Add index for better query performance
CREATE INDEX idx_profiles_default_clinic_id ON profiles(default_clinic_id);

-- Add comment explaining the field
COMMENT ON COLUMN profiles.default_clinic_id IS 'Default clinic for user operations (cash sessions, payments, etc.). Must belong to user''s organization.';

COMMIT;
