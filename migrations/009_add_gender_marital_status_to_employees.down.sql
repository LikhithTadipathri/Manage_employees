-- Remove Gender and MaritalStatus columns from employees table
ALTER TABLE employees DROP CONSTRAINT IF EXISTS check_gender;
ALTER TABLE employees DROP COLUMN IF EXISTS gender;
ALTER TABLE employees DROP COLUMN IF EXISTS marital_status;
