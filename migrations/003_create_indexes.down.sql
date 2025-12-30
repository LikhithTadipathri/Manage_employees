-- Migration: 003_create_indexes.down.sql
-- Description: Drop indexes (rollback)
-- Date: 2025-12-04

-- Drop users table indexes
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_created_at;

-- Drop employees table indexes
DROP INDEX IF EXISTS idx_employees_email;
DROP INDEX IF EXISTS idx_employees_created_at;
DROP INDEX IF EXISTS idx_employees_position;
DROP INDEX IF EXISTS idx_employees_hired_date;

-- Drop composite indexes
DROP INDEX IF EXISTS idx_employees_name;
DROP INDEX IF EXISTS idx_employees_salary_date;
