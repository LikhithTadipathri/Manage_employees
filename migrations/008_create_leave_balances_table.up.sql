-- Create leave_balances table
CREATE TABLE IF NOT EXISTS leave_balances (
    id SERIAL PRIMARY KEY,
    employee_id INTEGER NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    leave_type VARCHAR(50) NOT NULL,
    balance INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(employee_id, leave_type)
);

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_leave_balances_employee_id ON leave_balances(employee_id);
CREATE INDEX IF NOT EXISTS idx_leave_balances_employee_leave_type ON leave_balances(employee_id, leave_type);

-- Insert initial balances for existing employees
INSERT INTO leave_balances (employee_id, leave_type, balance, created_at, updated_at)
SELECT e.id, 'ANNUAL', 10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM employees e
ON CONFLICT (employee_id, leave_type) DO NOTHING;

INSERT INTO leave_balances (employee_id, leave_type, balance, created_at, updated_at)
SELECT e.id, 'SICK', 15, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM employees e
ON CONFLICT (employee_id, leave_type) DO NOTHING;

INSERT INTO leave_balances (employee_id, leave_type, balance, created_at, updated_at)
SELECT e.id, 'CASUAL', 10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM employees e
ON CONFLICT (employee_id, leave_type) DO NOTHING;
