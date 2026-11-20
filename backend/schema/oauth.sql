-- Mirrors migration 000032_create_oauth (sqlc schema only).

-- Dynamically registered public clients. id is the public client_id.
CREATE TABLE IF NOT EXISTS oauth_clients (
    id UUID PRIMARY KEY,
    client_name TEXT NOT NULL,
    redirect_uris TEXT[] NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- A grant is one approved connection of a user with a client ("Verbundene
-- Anwendung"). All tokens of a refresh chain belong to one grant; deleting
-- the grant revokes the whole chain.
CREATE TABLE IF NOT EXISTS oauth_grants (
    id UUID PRIMARY KEY,
    client_id UUID NOT NULL REFERENCES oauth_clients(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scope TEXT NOT NULL,
    resource TEXT NOT NULL,
    last_used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_oauth_grants_user_id ON oauth_grants(user_id);
CREATE INDEX IF NOT EXISTS idx_oauth_grants_client_id ON oauth_grants(client_id);

-- Authorization codes: single use (used_at), short-lived, bound to client,
-- redirect_uri and PKCE challenge. grant_id is set once the code was
-- exchanged, so a replayed code can revoke the tokens issued for it.
CREATE TABLE IF NOT EXISTS oauth_authorization_codes (
    id UUID PRIMARY KEY,
    code_hash TEXT NOT NULL UNIQUE,
    client_id UUID NOT NULL REFERENCES oauth_clients(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    redirect_uri TEXT NOT NULL,
    code_challenge TEXT NOT NULL,
    scope TEXT NOT NULL,
    resource TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL,
    grant_id UUID NULL REFERENCES oauth_grants(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_oauth_authorization_codes_expires_at ON oauth_authorization_codes(expires_at);

-- Access and refresh tokens. Refresh tokens rotate: used_at marks a refresh
-- token that was exchanged; presenting it again revokes the grant.
CREATE TABLE IF NOT EXISTS oauth_tokens (
    id UUID PRIMARY KEY,
    grant_id UUID NOT NULL REFERENCES oauth_grants(id) ON DELETE CASCADE,
    token_type TEXT NOT NULL CHECK (token_type IN ('access', 'refresh')),
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_oauth_tokens_grant_id ON oauth_tokens(grant_id);
