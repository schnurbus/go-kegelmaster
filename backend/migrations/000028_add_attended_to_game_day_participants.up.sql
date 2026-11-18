-- A participant can be booked on a game day without having been there (e.g. a
-- penalty for a forgotten item). Absent participants keep their fees but are
-- excluded from attendance and competition statistics.
ALTER TABLE game_day_participants ADD COLUMN attended BOOLEAN NOT NULL DEFAULT true;
