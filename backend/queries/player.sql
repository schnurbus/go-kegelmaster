-- name: CreatePlayer :one
INSERT INTO players (
    id,
    club_id,
    user_id,
    role_id,
    name,
    balance,
    start_balance,
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

-- name: GetPlayerByID :one
SELECT * FROM players
WHERE id = $1;

-- name: GetPlayersByClubID :many
SELECT * FROM players
WHERE club_id = $1
ORDER BY created_at DESC;

-- name: UpdatePlayer :one
UPDATE players
SET
    name = $1,
    balance = $2,
    start_balance = $3,
    user_id = $4,
    role_id = $5,
    updated_at = $6
WHERE id = $7
RETURNING *;

-- name: DeletePlayer :exec
DELETE FROM players
WHERE id = $1;

-- name: GetPlayerByUserIDAndClubID :one
SELECT * FROM players
WHERE user_id = $1 AND club_id = $2;
