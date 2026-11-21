-- Mirrors migration 000034_tour_and_setup (sqlc schema only).
ALTER TABLE users ADD COLUMN tour_seen_at TIMESTAMPTZ NULL;
ALTER TABLE users ADD COLUMN setup_hint_seen_at TIMESTAMPTZ NULL;
ALTER TABLE clubs ADD COLUMN setup_dismissed_at TIMESTAMPTZ NULL;
