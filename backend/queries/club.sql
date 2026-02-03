-- name: CreateClub :one
INSERT INTO clubs (
    id,
    name,
    balance,
    start_balance,
    base_fee,
    auto_tip_enabled,
    user_id,
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
    $8,
    $9
)
RETURNING *;

-- name: GetClubByID :one
SELECT * FROM clubs
WHERE id = $1;

-- name: GetAllClubs :many
SELECT * FROM clubs
ORDER BY created_at DESC;

-- name: GetClubsByUserID :many
SELECT * FROM clubs
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetClubsForUser :many
SELECT * FROM clubs c
WHERE c.user_id = $1
   OR EXISTS (
       SELECT 1 FROM players p
       WHERE p.club_id = c.id AND p.user_id = $1
   )
ORDER BY c.created_at DESC;

-- name: UpdateClub :one
UPDATE clubs
SET
    name = $1,
    balance = $2,
    start_balance = $3,
    base_fee = $4,
    auto_tip_enabled = $5,
    updated_at = $6
WHERE id = $7
RETURNING *;

-- name: UpdateClubBalance :exec
UPDATE clubs
SET balance = $1, updated_at = $2
WHERE id = $3;

-- name: DeleteClub :exec
DELETE FROM clubs
WHERE id = $1;

