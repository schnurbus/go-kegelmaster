package oauth

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
)

var (
	// ErrNotFound: unknown client, code, token or grant (also for malformed values).
	ErrNotFound = errors.New("oauth: not found")
	// ErrExpired: the code or token is past its expiry.
	ErrExpired = errors.New("oauth: expired")
	// ErrReused: an already exchanged code or refresh token was presented
	// again. The grant issued for it has been revoked.
	ErrReused = errors.New("oauth: code or refresh token reused, grant revoked")
)

// Client is a dynamically registered public client.
type Client struct {
	ID           string
	Name         string
	RedirectURIs []string
	CreatedAt    time.Time
}

// HasRedirectURI reports whether uri exactly matches a registered redirect URI.
func (c Client) HasRedirectURI(uri string) bool {
	return slices.Contains(c.RedirectURIs, uri)
}

// Code is an authorization code without its secret.
type Code struct {
	ID            string
	ClientID      string
	UserID        string
	RedirectURI   string
	CodeChallenge string
	Scope         string
	Resource      string
	ExpiresAt     time.Time
	GrantID       *string
}

// Grant is one approved connection of a user with a client. All tokens of a
// refresh chain belong to one grant.
type Grant struct {
	ID         string
	ClientID   string
	UserID     string
	Scope      string
	Resource   string
	LastUsedAt *time.Time
	CreatedAt  time.Time
}

// GrantInfo is a grant together with its client, for "Verbundene Anwendungen".
type GrantInfo struct {
	ID           string
	ClientID     string
	ClientName   string
	RedirectURIs []string
	Scope        string
	LastUsedAt   *time.Time
	CreatedAt    time.Time
}

// AccessToken is a resolved access token.
type AccessToken struct {
	ID             string
	GrantID        string
	ClientID       string
	UserID         string
	ExpiresAt      time.Time
	GrantCreatedAt time.Time
}

// IssuedTokens are the plaintext secrets handed to the client once.
type IssuedTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	Scope        string
}

type Repository struct {
	conn    *sql.DB
	queries *db.Queries
}

func NewRepository(conn *sql.DB) *Repository {
	return &Repository{conn: conn, queries: db.New(conn)}
}

// CreateClient registers a public client.
func (r *Repository) CreateClient(ctx context.Context, name string, redirectURIs []string, now time.Time) (Client, error) {
	row, err := r.queries.CreateOAuthClient(ctx, db.CreateOAuthClientParams{
		ID:           uuid.NewString(),
		ClientName:   name,
		RedirectUris: redirectURIs,
		CreatedAt:    now,
	})
	if err != nil {
		return Client{}, err
	}
	return clientFromRow(row), nil
}

// GetClient returns a client by its client_id.
func (r *Repository) GetClient(ctx context.Context, clientID string) (Client, error) {
	if _, err := uuid.Parse(clientID); err != nil {
		return Client{}, ErrNotFound
	}
	row, err := r.queries.GetOAuthClient(ctx, clientID)
	if errors.Is(err, sql.ErrNoRows) {
		return Client{}, ErrNotFound
	}
	if err != nil {
		return Client{}, err
	}
	return clientFromRow(row), nil
}

// CreateCode stores a new authorization code and returns its plaintext.
func (r *Repository) CreateCode(ctx context.Context, c Code, now time.Time) (string, error) {
	plain, err := generate(CodePrefix)
	if err != nil {
		return "", err
	}
	// Housekeeping: expired codes can never be exchanged again.
	if err := r.queries.DeleteExpiredOAuthAuthorizationCodes(ctx, now); err != nil {
		return "", err
	}
	err = r.queries.CreateOAuthAuthorizationCode(ctx, db.CreateOAuthAuthorizationCodeParams{
		ID:            uuid.NewString(),
		CodeHash:      HashToken(plain),
		ClientID:      c.ClientID,
		UserID:        c.UserID,
		RedirectUri:   c.RedirectURI,
		CodeChallenge: c.CodeChallenge,
		Scope:         c.Scope,
		Resource:      c.Resource,
		ExpiresAt:     now.Add(CodeTTL),
		CreatedAt:     now,
	})
	if err != nil {
		return "", err
	}
	return plain, nil
}

// ConsumeCode marks a code as used and returns it; a code can be consumed only
// once. A code that was already exchanged yields ErrReused, and the grant
// issued for it is revoked (OAuth 2.1, section 4.1.3). Expiry and the binding
// to client, redirect_uri and PKCE are checked by the caller.
func (r *Repository) ConsumeCode(ctx context.Context, plain string, now time.Time) (Code, error) {
	if !strings.HasPrefix(plain, CodePrefix) {
		return Code{}, ErrNotFound
	}
	hash := HashToken(plain)
	row, err := r.queries.ConsumeOAuthAuthorizationCode(ctx, db.ConsumeOAuthAuthorizationCodeParams{Now: now, CodeHash: hash})
	if err == nil {
		return codeFromRow(row), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Code{}, err
	}
	used, err := r.queries.GetOAuthAuthorizationCodeByHash(ctx, hash)
	if errors.Is(err, sql.ErrNoRows) {
		return Code{}, ErrNotFound
	}
	if err != nil {
		return Code{}, err
	}
	if used.GrantID != nil {
		if _, err := r.queries.DeleteOAuthGrant(ctx, *used.GrantID); err != nil {
			return Code{}, err
		}
	}
	return codeFromRow(used), ErrReused
}

// IssueForCode creates the grant for an exchanged code and its first token pair.
func (r *Repository) IssueForCode(ctx context.Context, c Code, now time.Time) (Grant, IssuedTokens, error) {
	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return Grant{}, IssuedTokens{}, err
	}
	defer func() { _ = tx.Rollback() }()
	q := r.queries.WithTx(tx)

	row, err := q.CreateOAuthGrant(ctx, db.CreateOAuthGrantParams{
		ID:        uuid.NewString(),
		ClientID:  c.ClientID,
		UserID:    c.UserID,
		Scope:     c.Scope,
		Resource:  c.Resource,
		CreatedAt: now,
	})
	if err != nil {
		return Grant{}, IssuedTokens{}, err
	}
	grantID := row.ID
	if err := q.SetOAuthAuthorizationCodeGrant(ctx, db.SetOAuthAuthorizationCodeGrantParams{GrantID: &grantID, Now: now, ID: c.ID}); err != nil {
		return Grant{}, IssuedTokens{}, err
	}
	tokens, err := issueTokens(ctx, q, grantID, row.Scope, now)
	if err != nil {
		return Grant{}, IssuedTokens{}, err
	}
	if err := tx.Commit(); err != nil {
		return Grant{}, IssuedTokens{}, err
	}
	return grantFromRow(row), tokens, nil
}

// UseRefreshToken marks a refresh token as exchanged (rotation) and returns
// its grant. A refresh token that was already exchanged yields ErrReused and
// revokes the whole grant (all tokens of the chain).
func (r *Repository) UseRefreshToken(ctx context.Context, plain string, now time.Time) (Grant, error) {
	if !strings.HasPrefix(plain, RefreshTokenPrefix) {
		return Grant{}, ErrNotFound
	}
	hash := HashToken(plain)
	row, err := r.queries.UseOAuthRefreshToken(ctx, db.UseOAuthRefreshTokenParams{Now: now, TokenHash: hash})
	if errors.Is(err, sql.ErrNoRows) {
		used, err := r.queries.GetOAuthTokenByHash(ctx, hash)
		if errors.Is(err, sql.ErrNoRows) {
			return Grant{}, ErrNotFound
		}
		if err != nil {
			return Grant{}, err
		}
		if used.TokenType != "refresh" || !used.UsedAt.Valid {
			return Grant{}, ErrNotFound
		}
		if _, err := r.queries.DeleteOAuthGrant(ctx, used.GrantID); err != nil {
			return Grant{}, err
		}
		return Grant{ID: used.GrantID, ClientID: used.ClientID, UserID: used.UserID, CreatedAt: used.GrantCreatedAt}, ErrReused
	}
	if err != nil {
		return Grant{}, err
	}
	if !now.Before(row.ExpiresAt) {
		return Grant{}, ErrExpired
	}
	return r.GetGrant(ctx, row.GrantID)
}

// IssueForGrant issues a new token pair for an existing grant (refresh) and
// removes expired tokens of the grant.
func (r *Repository) IssueForGrant(ctx context.Context, g Grant, now time.Time) (IssuedTokens, error) {
	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return IssuedTokens{}, err
	}
	defer func() { _ = tx.Rollback() }()
	q := r.queries.WithTx(tx)

	tokens, err := issueTokens(ctx, q, g.ID, g.Scope, now)
	if err != nil {
		return IssuedTokens{}, err
	}
	if err := q.DeleteExpiredOAuthTokensForGrant(ctx, db.DeleteExpiredOAuthTokensForGrantParams{GrantID: g.ID, Before: now}); err != nil {
		return IssuedTokens{}, err
	}
	if err := tx.Commit(); err != nil {
		return IssuedTokens{}, err
	}
	return tokens, nil
}

func issueTokens(ctx context.Context, q *db.Queries, grantID, scope string, now time.Time) (IssuedTokens, error) {
	access, err := generate(AccessTokenPrefix)
	if err != nil {
		return IssuedTokens{}, err
	}
	refresh, err := generate(RefreshTokenPrefix)
	if err != nil {
		return IssuedTokens{}, err
	}
	for _, t := range []struct {
		typ, plain string
		ttl        time.Duration
	}{
		{"access", access, AccessTokenTTL},
		{"refresh", refresh, RefreshTokenTTL},
	} {
		if err := q.CreateOAuthToken(ctx, db.CreateOAuthTokenParams{
			ID:        uuid.NewString(),
			GrantID:   grantID,
			TokenType: t.typ,
			TokenHash: HashToken(t.plain),
			ExpiresAt: now.Add(t.ttl),
			CreatedAt: now,
		}); err != nil {
			return IssuedTokens{}, err
		}
	}
	return IssuedTokens{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(AccessTokenTTL / time.Second),
		Scope:        scope,
	}, nil
}

// AuthenticateAccessToken resolves a plaintext access token.
func (r *Repository) AuthenticateAccessToken(ctx context.Context, plain string, now time.Time) (AccessToken, error) {
	if !strings.HasPrefix(plain, AccessTokenPrefix) {
		return AccessToken{}, ErrNotFound
	}
	row, err := r.queries.GetOAuthTokenByHash(ctx, HashToken(plain))
	if errors.Is(err, sql.ErrNoRows) {
		return AccessToken{}, ErrNotFound
	}
	if err != nil {
		return AccessToken{}, err
	}
	if row.TokenType != "access" {
		return AccessToken{}, ErrNotFound
	}
	if !now.Before(row.ExpiresAt) {
		return AccessToken{}, ErrExpired
	}
	return AccessToken{
		ID:             row.ID,
		GrantID:        row.GrantID,
		ClientID:       row.ClientID,
		UserID:         row.UserID,
		ExpiresAt:      row.ExpiresAt,
		GrantCreatedAt: row.GrantCreatedAt,
	}, nil
}

// RevokeToken revokes the grant of an access or refresh token (RFC 7009) if
// the token belongs to clientID. Unknown tokens are ignored.
func (r *Repository) RevokeToken(ctx context.Context, plain, clientID string) (revokedGrantID string, err error) {
	if !strings.HasPrefix(plain, AccessTokenPrefix) && !strings.HasPrefix(plain, RefreshTokenPrefix) {
		return "", nil
	}
	row, err := r.queries.GetOAuthTokenByHash(ctx, HashToken(plain))
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if row.ClientID != clientID {
		return "", nil
	}
	if _, err := r.queries.DeleteOAuthGrant(ctx, row.GrantID); err != nil {
		return "", err
	}
	return row.GrantID, nil
}

// TouchGrant records that the grant was used. Writes are throttled in SQL.
func (r *Repository) TouchGrant(ctx context.Context, id string, now time.Time) error {
	return r.queries.TouchOAuthGrant(ctx, db.TouchOAuthGrantParams{UsedAt: now, ID: id})
}

func (r *Repository) GetGrant(ctx context.Context, id string) (Grant, error) {
	row, err := r.queries.GetOAuthGrant(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Grant{}, ErrNotFound
	}
	if err != nil {
		return Grant{}, err
	}
	return grantFromRow(row), nil
}

// DeleteGrant revokes a grant and all its tokens.
func (r *Repository) DeleteGrant(ctx context.Context, id string) error {
	_, err := r.queries.DeleteOAuthGrant(ctx, id)
	return err
}

// ListGrants returns the connected applications of userID.
func (r *Repository) ListGrants(ctx context.Context, userID string) ([]GrantInfo, error) {
	rows, err := r.queries.ListOAuthGrantsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]GrantInfo, 0, len(rows))
	for _, row := range rows {
		out = append(out, GrantInfo{
			ID:           row.ID,
			ClientID:     row.ClientID,
			ClientName:   row.ClientName,
			RedirectURIs: row.RedirectUris,
			Scope:        row.Scope,
			LastUsedAt:   timePtr(row.LastUsedAt),
			CreatedAt:    row.CreatedAt,
		})
	}
	return out, nil
}

// DeleteGrantForUser revokes a grant of userID. Grants of other users are
// reported as ErrNotFound.
func (r *Repository) DeleteGrantForUser(ctx context.Context, id, userID string) error {
	n, err := r.queries.DeleteOAuthGrantForUser(ctx, db.DeleteOAuthGrantForUserParams{ID: id, UserID: userID})
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func clientFromRow(row db.OauthClient) Client {
	return Client{ID: row.ID, Name: row.ClientName, RedirectURIs: row.RedirectUris, CreatedAt: row.CreatedAt}
}

func codeFromRow(row db.OauthAuthorizationCode) Code {
	return Code{
		ID:            row.ID,
		ClientID:      row.ClientID,
		UserID:        row.UserID,
		RedirectURI:   row.RedirectUri,
		CodeChallenge: row.CodeChallenge,
		Scope:         row.Scope,
		Resource:      row.Resource,
		ExpiresAt:     row.ExpiresAt,
		GrantID:       row.GrantID,
	}
}

func grantFromRow(row db.OauthGrant) Grant {
	return Grant{
		ID:         row.ID,
		ClientID:   row.ClientID,
		UserID:     row.UserID,
		Scope:      row.Scope,
		Resource:   row.Resource,
		LastUsedAt: timePtr(row.LastUsedAt),
		CreatedAt:  row.CreatedAt,
	}
}

func timePtr(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}
