-- Migration: 001_create_users_table.up.sql
-- Description: Create users table
-- Date: 2025-12-04

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- -- Add comment to table
-- COMMENT ON TABLE users IS 'Stores user authentication and profile information';

-- -- Add comments to columns
-- COMMENT ON COLUMN users.id IS 'Unique user identifier';
-- COMMENT ON COLUMN users.username IS 'Unique username for login';
-- COMMENT ON COLUMN users.email IS 'User email address';
-- COMMENT ON COLUMN users.password_hash IS 'Bcrypt hashed password';
-- COMMENT ON COLUMN users.role IS 'User role: admin or user';
-- COMMENT ON COLUMN users.created_at IS 'Record creation timestamp';
-- COMMENT ON COLUMN users.updated_at IS 'Record last update timestamp';

-- Migration: 001_create_users_table.up.sql
-- Description: Create users table and insert sample user records
-- Date: 2025-12-04

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add comment to table
COMMENT ON TABLE users IS 'Stores user authentication and profile information';

-- Add comments to columns
COMMENT ON COLUMN users.id IS 'Unique user identifier';
COMMENT ON COLUMN users.username IS 'Unique username for login';
COMMENT ON COLUMN users.email IS 'User email address';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hashed password';
COMMENT ON COLUMN users.role IS 'User role: admin or user';
COMMENT ON COLUMN users.created_at IS 'Record creation timestamp';
COMMENT ON COLUMN users.updated_at IS 'Record last update timestamp';

-- Insert initial sample users with real timestamps
-- Passwords:
-- Admin: Admin@123
-- Manager: Manager@123
-- User: User@123

INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
VALUES
    ('Vinay', 'vinnu@gmail.com',
     '$2a$10$EixZaYVK1fsbw1ZfbX3OXePaWxn96p36Zfb8KMQnUUJ6Y4ML9RX2m',
     'admin',
     '2025-01-10 09:30:00', '2025-01-10 09:30:00')

    
