-- Migration: 005_add_user_id_to_employees.up.sql
-- Description: Add user_id column to employees table to link employees to users
-- Date: 2025-12-10

ALTER TABLE employees ADD COLUMN user_id INTEGER;

-- Add foreign key constraint
ALTER TABLE employees ADD CONSTRAINT fk_employees_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_employees_user_id ON employees(user_id);

-- Add unique constraint so each user can have at most one employee record
ALTER TABLE employees ADD CONSTRAINT unique_user_id_employees UNIQUE(user_id);
