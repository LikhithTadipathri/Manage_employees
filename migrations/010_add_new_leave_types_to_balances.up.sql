-- Initialize new leave types (Maternity, Unpaid, Personal) for existing employees
INSERT INTO leave_balances (employee_id, leave_type, balance, created_at, updated_at)
SELECT e.id, 'MATERNITY', 90, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM employees e
ON CONFLICT (employee_id, leave_type) DO NOTHING;

INSERT INTO leave_balances (employee_id, leave_type, balance, created_at, updated_at)
SELECT e.id, 'UNPAID', 10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM employees e
ON CONFLICT (employee_id, leave_type) DO NOTHING;

INSERT INTO leave_balances (employee_id, leave_type, balance, created_at, updated_at)
SELECT e.id, 'PERSONAL', 10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM employees e
ON CONFLICT (employee_id, leave_type) DO NOTHING;
