-- Add profiles table for user roles and additional information
CREATE TABLE profiles (
    id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    roles TEXT[] NOT NULL DEFAULT ARRAY['receptionist'], -- Array of roles: admin, doctor, receptionist, patient, dev
    avatar_url VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Add user_id column to existing tables to link with auth users (optional)
ALTER TABLE doctors 
ADD COLUMN user_id UUID REFERENCES auth.users(id);

ALTER TABLE patients 
ADD COLUMN user_id UUID REFERENCES auth.users(id);

-- Create indexes for better performance
CREATE INDEX idx_profiles_organization_id ON profiles(organization_id);
CREATE INDEX idx_profiles_roles ON profiles USING GIN(roles);
CREATE INDEX idx_doctors_user_id ON doctors(user_id);
CREATE INDEX idx_patients_user_id ON patients(user_id);

-- Add trigger for profiles updated_at
CREATE TRIGGER update_profiles_updated_at
    BEFORE UPDATE ON profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Row Level Security for profiles
ALTER TABLE profiles ENABLE ROW LEVEL SECURITY;

-- Policies for profiles table
CREATE POLICY "Users can view their own profile" 
ON profiles FOR SELECT 
TO authenticated 
USING (auth.uid() = id);

CREATE POLICY "Users can update their own profile" 
ON profiles FOR UPDATE 
TO authenticated 
USING (auth.uid() = id);

-- Admins can view all profiles within their organization
CREATE POLICY "Admins can view organization profiles" 
ON profiles FOR SELECT 
TO authenticated 
USING (
    EXISTS (
        SELECT 1 FROM profiles p
        WHERE p.id = auth.uid() 
        AND 'admin' = ANY(p.roles)
        AND p.organization_id = profiles.organization_id
    )
);

-- Admins can update profiles within their organization
CREATE POLICY "Admins can update organization profiles" 
ON profiles FOR UPDATE 
TO authenticated 
USING (
    EXISTS (
        SELECT 1 FROM profiles p
        WHERE p.id = auth.uid() 
        AND 'admin' = ANY(p.roles)
        AND p.organization_id = profiles.organization_id
    )
);

-- Function to automatically create a profile when a user signs up
-- Note: organization_id will need to be set separately after signup
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

-- Trigger to automatically create profile on user creation
CREATE TRIGGER on_auth_user_created
    AFTER INSERT ON auth.users
    FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();
