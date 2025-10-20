-- Remove timezone column from clinics table
ALTER TABLE clinics 
DROP COLUMN IF EXISTS timezone;
