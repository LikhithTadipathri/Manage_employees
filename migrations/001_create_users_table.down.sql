-- Migration: 001_create_users_table.down.sql
-- Description: Drop users table (rollback)
-- Date: 2025-12-04

DROP TABLE IF EXISTS users CASCADE;

-- \l
-- \c employee_db
-- \dt
-- \i 'D:/Go/src/Task/migrations/001_create_users_table.up.sql'