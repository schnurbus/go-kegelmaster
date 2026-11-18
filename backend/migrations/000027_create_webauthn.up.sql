-- Passkeys (WebAuthn credentials) as an additional way to sign in.
CREATE TABLE IF NOT EXISTS webauthn_credentials (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    credential_id BYTEA NOT NULL UNIQUE,
    public_key BYTEA NOT NULL,
    attestation_type TEXT NOT NULL DEFAULT '',
    aaguid BYTEA NULL,
    sign_count BIGINT NOT NULL DEFAULT 0,
    -- Comma-separated authenticator transports (usb, nfc, ble, internal, hybrid, ...).
    transports TEXT NOT NULL DEFAULT '',
    -- Raw authenticator flags byte (UP, UV, BE, BS, ...), see webauthn.NewCredentialFlags.
    flags SMALLINT NOT NULL DEFAULT 0,
    last_used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_webauthn_credentials_user_id ON webauthn_credentials(user_id);

-- Pending WebAuthn ceremonies. A row holds the challenge between "begin" and
-- "finish" and is deleted when consumed, so every challenge is single-use.
CREATE TABLE IF NOT EXISTS webauthn_sessions (
    id UUID PRIMARY KEY,
    ceremony TEXT NOT NULL CHECK (ceremony IN ('registration', 'login')),
    user_id UUID NULL REFERENCES users(id) ON DELETE CASCADE,
    session_data JSONB NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_webauthn_sessions_expires_at ON webauthn_sessions(expires_at);
