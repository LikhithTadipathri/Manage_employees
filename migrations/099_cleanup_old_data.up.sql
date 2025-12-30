-- Truncate all employee and leave data but keep user seed data
TRUNCATE TABLE leave_requests CASCADE;
TRUNCATE TABLE leave_balances CASCADE;
TRUNCATE TABLE employees CASCADE;

-- Only keep the 4 seed users
DELETE FROM users WHERE id > 4;

-- Reset sequences
ALTER SEQUENCE users_id_seq RESTART WITH 5;
ALTER SEQUENCE employees_id_seq RESTART WITH 3;
ALTER SEQUENCE leave_requests_id_seq RESTART WITH 1;
