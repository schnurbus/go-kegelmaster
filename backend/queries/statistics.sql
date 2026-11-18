-- ==================== PLAYER STATISTICS ====================
-- All statistics queries only consider published game days (is_draft = false)
-- within the inclusive date range [$2, $3].

-- name: StatsListGameDays :many
SELECT gd.id, gd.date
FROM game_days gd
WHERE gd.club_id = $1 AND gd.is_draft = false AND gd.date >= $2 AND gd.date <= $3
ORDER BY gd.date ASC, gd.id ASC;

-- name: StatsListParticipants :many
SELECT gdp.id AS participant_id, gdp.game_day_id, gdp.player_id, p.gender
FROM game_day_participants gdp
JOIN game_days gd ON gd.id = gdp.game_day_id
JOIN players p ON p.id = gdp.player_id
WHERE gd.club_id = $1 AND gd.is_draft = false AND gd.date >= $2 AND gd.date <= $3;

-- name: StatsListFees :many
SELECT gdf.game_day_participant_id, gdf.penalty_type_id, gdf.penalty_type_name,
       gdf.penalty_type_price, gdf.count, gdf.quantity_scale
FROM game_day_fees gdf
JOIN game_day_participants gdp ON gdp.id = gdf.game_day_participant_id
JOIN game_days gd ON gd.id = gdp.game_day_id
WHERE gd.club_id = $1 AND gd.is_draft = false AND gd.date >= $2 AND gd.date <= $3;

-- name: StatsListCompetitionValues :many
SELECT gdcv.game_day_participant_id, gdcv.competition_id, gdcv.value
FROM game_day_competition_values gdcv
JOIN game_day_participants gdp ON gdp.id = gdcv.game_day_participant_id
JOIN game_days gd ON gd.id = gdp.game_day_id
WHERE gd.club_id = $1 AND gd.is_draft = false AND gd.date >= $2 AND gd.date <= $3;

-- StatsPlayerBalanceDeltaBefore sums balance-relevant transactions (base_fee, fee, deposit) before a date.
-- name: StatsPlayerBalanceDeltaBefore :one
SELECT COALESCE(SUM(amount), 0)::bigint AS total
FROM transactions
WHERE player_id = $1 AND transaction_type IN ('base_fee', 'fee', 'deposit') AND transaction_date < $2;

-- name: StatsListPlayerTransactions :many
SELECT transaction_date, transaction_type, amount
FROM transactions
WHERE player_id = $1 AND transaction_type IN ('base_fee', 'fee', 'deposit')
  AND transaction_date >= $2 AND transaction_date <= $3
ORDER BY transaction_date ASC, created_at ASC;
