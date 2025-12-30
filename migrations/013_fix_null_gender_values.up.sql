-- Update existing NULL gender values to 'Male' as default
UPDATE employees SET gender = 'Male' WHERE gender IS NULL OR gender = '';

-- Now apply the NOT NULL constraint (if not already there)
-- This varies by database type
