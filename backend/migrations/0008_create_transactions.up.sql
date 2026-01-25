CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    player_id UUID REFERENCES players(id) ON DELETE CASCADE,
    transaction_type TEXT NOT NULL CHECK (transaction_type IN ('base_fee', 'fee', 'deposit', 'tip', 'expense')),
    
    -- Amount in cents (negative = debt/expense, positive = credit/income)
    amount INTEGER NOT NULL,
    
    -- Auto-generated descriptions for system transactions
    description TEXT DEFAULT '',
    
    -- Source tracking (nullable for manual transactions)
    game_day_fee_id UUID REFERENCES game_day_fees(id) ON DELETE CASCADE,
    game_day_id UUID REFERENCES game_days(id) ON DELETE CASCADE,
    
    -- Balance snapshots for audit trail
    player_balance_before INTEGER,
    player_balance_after INTEGER,
    club_balance_before INTEGER NOT NULL,
    club_balance_after INTEGER NOT NULL,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_transactions_club_id ON transactions(club_id);
CREATE INDEX IF NOT EXISTS idx_transactions_player_id ON transactions(player_id);
CREATE INDEX IF NOT EXISTS idx_transactions_game_day_id ON transactions(game_day_id);
CREATE INDEX IF NOT EXISTS idx_transactions_game_day_fee_id ON transactions(game_day_fee_id);
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(transaction_type);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_club_created ON transactions(club_id, created_at DESC);
