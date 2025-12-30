-- Add gender and marital_status columns to employees table if they don't exist
ALTER TABLE employees ADD COLUMN IF NOT EXISTS gender VARCHAR(10);
ALTER TABLE employees ADD COLUMN IF NOT EXISTS marital_status BOOLEAN DEFAULT FALSE;

-- Add constraint for valid genders (ignore if already exists)
DO $$
BEGIN
    ALTER TABLE employees ADD CONSTRAINT check_gender_valid CHECK (gender IN ('Male', 'Female'));
EXCEPTION WHEN OTHERS THEN
    NULL; -- Constraint might already exist
END $$;
