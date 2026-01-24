-- name: CreatePenaltyType :one
INSERT INTO penalty_types (
    id,
    club_id,
    name,
    description,
    price,
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

-- name: GetPenaltyTypeByID :one
SELECT * FROM penalty_types
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetPenaltyTypesByClubID :many
SELECT * FROM penalty_types
WHERE club_id = $1 AND deleted_at IS NULL
ORDER BY display_order ASC, created_at ASC;

-- name: MarkPenaltyTypeAsReplaced :exec
UPDATE penalty_types
SET deleted_at = $1,
    replaced_by_id = $2
WHERE id = $3 AND deleted_at IS NULL;

-- name: DeletePenaltyType :exec
UPDATE penalty_types
SET deleted_at = $1
WHERE id = $2 AND deleted_at IS NULL;

-- name: UpdatePenaltyTypeDisplayOrder :exec
UPDATE penalty_types
SET display_order = $1,
    updated_at = $2
WHERE id = $3 AND deleted_at IS NULL;

-- name: GetMaxDisplayOrderByClubID :one
SELECT COALESCE(MAX(display_order), 0) as max_display_order
FROM penalty_types
WHERE club_id = $1 AND deleted_at IS NULL;

