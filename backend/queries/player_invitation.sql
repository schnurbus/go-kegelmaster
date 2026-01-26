-- name: CreatePlayerInvitation :one
INSERT INTO player_invitations (
    id,
    player_id,
    email,
    token,
    expires_at,
    created_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetPlayerInvitationByToken :one
SELECT * FROM player_invitations
WHERE token = $1;

-- name: AcceptPlayerInvitation :exec
UPDATE player_invitations
SET accepted_at = $1
WHERE id = $2;

-- name: GetPlayerInvitationByID :one
SELECT * FROM player_invitations
WHERE id = $1;

-- name: GetPlayerInvitationsByPlayerID :many
SELECT * FROM player_invitations
WHERE player_id = $1
ORDER BY created_at DESC;
