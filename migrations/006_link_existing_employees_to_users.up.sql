-- Migration 006: Link existing employees to users by username pattern matching
-- This migration links employees created before Migration 005 (which had NULL user_id)
-- to their corresponding users using a naming convention: first_name_last_name

-- Link employees to users where username matches first_name_last_name pattern
UPDATE employees 
SET user_id = (
  SELECT id FROM users 
  WHERE LOWER(users.username) = LOWER(CONCAT(employees.first_name, '_', employees.last_name))
  LIMIT 1
)
WHERE user_id IS NULL 
  AND EXISTS (
    SELECT 1 FROM users 
    WHERE LOWER(users.username) = LOWER(CONCAT(employees.first_name, '_', employees.last_name))
  );

-- Log: Check results with:
-- SELECT COUNT(*) as linked FROM employees WHERE user_id IS NOT NULL;
-- SELECT COUNT(*) as unlinked FROM employees WHERE user_id IS NULL;
-- SELECT id, first_name, last_name, user_id FROM employees WHERE user_id IS NULL;
