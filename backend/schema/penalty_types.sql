-- Schema for penalty_types table and related foreign key tables
-- Extracted from migrations for sqlc code generation

CREATE TABLE IF NOT EXISTS penalty_types (
    id UUID PRIMARY KEY,
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    price INTEGER NOT NULL DEFAULT 0,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    replaced_by_id UUID REFERENCES penalty_types(id) ON DELETE SET NULL
);

