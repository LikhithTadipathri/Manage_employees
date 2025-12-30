-- Add Paternity leave type for existing employees
INSERT INTO leave_balances (employee_id, leave_type, balance, created_at, updated_at)
SELECT e.id, 'PATERNITY', 7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM employees e
ON CONFLICT (employee_id, leave_type) DO NOTHING;
