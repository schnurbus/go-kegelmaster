-- name: CreateCompetition :one
INSERT INTO competitions (
    id,
    club_id,
    name,
    scoring_type,
    is_gender_specific,
    display_order,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetCompetitionByID :one
SELECT * FROM competitions WHERE id = $1;

-- name: GetCompetitionsByClubID :many
SELECT * FROM competitions
WHERE club_id = $1
ORDER BY display_order ASC, created_at ASC;

-- name: UpdateCompetition :one
UPDATE competitions
SET
    name = $1,
    scoring_type = $2,
    is_gender_specific = $3,
    display_order = $4,
    updated_at = $5
WHERE id = $6
RETURNING *;

-- name: DeleteCompetition :exec
DELETE FROM competitions WHERE id = $1;

-- name: GetMaxCompetitionDisplayOrderByClubID :one
SELECT COALESCE(MAX(display_order), 0)::int as max_display_order
FROM competitions
WHERE club_id = $1;
