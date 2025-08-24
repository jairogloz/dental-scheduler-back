-- Migration: Fix infinite recursion in profiles table RLS policies
-- Date: 2025-08-24
-- Issue: Infinite recursion detected in policy for relation 'profiles'
-- Cause: Admin policy querying same table it protects

BEGIN;

-- Drop the problematic admin policies that cause infinite recursion
DROP POLICY IF EXISTS "Admins can view organization profiles" ON profiles;
DROP POLICY IF EXISTS "Admins can update organization profiles" ON profiles;
DROP POLICY IF EXISTS "Admins can view all profiles" ON profiles;

-- Ensure the INSERT policy exists for user profile creation
-- (This may already exist, but we're being safe)
DROP POLICY IF EXISTS "Users can insert their own profile" ON profiles;
CREATE POLICY "Users can insert their own profile" 
ON profiles FOR INSERT 
TO authenticated 
WITH CHECK (auth.uid() = id);

-- Update the trigger function to ensure default roles are set
-- This replaces any existing version to ensure consistency
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO public.profiles (id, email, full_name, roles)
    VALUES (
        NEW.id,
        NEW.email,
        COALESCE(NEW.raw_user_meta_data->>'full_name', ''),
        ARRAY['receptionist']  -- Default role for new users
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Ensure the trigger exists (recreate to be safe)
DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
CREATE TRIGGER on_auth_user_created
    AFTER INSERT ON auth.users
    FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

COMMIT;
