-- Jeder Spieler hat genau eine Rolle (GLOSSARY.md). Bisher setzte das Löschen
-- einer Rolle players.role_id still auf NULL. Jetzt verhindert RESTRICT das
-- Löschen einer Rolle mit Spielern, und NOT NULL lehnt Spieler ohne Rolle ab.
-- Gibt es im Bestand Spieler ohne Rolle, bricht die Migration ab, statt eine
-- Rolle zu raten; diese Spieler müssen vorher von Hand eine Rolle bekommen:
--   SELECT p.id, p.name, c.name AS club FROM players p JOIN clubs c ON c.id = p.club_id WHERE p.role_id IS NULL;
DO $$
DECLARE
    missing INTEGER;
BEGIN
    SELECT COUNT(*) INTO missing FROM players WHERE role_id IS NULL;
    IF missing > 0 THEN
        RAISE EXCEPTION 'Migration 000033: % Spieler ohne Rolle gefunden. Bitte zuerst jedem Spieler eine Rolle zuordnen (SELECT id, name, club_id FROM players WHERE role_id IS NULL).', missing;
    END IF;
END $$;

ALTER TABLE players DROP CONSTRAINT IF EXISTS players_role_id_fkey;
ALTER TABLE players
    ADD CONSTRAINT players_role_id_fkey FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE RESTRICT;
ALTER TABLE players ALTER COLUMN role_id SET NOT NULL;
