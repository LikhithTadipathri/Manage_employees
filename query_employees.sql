-- Check users and employees
SELECT 
    u.id as user_id, 
    u.username, 
    u.role,
    e.id as employee_id,
    e.first_name,
    e.user_id as emp_user_id
FROM users u
LEFT JOIN employees e ON u.id = e.user_id
ORDER BY u.id;
