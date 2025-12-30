-- Add salary_deduction column to leave_requests table
ALTER TABLE leave_requests
ADD COLUMN salary_deduction DECIMAL(10, 2) DEFAULT 0;
