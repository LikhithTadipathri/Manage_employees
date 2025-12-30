-- Rollback Migration 006: Reset user_id to NULL for previously linked employees
-- This is a destructive operation - only used for rollback

-- Identify and reset employee-user links created in this migration
-- We reset user_id to NULL for employees without explicit user accounts
UPDATE employees 
SET user_id = NULL
WHERE user_id IS NOT NULL
  AND id NOT IN (
    -- Keep links for employees that were explicitly created with user accounts post-Migration 005
    SELECT DISTINCT employee_id 
    FROM (
      SELECT MIN(e.id) as employee_id 
      FROM employees e
      INNER JOIN users u ON e.user_id = u.id
      WHERE e.created_at > NOW() - INTERVAL '1 day'
      GROUP BY e.user_id
    ) as recent_links
  );
