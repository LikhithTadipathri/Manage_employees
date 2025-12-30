-- Delete all records to start fresh
-- Order matters: delete child tables first, then parent tables

-- Disable foreign key constraints temporarily (PostgreSQL)
-- This ensures we can delete everything

-- Delete all leave requests first (they reference employees)
DELETE FROM leave_requests;

-- Delete all employees (they reference users)
DELETE FROM employees;

-- Delete all users
DELETE FROM users;

-- Reset sequences/auto-increment for PostgreSQL
ALTER SEQUENCE IF EXISTS leave_requests_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS employees_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS users_id_seq RESTART WITH 1;

-- Reset for SQLite (if using SQLite, the DELETE statements are enough as it doesn't use sequences)
-- Just verify counts
SELECT COUNT(*) as leave_requests_count FROM leave_requests;
SELECT COUNT(*) as employees_count FROM employees;
SELECT COUNT(*) as users_count FROM users;
