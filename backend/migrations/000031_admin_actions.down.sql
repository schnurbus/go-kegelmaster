ALTER TABLE users
    DROP COLUMN IF EXISTS sessions_valid_after,
    DROP COLUMN IF EXISTS disabled_at;
