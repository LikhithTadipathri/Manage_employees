-- Quick database cleanup: Remove all employee/leave data but keep users
-- This allows the application to recreate fresh seed data on next startup

-- Clear leave requests and leave balances
TRUNCATE TABLE leave_requests CASCADE;
TRUNCATE TABLE leave_balances CASCADE;

-- Clear employees (but keep user accounts)
TRUNCATE TABLE employees CASCADE;

-- Reset sequences (for PostgreSQL)
ALTER SEQUENCE IF EXISTS leave_requests_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS leave_balances_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS employees_id_seq RESTART WITH 1;

-- For SQLite, this would just be DELETE statements (no sequences)
-- SQLite users should just delete and re-insert if needed
