-- name: CreateAPIToken :one
INSERT INTO api_tokens (
    id,
    user_id,
    name,
    token_hash,
    token_prefix,
    expires_at,
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
    $7
)
RETURNING *;

-- name: GetAPITokenByHash :one
SELECT * FROM api_tokens
WHERE token_hash = $1;

-- name: ListAPITokensByUserID :many
SELECT * FROM api_tokens
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CountAPITokensByUserID :one
SELECT COUNT(*) FROM api_tokens
WHERE user_id = $1;

-- name: DeleteAPIToken :execrows
DELETE FROM api_tokens
WHERE id = $1 AND user_id = $2;

-- name: DeleteAPITokensByUserID :exec
DELETE FROM api_tokens
WHERE user_id = $1;

-- name: TouchAPIToken :exec
-- Throttled to one write per minute per token so busy clients do not hammer the row.
UPDATE api_tokens
SET last_used_at = sqlc.arg(used_at)::timestamptz
WHERE id = sqlc.arg(id)
  AND (last_used_at IS NULL OR last_used_at < sqlc.arg(used_at)::timestamptz - INTERVAL '1 minute');
