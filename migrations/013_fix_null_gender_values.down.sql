-- Rollback: Set gender back to NULL for records that were updated
-- (Only useful if you need to undo the migration)
UPDATE employees SET gender = NULL WHERE gender = 'Male' AND user_id IS NOT NULL;
