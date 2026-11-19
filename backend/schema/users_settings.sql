-- Mirrors migration 000029_add_gameday_settled_notifications (sqlc schema only).
ALTER TABLE users ADD COLUMN notify_gameday_settled BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS game_day_notifications (
    id UUID PRIMARY KEY,
    game_day_id UUID NOT NULL UNIQUE REFERENCES game_days(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
