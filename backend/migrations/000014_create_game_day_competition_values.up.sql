CREATE TABLE IF NOT EXISTS game_day_competition_values (
    id UUID PRIMARY KEY,
    game_day_participant_id UUID NOT NULL REFERENCES game_day_participants(id) ON DELETE CASCADE,
    competition_id UUID NOT NULL REFERENCES competitions(id) ON DELETE RESTRICT,
    value INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(game_day_participant_id, competition_id)
);

CREATE INDEX IF NOT EXISTS idx_game_day_competition_values_participant ON game_day_competition_values(game_day_participant_id);
CREATE INDEX IF NOT EXISTS idx_game_day_competition_values_competition ON game_day_competition_values(competition_id);
