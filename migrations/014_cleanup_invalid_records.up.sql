-- Migration: 014_cleanup_invalid_records.up.sql
-- Description: Clean up any records with NULL or empty gender values
-- This migration removes problematic records that don't meet gender validation requirements

-- Delete leave requests for employees that have NULL/empty gender (to maintain referential integrity)
DELETE FROM leave_requests 
WHERE employee_id IN (
    SELECT id FROM employees 
    WHERE gender IS NULL OR gender = '' OR gender NOT IN ('Male', 'Female')
);

-- Delete leave balances for employees that have NULL/empty gender
DELETE FROM leave_balances 
WHERE employee_id IN (
    SELECT id FROM employees 
    WHERE gender IS NULL OR gender = '' OR gender NOT IN ('Male', 'Female')
);

-- Delete employees with invalid gender
DELETE FROM employees 
WHERE gender IS NULL OR gender = '' OR gender NOT IN ('Male', 'Female');

-- Set any remaining NULL genders to 'Male' (safety measure)
UPDATE employees 
SET gender = 'Male' 
WHERE gender IS NULL OR gender = '';
