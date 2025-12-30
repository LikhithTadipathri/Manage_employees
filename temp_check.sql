SELECT u.id, u.username, u.role, e.id as employee_id, e.first_name, e.last_name, e.user_id FROM users u LEFT JOIN employees e ON u.id = e.id ORDER BY u.id;
