-- Add color column to doctors table to store hex color codes
ALTER TABLE doctors ADD COLUMN color VARCHAR(7) NOT NULL DEFAULT '#3B82F6';

-- Add a check constraint to ensure valid hex color format
ALTER TABLE doctors ADD CONSTRAINT doctors_color_format_check 
    CHECK (color ~ '^#[0-9A-Fa-f]{6}$');

-- Add comment to document the column
COMMENT ON COLUMN doctors.color IS 'Hex color code for doctor (e.g., #3B82F6 for blue). Used for calendar visualization.';
