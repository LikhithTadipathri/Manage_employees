-- Create superadmin user
-- Username: superadmin
-- Password: passkey (bcrypt hashed)
-- Hash: $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36DuP32e
-- This is the bcrypt hash of "passkey" generated with cost 10

INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
VALUES (
  'superadmin',
  'superadmin@system.local',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36DuP32e',
  'admin',
  NOW(),
  NOW()
) ON CONFLICT (username) DO NOTHING;
