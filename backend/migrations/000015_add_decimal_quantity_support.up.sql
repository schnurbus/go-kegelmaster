-- penalty_types: allow decimal quantity per fee type
ALTER TABLE penalty_types
    ADD COLUMN IF NOT EXISTS allows_decimal_quantity BOOLEAN NOT NULL DEFAULT false;

-- game_day_fees: snapshot scale for quantity (1 = integer, 100 = 2 decimal places)
ALTER TABLE game_day_fees
    ADD COLUMN IF NOT EXISTS quantity_scale INTEGER NOT NULL DEFAULT 1;
