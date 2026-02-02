-- transaction_date: for fee/base_fee use game day date; for manual use created_at date
ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS transaction_date DATE NOT NULL DEFAULT CURRENT_DATE;

-- Backfill existing rows: use created_at date
UPDATE transactions
SET transaction_date = (created_at AT TIME ZONE 'UTC')::date;
