ALTER TABLE players
ADD COLUMN partner_id UUID NULL REFERENCES players(id) ON DELETE SET NULL,
ADD CONSTRAINT players_partner_not_self CHECK (id != partner_id);
