-- Rollback: Remove Paternity leave type
DELETE FROM leave_balances WHERE leave_type = 'PATERNITY';
