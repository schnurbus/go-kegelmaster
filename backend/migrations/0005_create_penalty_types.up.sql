CREATE TABLE IF NOT EXISTS penalty_types (
    id UUID PRIMARY KEY,
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    price INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    replaced_by_id UUID REFERENCES penalty_types(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_penalty_types_club_id ON penalty_types(club_id);
CREATE INDEX IF NOT EXISTS idx_penalty_types_deleted_at ON penalty_types(deleted_at);
CREATE INDEX IF NOT EXISTS idx_penalty_types_replaced_by_id ON penalty_types(replaced_by_id);

-- Unique constraint: club_id + name must be unique for non-deleted penalty types
CREATE UNIQUE INDEX IF NOT EXISTS idx_penalty_types_club_name_unique 
    ON penalty_types(club_id, name) 
    WHERE deleted_at IS NULL;

