-- Migration: 003_create_indexes.up.sql
-- Description: Create indexes for performance optimization
-- Date: 2025-12-04

-- Users table indexes
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- Employees table indexes
CREATE INDEX IF NOT EXISTS idx_employees_email ON employees(email);
CREATE INDEX IF NOT EXISTS idx_employees_created_at ON employees(created_at);
CREATE INDEX IF NOT EXISTS idx_employees_position ON employees(position);
CREATE INDEX IF NOT EXISTS idx_employees_hired_date ON employees(hired_date);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_employees_name ON employees(last_name, first_name);
CREATE INDEX IF NOT EXISTS idx_employees_salary_date ON employees(salary, hired_date);
