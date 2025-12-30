-- Migration: 005_add_user_id_to_employees.down.sql
-- Description: Rollback user_id column addition
-- Date: 2025-12-10

-- Drop unique constraint
ALTER TABLE employees DROP CONSTRAINT IF EXISTS unique_user_id_employees;

-- Drop foreign key constraint
ALTER TABLE employees DROP CONSTRAINT IF EXISTS fk_employees_user_id;

-- Drop index
DROP INDEX IF EXISTS idx_employees_user_id;

-- Drop the column
ALTER TABLE employees DROP COLUMN IF EXISTS user_id;
