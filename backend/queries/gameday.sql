-- ==================== GAME DAYS ====================

-- name: CreateGameDay :one
INSERT INTO game_days (id, club_id, date, notes, is_draft, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetGameDayByID :one
SELECT * FROM game_days WHERE id = $1;

-- name: GetGameDaysByClubID :many
SELECT * FROM game_days 
WHERE club_id = $1 
ORDER BY date DESC;

-- name: UpdateGameDay :one
UPDATE game_days 
SET date = $2, notes = $3, is_draft = $4, updated_at = $5
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
SELECT gdp.id, gdp.game_day_id, gdp.player_id, gdp.created_at, p.name as player_name, p.gender as player_gender
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
    count, quantity_scale, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: UpdateGameDayFee :one
UPDATE game_day_fees
SET count = $2, quantity_scale = $3, updated_at = $4
WHERE id = $1
RETURNING *;

-- name: UpsertGameDayFee :one
INSERT INTO game_day_fees (
    id, game_day_participant_id, penalty_type_id,
    penalty_type_name, penalty_type_description, penalty_type_price,
    count, quantity_scale, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (game_day_participant_id, penalty_type_id)
DO UPDATE SET 
    count = EXCLUDED.count,
    quantity_scale = EXCLUDED.quantity_scale,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: GetGameDayFeesByParticipant :many
SELECT * FROM game_day_fees 
WHERE game_day_participant_id = $1
ORDER BY penalty_type_name ASC;

-- name: GetGameDayFeesByParticipantForUpdate :many
SELECT * FROM game_day_fees
WHERE game_day_participant_id = $1
ORDER BY penalty_type_name ASC
FOR UPDATE;

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

-- ==================== COMPETITION VALUES ====================

-- name: CreateGameDayCompetitionValue :one
INSERT INTO game_day_competition_values (
    id, game_day_participant_id, competition_id, value, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpsertGameDayCompetitionValue :one
INSERT INTO game_day_competition_values (
    id, game_day_participant_id, competition_id, value, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (game_day_participant_id, competition_id)
DO UPDATE SET
    value = EXCLUDED.value,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: GetGameDayCompetitionValuesByParticipant :many
SELECT * FROM game_day_competition_values
WHERE game_day_participant_id = $1
ORDER BY competition_id ASC;

-- name: GetGameDayCompetitionValuesByGameDay :many
SELECT gdcv.*, gdp.player_id, p.name as player_name
FROM game_day_competition_values gdcv
JOIN game_day_participants gdp ON gdp.id = gdcv.game_day_participant_id
JOIN players p ON p.id = gdp.player_id
WHERE gdp.game_day_id = $1
ORDER BY p.name ASC, gdcv.competition_id ASC;

-- name: DeleteGameDayCompetitionValue :exec
DELETE FROM game_day_competition_values WHERE id = $1;

-- name: DeleteGameDayCompetitionValueByParticipantAndCompetition :exec
DELETE FROM game_day_competition_values
WHERE game_day_participant_id = $1 AND competition_id = $2;

-- name: DeleteAllCompetitionValuesByParticipant :exec
DELETE FROM game_day_competition_values WHERE game_day_participant_id = $1;

-- ==================== SUMMARIES ====================

-- name: GetGameDaySummariesByClubID :many
SELECT 
    gd.id,
    gd.club_id,
    gd.date,
    gd.notes,
    gd.is_draft,
    gd.created_at,
    gd.updated_at,
    COALESCE(p.participant_count, 0)::int as participant_count,
    COALESCE(t.penalty_fee_total, 0)::int as penalty_fee_total
FROM game_days gd
LEFT JOIN (
    SELECT game_day_id, COUNT(*)::int as participant_count
    FROM game_day_participants
    GROUP BY game_day_id
) p ON gd.id = p.game_day_id
LEFT JOIN (
    SELECT game_day_id, SUM(amount)::int as penalty_fee_total
    FROM transactions
    WHERE transaction_type = 'fee'
    GROUP BY game_day_id
) t ON gd.id = t.game_day_id
WHERE gd.club_id = $1
ORDER BY gd.date DESC;

-- ==================== PLAYER PENALTY HISTORY (DASHBOARD) ====================

-- name: GetPlayerPenaltyHistoryByClubAndPlayer :many
SELECT gd.id as game_day_id, gd.date, gdf.penalty_type_id, gdf.penalty_type_name, gdf.count, gdf.quantity_scale
FROM game_days gd
JOIN game_day_participants gdp ON gdp.game_day_id = gd.id AND gdp.player_id = $2
LEFT JOIN game_day_fees gdf ON gdf.game_day_participant_id = gdp.id
WHERE gd.club_id = $1 AND gd.date >= $3
ORDER BY gd.date ASC, gdf.penalty_type_name ASC;

-- ==================== PLAYER COMPETITION HISTORY (DASHBOARD) ====================

-- name: GetPlayerCompetitionHistoryByClubAndPlayer :many
SELECT gd.id as game_day_id, gd.date, gdcv.competition_id, c.name as competition_name, gdcv.value
FROM game_days gd
JOIN game_day_participants gdp ON gdp.game_day_id = gd.id AND gdp.player_id = $2
JOIN game_day_competition_values gdcv ON gdcv.game_day_participant_id = gdp.id
JOIN competitions c ON c.id = gdcv.competition_id
WHERE gd.club_id = $1 AND gd.date >= $3
ORDER BY gd.date ASC, c.display_order ASC, c.name ASC;
