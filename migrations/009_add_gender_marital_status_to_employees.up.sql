-- Add Gender and MaritalStatus columns to employees table
ALTER TABLE employees ADD COLUMN gender VARCHAR(10);
ALTER TABLE employees ADD COLUMN marital_status BOOLEAN DEFAULT FALSE; -- TRUE = Married, FALSE = Not Married

-- Add constraint for valid genders
ALTER TABLE employees ADD CONSTRAINT check_gender CHECK (gender IN ('Male', 'Female'));

-- Update timestamp
ALTER TABLE employees ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;
