CREATE TABLE IF NOT EXISTS competitions (
    id UUID PRIMARY KEY,
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    scoring_type TEXT NOT NULL CHECK (scoring_type IN ('winner', 'loser', 'both')),
    is_gender_specific BOOLEAN NOT NULL DEFAULT false,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_competitions_club_id ON competitions(club_id);
