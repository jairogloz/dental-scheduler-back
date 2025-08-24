-- Drop trigger and function
DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
DROP FUNCTION IF EXISTS public.handle_new_user();

-- Drop policies
DROP POLICY IF EXISTS "Admins can update organization profiles" ON profiles;
DROP POLICY IF EXISTS "Admins can view organization profiles" ON profiles;
DROP POLICY IF EXISTS "Users can update their own profile" ON profiles;
DROP POLICY IF EXISTS "Users can view their own profile" ON profiles;

-- Disable RLS
ALTER TABLE profiles DISABLE ROW LEVEL SECURITY;

-- Drop triggers
DROP TRIGGER IF EXISTS update_profiles_updated_at ON profiles;

-- Drop indexes
DROP INDEX IF EXISTS idx_patients_user_id;
DROP INDEX IF EXISTS idx_doctors_user_id;
DROP INDEX IF EXISTS idx_profiles_roles;
DROP INDEX IF EXISTS idx_profiles_organization_id;

-- Drop columns (in reverse order)
ALTER TABLE patients DROP COLUMN IF EXISTS user_id;
ALTER TABLE doctors DROP COLUMN IF EXISTS user_id;

-- Drop table
DROP TABLE IF EXISTS profiles;
