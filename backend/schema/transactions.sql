-- Schema for transactions table
-- Extracted from migrations for sqlc code generation

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS clubs (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    balance INTEGER NOT NULL DEFAULT 0,
    base_fee INTEGER NOT NULL DEFAULT 0,
    auto_tip_enabled BOOLEAN NOT NULL DEFAULT true,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS players (
    id UUID PRIMARY KEY,
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    balance INTEGER NOT NULL DEFAULT 0,
    start_balance INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY,
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    pays_base_fee BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(club_id, name)
);

CREATE TABLE IF NOT EXISTS game_days (
    id UUID PRIMARY KEY,
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    notes TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS game_day_fees (
    id UUID PRIMARY KEY,
    game_day_participant_id UUID NOT NULL,
    penalty_type_id UUID NOT NULL,
    penalty_type_name TEXT NOT NULL,
    penalty_type_description TEXT DEFAULT '',
    penalty_type_price INTEGER NOT NULL,
    count INTEGER NOT NULL DEFAULT 1 CHECK (count > 0),
    quantity_scale INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

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
