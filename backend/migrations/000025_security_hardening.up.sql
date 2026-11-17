-- Sessions issued before the last password change become invalid.
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_changed_at TIMESTAMPTZ NULL;

-- Password reset tokens are stored as SHA-256 hashes from now on.
-- Outstanding plaintext tokens can no longer be matched, so drop them.
DELETE FROM password_reset_tokens;
