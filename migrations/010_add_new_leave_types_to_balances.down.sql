-- Remove new leave types from leave_balances
DELETE FROM leave_balances WHERE leave_type IN ('MATERNITY', 'UNPAID', 'PERSONAL');
