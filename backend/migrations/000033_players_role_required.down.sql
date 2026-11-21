ALTER TABLE players ALTER COLUMN role_id DROP NOT NULL;
ALTER TABLE players DROP CONSTRAINT IF EXISTS players_role_id_fkey;
ALTER TABLE players
    ADD CONSTRAINT players_role_id_fkey FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE SET NULL;
