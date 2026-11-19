-- name: ClaimGameDaySettledNotification :execrows
-- Affects one row for the first caller only, so a game day triggers at most one notification.
INSERT INTO game_day_notifications (id, game_day_id, created_at, updated_at)
VALUES ($1, $2, $3, $3)
ON CONFLICT (game_day_id) DO NOTHING;

-- name: ListGameDaySettledRecipients :many
-- Players of the club linked to a user who opted in to game day e-mails, with
-- the sum of their bookings for the game day. Inactive players are only
-- included when they were charged on that day.
SELECT
    u.id AS user_id,
    u.email,
    p.id AS player_id,
    p.name AS player_name,
    p.balance AS player_balance,
    COALESCE(SUM(t.amount), 0)::integer AS game_day_amount
FROM players p
JOIN users u ON u.id = p.user_id
LEFT JOIN transactions t ON t.player_id = p.id AND t.game_day_id = sqlc.arg(game_day_id)
WHERE p.club_id = sqlc.arg(club_id)
  AND u.notify_gameday_settled = true
GROUP BY u.id, u.email, p.id, p.name, p.balance, p.inactive
HAVING p.inactive = false OR COUNT(t.id) > 0
ORDER BY u.email, p.name;
