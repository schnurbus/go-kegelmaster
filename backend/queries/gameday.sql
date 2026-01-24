-- ==================== GAME DAYS ====================

-- name: CreateGameDay :one
INSERT INTO game_days (id, club_id, date, notes, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetGameDayByID :one
SELECT * FROM game_days WHERE id = $1;

-- name: GetGameDaysByClubID :many
SELECT * FROM game_days 
WHERE club_id = $1 
ORDER BY date DESC;

-- name: UpdateGameDay :one
UPDATE game_days 
SET date = $2, notes = $3, updated_at = $4
WHERE id = $1
RETURNING *;

-- name: DeleteGameDay :exec
DELETE FROM game_days WHERE id = $1;

-- name: CheckGameDayExistsByClubAndDate :one
SELECT COUNT(*) > 0 as exists
FROM game_days
WHERE club_id = $1 AND date = $2;

-- ==================== PARTICIPANTS ====================

-- name: CreateGameDayParticipant :one
INSERT INTO game_day_participants (id, game_day_id, player_id, created_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetGameDayParticipants :many
SELECT gdp.*, p.name as player_name
FROM game_day_participants gdp
JOIN players p ON p.id = gdp.player_id
WHERE gdp.game_day_id = $1
ORDER BY gdp.created_at ASC;

-- name: DeleteGameDayParticipant :exec
DELETE FROM game_day_participants 
WHERE game_day_id = $1 AND player_id = $2;

-- name: GetParticipantByGameDayAndPlayer :one
SELECT * FROM game_day_participants 
WHERE game_day_id = $1 AND player_id = $2;

-- name: DeleteAllParticipantsByGameDay :exec
DELETE FROM game_day_participants WHERE game_day_id = $1;

-- ==================== FEES ====================

-- name: CreateGameDayFee :one
INSERT INTO game_day_fees (
    id, game_day_participant_id, penalty_type_id,
    penalty_type_name, penalty_type_description, penalty_type_price,
    count, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateGameDayFee :one
UPDATE game_day_fees
SET count = $2, updated_at = $3
WHERE id = $1
RETURNING *;

-- name: UpsertGameDayFee :one
INSERT INTO game_day_fees (
    id, game_day_participant_id, penalty_type_id,
    penalty_type_name, penalty_type_description, penalty_type_price,
    count, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (game_day_participant_id, penalty_type_id)
DO UPDATE SET 
    count = EXCLUDED.count,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: GetGameDayFeesByParticipant :many
SELECT * FROM game_day_fees 
WHERE game_day_participant_id = $1
ORDER BY penalty_type_name ASC;

-- name: GetGameDayFeesByGameDay :many
SELECT gdf.*, gdp.player_id, p.name as player_name
FROM game_day_fees gdf
JOIN game_day_participants gdp ON gdp.id = gdf.game_day_participant_id
JOIN players p ON p.id = gdp.player_id
WHERE gdp.game_day_id = $1
ORDER BY p.name ASC, gdf.penalty_type_name ASC;

-- name: DeleteGameDayFee :exec
DELETE FROM game_day_fees WHERE id = $1;

-- name: DeleteGameDayFeeByParticipantAndType :exec
DELETE FROM game_day_fees 
WHERE game_day_participant_id = $1 AND penalty_type_id = $2;

-- name: DeleteAllFeesByParticipant :exec
DELETE FROM game_day_fees WHERE game_day_participant_id = $1;
