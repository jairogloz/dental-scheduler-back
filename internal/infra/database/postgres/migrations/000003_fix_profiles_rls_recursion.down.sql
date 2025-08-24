-- Rollback: Fix infinite recursion in profiles table RLS policies
-- Note: This rollback removes the INSERT policy but does NOT restore the problematic admin policies
-- Admin operations should be handled at the application layer

BEGIN;

-- Remove the INSERT policy (though this is unlikely to cause issues)
DROP POLICY IF EXISTS "Users can insert their own profile" ON profiles;

-- Restore the original trigger function (without explicit roles)
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO public.profiles (id, email, full_name)
    VALUES (
        NEW.id,
        NEW.email,
        COALESCE(NEW.raw_user_meta_data->>'full_name', '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Note: We do NOT recreate the problematic admin policies as they cause infinite recursion
-- Admin access should be implemented at the application layer

COMMIT;
