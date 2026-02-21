-- name: CreatePlayer :one
INSERT INTO players (
    id,
    club_id,
    user_id,
    role_id,
    name,
    balance,
    start_balance,
    gender,
    inactive,
    partner_id,
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
    $9,
    $10,
    $11,
    $12
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
    gender = $6,
    inactive = $7,
    partner_id = $8,
    updated_at = $9
WHERE id = $10
RETURNING *;

-- name: UpdatePlayerBalance :exec
UPDATE players
SET balance = $1, updated_at = $2
WHERE id = $3;

-- name: DeletePlayer :exec
DELETE FROM players
WHERE id = $1;

-- name: GetPlayerByUserIDAndClubID :one
SELECT * FROM players
WHERE user_id = $1 AND club_id = $2;

-- name: CountPlayersByRoleIDForClub :many
SELECT role_id, COUNT(*) AS count
FROM players
WHERE club_id = $1 AND role_id IS NOT NULL
GROUP BY role_id;
