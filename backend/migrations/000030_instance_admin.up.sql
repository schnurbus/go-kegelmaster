-- Instanz-Administration (Betreiber-Übersicht, KEG-7).
-- is_instance_admin: Betreiber-Rolle, strikt getrennt von den Klub-Berechtigungen.
--   Vergabe nur per CLI (`kegelmaster admin grant`) oder INSTANCE_ADMIN_EMAILS, nie über die Weboberfläche.
-- last_login_at: letzter erfolgreicher Login (Passwort oder Passkey); NULL = unbekannt.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS is_instance_admin BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS last_login_at     TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
