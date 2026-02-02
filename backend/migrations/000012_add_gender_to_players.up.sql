ALTER TABLE players
ADD COLUMN IF NOT EXISTS gender TEXT CHECK (gender IS NULL OR gender IN ('male', 'female'));
