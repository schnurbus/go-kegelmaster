ALTER TABLE users DROP COLUMN IF EXISTS password_changed_at;
DELETE FROM password_reset_tokens;
