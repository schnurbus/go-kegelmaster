-- Add display_order column for custom sorting
ALTER TABLE penalty_types ADD COLUMN display_order INTEGER NOT NULL DEFAULT 0;

-- Set display_order for existing entries based on created_at
UPDATE penalty_types
SET display_order = subquery.row_number
FROM (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY club_id ORDER BY created_at ASC) as row_number
    FROM penalty_types
    WHERE deleted_at IS NULL
) AS subquery
WHERE penalty_types.id = subquery.id;

-- Create index for efficient sorting
CREATE INDEX IF NOT EXISTS idx_penalty_types_club_display_order ON penalty_types(club_id, display_order);

