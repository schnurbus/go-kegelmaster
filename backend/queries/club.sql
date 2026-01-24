-- name: CreateClub :one
INSERT INTO clubs (
    id,
    name,
    balance,
    base_fee,
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
    $7
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

-- name: UpdateClub :one
UPDATE clubs
SET
    name = $1,
    balance = $2,
    base_fee = $3,
    updated_at = $4
WHERE id = $5
RETURNING *;

-- name: DeleteClub :exec
DELETE FROM clubs
WHERE id = $1;

