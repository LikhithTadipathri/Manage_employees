-- Remove superadmin user if it exists
DELETE FROM users WHERE username = 'superadmin';
