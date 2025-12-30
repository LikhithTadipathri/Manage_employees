-- Add notes column to leave_requests table if it doesn't exist
-- For PostgreSQL
ALTER TABLE leave_requests ADD COLUMN IF NOT EXISTS notes TEXT;

-- If using SQLite, use this instead (comment out the above line):
-- ALTER TABLE leave_requests ADD COLUMN notes TEXT;
