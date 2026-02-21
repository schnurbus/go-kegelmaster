ALTER TABLE players
DROP CONSTRAINT IF EXISTS players_partner_not_self,
DROP COLUMN IF EXISTS partner_id;
