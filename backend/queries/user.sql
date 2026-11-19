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


-- name: UpdateUserLastLogin :exec
-- Does not touch updated_at: a login is no change to the account itself.
UPDATE users
SET last_login_at = $2
WHERE id = $1;
