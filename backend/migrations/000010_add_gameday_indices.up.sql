-- Add indices for better query performance on game day summaries
-- Note: idx_transactions_game_day_id already exists from 0008_create_transactions.up.sql

CREATE INDEX IF NOT EXISTS idx_game_day_participants_game_day_id 
    ON game_day_participants(game_day_id);

CREATE INDEX IF NOT EXISTS idx_transactions_game_day_id_transaction_type 
    ON transactions(game_day_id, transaction_type);
