CREATE TABLE IF NOT EXISTS player_invitations (
    id UUID PRIMARY KEY,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_player_invitations_token ON player_invitations(token);
CREATE INDEX IF NOT EXISTS idx_player_invitations_player_id ON player_invitations(player_id);
