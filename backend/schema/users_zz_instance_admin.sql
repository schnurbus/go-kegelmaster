-- Mirrors migration 000030_instance_admin (sqlc schema only).
ALTER TABLE users ADD COLUMN is_instance_admin BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN last_login_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
