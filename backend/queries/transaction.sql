-- ==================== TRANSACTIONS ====================

-- name: CreateTransaction :one
INSERT INTO transactions (
    id, club_id, player_id, transaction_type,
    amount, description, game_day_fee_id, game_day_id,
    player_balance_before, player_balance_after,
    club_balance_before, club_balance_after,
    created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING *;

-- name: GetTransactionByID :one
SELECT t.*, p.name as player_name
FROM transactions t
LEFT JOIN players p ON p.id = t.player_id
WHERE t.id = $1;

-- name: ListTransactionsByClub :many
SELECT t.*, p.name as player_name
FROM transactions t
LEFT JOIN players p ON p.id = t.player_id
WHERE t.club_id = $1
ORDER BY t.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountTransactionsByClub :one
SELECT COUNT(*) FROM transactions WHERE club_id = $1;

-- name: ListTransactionsByPlayer :many
SELECT t.*, p.name as player_name
FROM transactions t
LEFT JOIN players p ON p.id = t.player_id
WHERE t.club_id = $1 AND t.player_id = $2
ORDER BY t.created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountTransactionsByPlayer :one
SELECT COUNT(*) FROM transactions 
WHERE club_id = $1 AND player_id = $2;

-- name: ListTransactionsByGameDay :many
SELECT t.*, p.name as player_name
FROM transactions t
LEFT JOIN players p ON p.id = t.player_id
WHERE t.game_day_id = $1
ORDER BY t.created_at ASC;

-- name: ListTransactionsByType :many
SELECT t.*, p.name as player_name
FROM transactions t
LEFT JOIN players p ON p.id = t.player_id
WHERE t.club_id = $1 AND t.transaction_type = $2
ORDER BY t.created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountTransactionsByType :one
SELECT COUNT(*) FROM transactions 
WHERE club_id = $1 AND transaction_type = $2;

-- name: UpdateTransaction :one
UPDATE transactions
SET amount = $2, description = $3, updated_at = $4
WHERE id = $1
RETURNING *;

-- name: DeleteTransaction :exec
DELETE FROM transactions WHERE id = $1;

-- name: GetTransactionByGameDayFee :one
SELECT * FROM transactions WHERE game_day_fee_id = $1;

-- name: SumTransactionsByPlayer :one
SELECT COALESCE(SUM(
    CASE 
        WHEN transaction_type IN ('base_fee', 'fee') THEN amount
        WHEN transaction_type = 'deposit' THEN -amount
        ELSE 0
    END
), 0) as total
FROM transactions
WHERE player_id = $1;

-- name: SumTransactionsByClub :one
SELECT COALESCE(SUM(
    CASE 
        WHEN transaction_type IN ('deposit', 'tip') THEN amount
        WHEN transaction_type = 'expense' THEN amount
        ELSE 0
    END
), 0) as total
FROM transactions
WHERE club_id = $1;

-- name: GetGameDayTransactionSummary :one
SELECT 
    COALESCE(SUM(CASE WHEN transaction_type = 'base_fee' THEN ABS(amount) ELSE 0 END), 0) as base_fee_total,
    COALESCE(SUM(CASE WHEN transaction_type = 'fee' THEN ABS(amount) ELSE 0 END), 0) as fee_total,
    COUNT(CASE WHEN transaction_type = 'base_fee' THEN 1 END) as base_fee_count,
    COUNT(CASE WHEN transaction_type = 'fee' THEN 1 END) as fee_count
FROM transactions
WHERE game_day_id = $1;
