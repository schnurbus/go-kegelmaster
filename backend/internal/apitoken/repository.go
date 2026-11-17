// Package apitoken manages personal API tokens that let non-browser clients
// (scripts, agents) authenticate with an "Authorization: Bearer <token>" header.
package apitoken

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
)

const (
	// Prefix marks a string as a Kegelmaster API token, which makes leaked
	// tokens easy to recognise (e.g. by secret scanners).
	Prefix = "kgm_"
	// displayPrefixLen is how much of the plaintext token is stored for display.
	displayPrefixLen = len(Prefix) + 8
	randomBytes      = 32
)

var (
	ErrNotFound = errors.New("api token not found")
	ErrExpired  = errors.New("api token expired")
)

// Token is an API token without its secret.
type Token struct {
	ID         string
	UserID     string
	Name       string
	Prefix     string
	ExpiresAt  *time.Time
	LastUsedAt *time.Time
	CreatedAt  time.Time
}

type Repository struct {
	queries *db.Queries
}

func NewRepository(dbConn *sql.DB) *Repository {
	return &Repository{queries: db.New(dbConn)}
}

// HashToken returns the value stored in the database for a token.
// Tokens carry 256 bits of randomness, so a plain SHA-256 is sufficient.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func generate() (string, error) {
	buf := make([]byte, randomBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return Prefix + base64.RawURLEncoding.EncodeToString(buf), nil
}

// Create stores a new token for userID and returns it together with the
// plaintext secret. The plaintext cannot be recovered later.
func (r *Repository) Create(ctx context.Context, userID, name string, expiresAt *time.Time) (Token, string, error) {
	plain, err := generate()
	if err != nil {
		return Token{}, "", err
	}
	row, err := r.queries.CreateAPIToken(ctx, db.CreateAPITokenParams{
		ID:          uuid.NewString(),
		UserID:      userID,
		Name:        name,
		TokenHash:   HashToken(plain),
		TokenPrefix: plain[:displayPrefixLen],
		ExpiresAt:   nullTime(expiresAt),
		CreatedAt:   time.Now().UTC(),
	})
	if err != nil {
		return Token{}, "", err
	}
	return fromRow(row), plain, nil
}

// Authenticate resolves a plaintext token. It returns ErrNotFound for unknown
// or malformed tokens and ErrExpired for tokens past their expiry.
func (r *Repository) Authenticate(ctx context.Context, plain string, now time.Time) (Token, error) {
	if !strings.HasPrefix(plain, Prefix) {
		return Token{}, ErrNotFound
	}
	row, err := r.queries.GetAPITokenByHash(ctx, HashToken(plain))
	if errors.Is(err, sql.ErrNoRows) {
		return Token{}, ErrNotFound
	}
	if err != nil {
		return Token{}, err
	}
	t := fromRow(row)
	if t.ExpiresAt != nil && !now.Before(*t.ExpiresAt) {
		return Token{}, ErrExpired
	}
	return t, nil
}

// Touch records that the token was used. Writes are throttled in SQL.
func (r *Repository) Touch(ctx context.Context, id string, now time.Time) error {
	return r.queries.TouchAPIToken(ctx, db.TouchAPITokenParams{
		ID:     id,
		UsedAt: now,
	})
}

func (r *Repository) ListByUserID(ctx context.Context, userID string) ([]Token, error) {
	rows, err := r.queries.ListAPITokensByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]Token, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromRow(row))
	}
	return out, nil
}

func (r *Repository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	return r.queries.CountAPITokensByUserID(ctx, userID)
}

// Delete removes a token of userID. Tokens of other users are reported as ErrNotFound.
func (r *Repository) Delete(ctx context.Context, id, userID string) error {
	n, err := r.queries.DeleteAPIToken(ctx, db.DeleteAPITokenParams{ID: id, UserID: userID})
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteByUserID(ctx context.Context, userID string) error {
	return r.queries.DeleteAPITokensByUserID(ctx, userID)
}

func fromRow(row db.ApiToken) Token {
	return Token{
		ID:         row.ID,
		UserID:     row.UserID,
		Name:       row.Name,
		Prefix:     row.TokenPrefix,
		ExpiresAt:  timePtr(row.ExpiresAt),
		LastUsedAt: timePtr(row.LastUsedAt),
		CreatedAt:  row.CreatedAt,
	}
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func timePtr(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}
