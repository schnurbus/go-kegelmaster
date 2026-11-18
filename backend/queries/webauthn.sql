-- name: CreateWebAuthnCredential :one
INSERT INTO webauthn_credentials (
    id,
    user_id,
    name,
    credential_id,
    public_key,
    attestation_type,
    aaguid,
    sign_count,
    transports,
    flags,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11
)
RETURNING *;

-- name: GetWebAuthnCredentialByCredentialID :one
SELECT * FROM webauthn_credentials
WHERE credential_id = $1;

-- name: ListWebAuthnCredentialsByUserID :many
SELECT * FROM webauthn_credentials
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CountWebAuthnCredentialsByUserID :one
SELECT COUNT(*) FROM webauthn_credentials
WHERE user_id = $1;

-- name: UpdateWebAuthnCredentialAfterLogin :exec
UPDATE webauthn_credentials
SET sign_count = $2, flags = $3, last_used_at = $4, updated_at = $4
WHERE id = $1;

-- name: RenameWebAuthnCredential :one
UPDATE webauthn_credentials
SET name = $3, updated_at = $4
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteWebAuthnCredential :execrows
DELETE FROM webauthn_credentials
WHERE id = $1 AND user_id = $2;

-- name: DeleteWebAuthnCredentialsByUserID :exec
DELETE FROM webauthn_credentials
WHERE user_id = $1;

-- name: CreateWebAuthnSession :exec
INSERT INTO webauthn_sessions (
    id,
    ceremony,
    user_id,
    session_data,
    expires_at,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $6
);

-- name: ConsumeWebAuthnSession :one
-- Deleting and returning in one statement makes each challenge single-use.
DELETE FROM webauthn_sessions
WHERE id = $1 AND ceremony = $2
RETURNING *;

-- name: DeleteExpiredWebAuthnSessions :exec
DELETE FROM webauthn_sessions
WHERE expires_at < $1;
