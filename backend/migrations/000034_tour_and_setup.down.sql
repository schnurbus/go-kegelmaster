ALTER TABLE clubs DROP COLUMN IF EXISTS setup_dismissed_at;
ALTER TABLE users
    DROP COLUMN IF EXISTS setup_hint_seen_at,
    DROP COLUMN IF EXISTS tour_seen_at;
