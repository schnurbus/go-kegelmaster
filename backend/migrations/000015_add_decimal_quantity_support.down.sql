ALTER TABLE game_day_fees
    DROP COLUMN IF EXISTS quantity_scale;

ALTER TABLE penalty_types
    DROP COLUMN IF EXISTS allows_decimal_quantity;
