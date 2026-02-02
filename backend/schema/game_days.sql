-- Game Days table
CREATE TABLE IF NOT EXISTS game_days (
    id UUID PRIMARY KEY,
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    notes TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_game_days_club_id ON game_days(club_id);
CREATE INDEX IF NOT EXISTS idx_game_days_date ON game_days(date);
CREATE INDEX IF NOT EXISTS idx_game_days_club_date ON game_days(club_id, date);

-- Game Day Participants (many-to-many: game_days ↔ players)
CREATE TABLE IF NOT EXISTS game_day_participants (
    id UUID PRIMARY KEY,
    game_day_id UUID NOT NULL REFERENCES game_days(id) ON DELETE CASCADE,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    UNIQUE(game_day_id, player_id)
);

CREATE INDEX IF NOT EXISTS idx_game_day_participants_game_day ON game_day_participants(game_day_id);
CREATE INDEX IF NOT EXISTS idx_game_day_participants_player ON game_day_participants(player_id);

-- Game Day Fees (with penalty type snapshot for historical accuracy)
CREATE TABLE IF NOT EXISTS game_day_fees (
    id UUID PRIMARY KEY,
    game_day_participant_id UUID NOT NULL REFERENCES game_day_participants(id) ON DELETE CASCADE,
    penalty_type_id UUID NOT NULL REFERENCES penalty_types(id) ON DELETE RESTRICT,
    -- Snapshot fields preserve historical data when penalty types change
    penalty_type_name TEXT NOT NULL,
    penalty_type_description TEXT DEFAULT '',
    penalty_type_price INTEGER NOT NULL,  -- Cent-Betrag
    count INTEGER NOT NULL DEFAULT 1 CHECK (count > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    UNIQUE(game_day_participant_id, penalty_type_id)
);

CREATE INDEX IF NOT EXISTS idx_game_day_fees_participant ON game_day_fees(game_day_participant_id);
CREATE INDEX IF NOT EXISTS idx_game_day_fees_penalty_type ON game_day_fees(penalty_type_id);

-- Game Day Competition Values (one value per participant per competition)
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
