-- Mirrors migration 000025_security_hardening (sqlc schema only).
ALTER TABLE users ADD COLUMN password_changed_at TIMESTAMPTZ NULL;
