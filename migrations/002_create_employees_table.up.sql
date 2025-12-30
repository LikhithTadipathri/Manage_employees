-- -- Migration: 002_create_employees_table.up.sql
-- -- Description: Create employees table
-- -- Date: 2025-12-04

-- CREATE TABLE IF NOT EXISTS employees (
--     id SERIAL PRIMARY KEY,
--     first_name VARCHAR(100) NOT NULL,
--     last_name VARCHAR(100) NOT NULL,
--     email VARCHAR(100) NOT NULL UNIQUE,
--     phone VARCHAR(20) NOT NULL,
--     position VARCHAR(100) NOT NULL,
--     salary DECIMAL(10, 2) NOT NULL,
--     hired_date TIMESTAMP NOT NULL,
--     created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
-- );

-- -- Add comment to table
-- COMMENT ON TABLE employees IS 'Stores employee information and details';

-- -- Add comments to columns
-- COMMENT ON COLUMN employees.id IS 'Unique employee identifier';
-- COMMENT ON COLUMN employees.first_name IS 'Employee first name';
-- COMMENT ON COLUMN employees.last_name IS 'Employee last name';
-- COMMENT ON COLUMN employees.email IS 'Employee email address (unique)';
-- COMMENT ON COLUMN employees.phone IS 'Employee phone number';
-- COMMENT ON COLUMN employees.position IS 'Job position/title';
-- COMMENT ON COLUMN employees.salary IS 'Annual salary (2 decimal places)';
-- COMMENT ON COLUMN employees.hired_date IS 'Date employee was hired';
-- COMMENT ON COLUMN employees.created_at IS 'Record creation timestamp';
-- COMMENT ON COLUMN employees.updated_at IS 'Record last update timestamp';

-- Migration: 002_create_employees_table.up.sql
-- Description: Create employees table and insert sample employee records
-- Date: 2025-12-04

CREATE TABLE IF NOT EXISTS employees (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    phone VARCHAR(20) NOT NULL,
    position VARCHAR(100) NOT NULL,
    salary DECIMAL(10, 2) NOT NULL,
    hired_date TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add comment to table
COMMENT ON TABLE employees IS 'Stores employee information and details';

-- Add comments to columns
COMMENT ON COLUMN employees.id IS 'Unique employee identifier';
COMMENT ON COLUMN employees.first_name IS 'Employee first name';
COMMENT ON COLUMN employees.last_name IS 'Employee last name';
COMMENT ON COLUMN employees.email IS 'Employee email address (unique)';
COMMENT ON COLUMN employees.phone IS 'Employee phone number';
COMMENT ON COLUMN employees.position IS 'Job position/title';
COMMENT ON COLUMN employees.salary IS 'Annual salary (2 decimal places)';
COMMENT ON COLUMN employees.hired_date IS 'Date employee was hired';
COMMENT ON COLUMN employees.created_at IS 'Record creation timestamp';
COMMENT ON COLUMN employees.updated_at IS 'Record last update timestamp';

-- Insert sample employee records
INSERT INTO employees (first_name, last_name, email, phone, position, salary, hired_date, created_at, updated_at)
VALUES
    ('Priya', 'Sharma', 'priya.sharma@example.com', '+91-9876543210',
     'HR Executive', 480000.00,
     '2025-02-01 09:00:00', '2025-02-01 09:00:00', '2025-02-01 09:00:00'),

    ('Rahul', 'Verma', 'rahul.verma@example.com', '+91-9123456780',
     'Software Engineer', 720000.00,
     '2025-03-10 10:30:00', '2025-03-10 10:30:00', '2025-03-10 10:30:00');
