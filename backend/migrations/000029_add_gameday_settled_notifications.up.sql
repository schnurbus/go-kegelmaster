-- E-Mail-Benachrichtigung, sobald ein Spieltag abgerechnet (veröffentlicht) wurde.
-- Opt-in: für bestehende und neue Benutzer standardmäßig aus.
ALTER TABLE users ADD COLUMN notify_gameday_settled BOOLEAN NOT NULL DEFAULT false;

-- Merkt sich, für welche Spieltage die Benachrichtigung bereits verschickt wurde,
-- damit ein erneutes Veröffentlichen (Entwurf -> veröffentlicht -> Entwurf -> ...)
-- keine weiteren E-Mails auslöst.
CREATE TABLE IF NOT EXISTS game_day_notifications (
    id UUID PRIMARY KEY,
    game_day_id UUID NOT NULL UNIQUE REFERENCES game_days(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Bereits veröffentlichte Spieltage gelten als benachrichtigt.
INSERT INTO game_day_notifications (id, game_day_id, created_at, updated_at)
SELECT gen_random_uuid(), id, now(), now() FROM game_days WHERE is_draft = false
ON CONFLICT (game_day_id) DO NOTHING;
