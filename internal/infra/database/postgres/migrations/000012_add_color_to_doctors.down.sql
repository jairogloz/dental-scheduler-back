-- Remove color column from doctors table
ALTER TABLE doctors DROP CONSTRAINT IF EXISTS doctors_color_format_check;
ALTER TABLE doctors DROP COLUMN IF EXISTS color;
