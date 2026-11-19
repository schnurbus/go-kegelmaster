-- Admin-Eingriffe (Betreiber-Übersicht Phase 3, KEG-10).
-- disabled_at: Konto vom Instanz-Admin gesperrt; Login, Cookie-Session und API-Tokens werden abgelehnt.
-- sessions_valid_after: Browser-Sessions (JWT), die vor diesem Zeitpunkt ausgestellt wurden, sind ungültig
--   (zusätzlich zu password_changed_at).
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS disabled_at          TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS sessions_valid_after TIMESTAMPTZ NULL;
