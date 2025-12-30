-- Add notes column to leave_requests table
ALTER TABLE leave_requests
ADD COLUMN notes TEXT;
