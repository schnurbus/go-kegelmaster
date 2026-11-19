DROP INDEX IF EXISTS idx_users_created_at;

ALTER TABLE users
    DROP COLUMN IF EXISTS last_login_at,
    DROP COLUMN IF EXISTS is_instance_admin;
