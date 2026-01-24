DROP INDEX IF EXISTS idx_penalty_types_club_display_order;
ALTER TABLE penalty_types DROP COLUMN IF EXISTS display_order;

