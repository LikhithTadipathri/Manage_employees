-- Check if leave_balances table exists
SELECT 
  EXISTS (SELECT 1 FROM information_schema.tables 
          WHERE table_schema='public' 
          AND table_name='leave_balances') AS leave_balances_exists,
  EXISTS (SELECT 1 FROM information_schema.tables 
          WHERE table_schema='public' 
          AND table_name='leave_requests') AS leave_requests_exists,
  EXISTS (SELECT 1 FROM information_schema.tables 
          WHERE table_schema='public' 
          AND table_name='employees') AS employees_exists;

-- If leave_balances exists, check its structure
SELECT table_name, column_name, data_type 
FROM information_schema.columns 
WHERE table_schema='public' AND table_name='leave_balances'
ORDER BY ordinal_position;

-- If leave_balances exists, check how many records
SELECT COUNT(*) as leave_balance_count FROM leave_balances;
