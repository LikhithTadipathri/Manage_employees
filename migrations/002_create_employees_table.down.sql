-- Migration: 002_create_employees_table.down.sql
-- Description: Drop employees table (rollback)
-- Date: 2025-12-04

DROP TABLE IF EXISTS employees CASCADE;
