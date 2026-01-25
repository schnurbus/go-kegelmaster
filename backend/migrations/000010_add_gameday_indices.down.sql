-- Drop indices for game day summaries
-- Note: idx_transactions_game_day_id is not dropped as it was created in 0008_create_transactions.up.sql

DROP INDEX IF EXISTS idx_transactions_game_day_id_transaction_type;
DROP INDEX IF EXISTS idx_game_day_participants_game_day_id;
