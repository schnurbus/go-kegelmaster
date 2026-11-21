-- Rundgang und Ersteinrichtung (GLOSSARY.md).
-- users.tour_seen_at: Rundgang gesehen (beendet oder übersprungen); NULL = noch nicht.
-- users.setup_hint_seen_at: Sprechblase zur Ersteinrichtung-Karte gesehen; NULL = noch nicht.
-- clubs.setup_dismissed_at: Ersteinrichtung-Karte ausgeblendet. Hängt am Club, nicht
--   am Besitzer, und gilt deshalb nach einer Übertragung weiter.
-- Bestehende Konten und Clubs gelten als eingerichtet: Sie bekommen weder Rundgang
-- noch Ersteinrichtung aufgedrängt.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS tour_seen_at       TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS setup_hint_seen_at TIMESTAMPTZ NULL;
UPDATE users SET tour_seen_at = now(), setup_hint_seen_at = now();

ALTER TABLE clubs ADD COLUMN IF NOT EXISTS setup_dismissed_at TIMESTAMPTZ NULL;
UPDATE clubs SET setup_dismissed_at = now();
