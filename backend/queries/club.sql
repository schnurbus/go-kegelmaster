-- name: CreateClub :one
INSERT INTO clubs (
    id,
    name,
    balance,
    start_balance,
    base_fee,
    auto_tip_enabled,
    couples_mode_enabled,
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
    $9,
    $10
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
    couples_mode_enabled = $6,
    updated_at = $7
WHERE id = $8
RETURNING *;

-- name: UpdateClubBalance :exec
UPDATE clubs
SET balance = $1, updated_at = $2
WHERE id = $3;

-- name: UpdateClubOwner :one
UPDATE clubs
SET user_id = $1, updated_at = $2
WHERE id = $3
RETURNING *;

-- name: DeleteClub :exec
DELETE FROM clubs
WHERE id = $1;


-- name: IncrementClubBalance :one
UPDATE clubs
SET balance = balance + $1, updated_at = $2
WHERE id = $3
RETURNING balance;

-- name: GetClubSetupCounts :one
-- Data behind the Ersteinrichtung of a club; own_players counts the players
-- linked to the given user (the owner).
SELECT
    (SELECT COUNT(*) FROM penalty_types pt WHERE pt.club_id = sqlc.arg(club_id))::bigint AS penalty_types,
    (SELECT COUNT(*) FROM roles r WHERE r.club_id = sqlc.arg(club_id))::bigint AS roles,
    (SELECT COUNT(*) FROM players p WHERE p.club_id = sqlc.arg(club_id))::bigint AS players,
    (SELECT COUNT(*) FROM game_days g WHERE g.club_id = sqlc.arg(club_id) AND NOT g.is_draft)::bigint AS published_game_days,
    (SELECT COUNT(*) FROM player_invitations i JOIN players ip ON ip.id = i.player_id WHERE ip.club_id = sqlc.arg(club_id))::bigint AS invitations,
    (SELECT COUNT(*) FROM competitions co WHERE co.club_id = sqlc.arg(club_id))::bigint AS competitions,
    (SELECT COUNT(*) FROM players op WHERE op.club_id = sqlc.arg(club_id) AND op.user_id = sqlc.arg(user_id))::bigint AS own_players;

-- name: DismissClubSetup :exec
-- Hides the Ersteinrichtung card; an existing timestamp is kept.
UPDATE clubs
SET setup_dismissed_at = COALESCE(setup_dismissed_at, sqlc.arg(now)::timestamptz)
WHERE id = sqlc.arg(id);
