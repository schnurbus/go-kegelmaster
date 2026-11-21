-- name: CreateOAuthClient :one
INSERT INTO oauth_clients (
    id,
    client_name,
    redirect_uris,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $4
)
RETURNING *;

-- name: GetOAuthClient :one
SELECT * FROM oauth_clients
WHERE id = $1;

-- name: CreateOAuthAuthorizationCode :exec
INSERT INTO oauth_authorization_codes (
    id,
    code_hash,
    client_id,
    user_id,
    redirect_uri,
    code_challenge,
    scope,
    resource,
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
    $8,
    $9,
    $10,
    $10
);

-- name: DeleteExpiredOAuthAuthorizationCodes :exec
-- Housekeeping: expired codes can no longer be exchanged.
DELETE FROM oauth_authorization_codes
WHERE expires_at < sqlc.arg(before)::timestamptz;

-- name: ConsumeOAuthAuthorizationCode :one
-- Marks a code as used. Only the first caller gets the row, which makes codes single use.
UPDATE oauth_authorization_codes
SET used_at = sqlc.arg(now)::timestamptz,
    updated_at = sqlc.arg(now)::timestamptz
WHERE code_hash = sqlc.arg(code_hash)
  AND used_at IS NULL
RETURNING *;

-- name: GetOAuthAuthorizationCodeByHash :one
SELECT * FROM oauth_authorization_codes
WHERE code_hash = $1;

-- name: SetOAuthAuthorizationCodeGrant :exec
UPDATE oauth_authorization_codes
SET grant_id = sqlc.arg(grant_id),
    updated_at = sqlc.arg(now)::timestamptz
WHERE id = sqlc.arg(id);

-- name: CreateOAuthGrant :one
INSERT INTO oauth_grants (
    id,
    client_id,
    user_id,
    scope,
    resource,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $6
)
RETURNING *;

-- name: GetOAuthGrant :one
SELECT * FROM oauth_grants
WHERE id = $1;

-- name: ListOAuthGrantsByUserID :many
SELECT
    g.id,
    g.client_id,
    g.scope,
    g.last_used_at,
    g.created_at,
    c.client_name,
    c.redirect_uris
FROM oauth_grants g
JOIN oauth_clients c ON c.id = g.client_id
WHERE g.user_id = $1
ORDER BY g.created_at DESC;

-- name: DeleteOAuthGrant :execrows
DELETE FROM oauth_grants
WHERE id = $1;

-- name: DeleteOAuthGrantForUser :execrows
DELETE FROM oauth_grants
WHERE id = $1 AND user_id = $2;

-- name: TouchOAuthGrant :exec
-- Throttled to one write per minute per grant so busy clients do not hammer the row.
UPDATE oauth_grants
SET last_used_at = sqlc.arg(used_at)::timestamptz
WHERE id = sqlc.arg(id)
  AND (last_used_at IS NULL OR last_used_at < sqlc.arg(used_at)::timestamptz - INTERVAL '1 minute');

-- name: CreateOAuthToken :exec
INSERT INTO oauth_tokens (
    id,
    grant_id,
    token_type,
    token_hash,
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
    $6
);

-- name: GetOAuthTokenByHash :one
SELECT
    t.id,
    t.grant_id,
    t.token_type,
    t.expires_at,
    t.used_at,
    g.client_id,
    g.user_id,
    g.created_at AS grant_created_at
FROM oauth_tokens t
JOIN oauth_grants g ON g.id = t.grant_id
WHERE t.token_hash = $1;

-- name: UseOAuthRefreshToken :one
-- Marks a refresh token as exchanged (rotation). Only the first caller gets the row.
UPDATE oauth_tokens
SET used_at = sqlc.arg(now)::timestamptz,
    updated_at = sqlc.arg(now)::timestamptz
WHERE token_hash = sqlc.arg(token_hash)
  AND token_type = 'refresh'
  AND used_at IS NULL
RETURNING id, grant_id, expires_at;

-- name: DeleteExpiredOAuthTokensForGrant :exec
DELETE FROM oauth_tokens
WHERE grant_id = sqlc.arg(grant_id)
  AND expires_at < sqlc.arg(before)::timestamptz;
