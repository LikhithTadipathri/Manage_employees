-- Remove salary_deduction column from leave_requests table
ALTER TABLE leave_requests
DROP COLUMN salary_deduction;
