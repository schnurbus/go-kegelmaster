-- Mirrors migration 000031_admin_actions (sqlc schema only).
ALTER TABLE users ADD COLUMN disabled_at TIMESTAMPTZ NULL;
ALTER TABLE users ADD COLUMN sessions_valid_after TIMESTAMPTZ NULL;
