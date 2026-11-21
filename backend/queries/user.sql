-- name: CreateUser :one
INSERT INTO users (
    id,
    email,
    password_hash,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, password_changed_at = $3, updated_at = $3
WHERE id = $1;

-- name: UpdateUserNotifyGameDaySettled :one
UPDATE users
SET notify_gameday_settled = $2, updated_at = $3
WHERE id = $1
RETURNING *;


-- name: MarkUserTourFlags :one
-- Sets "Rundgang gesehen" / "Hinweis zur Ersteinrichtung gesehen" once; an
-- existing timestamp is kept.
UPDATE users
SET tour_seen_at = CASE WHEN sqlc.arg(tour_seen)::bool THEN COALESCE(tour_seen_at, sqlc.arg(now)::timestamptz) ELSE tour_seen_at END,
    setup_hint_seen_at = CASE WHEN sqlc.arg(setup_hint_seen)::bool THEN COALESCE(setup_hint_seen_at, sqlc.arg(now)::timestamptz) ELSE setup_hint_seen_at END
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateUserLastLogin :exec
-- Does not touch updated_at: a login is no change to the account itself.
UPDATE users
SET last_login_at = $2
WHERE id = $1;
