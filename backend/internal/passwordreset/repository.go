package passwordreset

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
)

var ErrNotFound = errors.New("password reset token not found")

type Repository struct {
	queries *db.Queries
}

func NewRepository(dbConn *sql.DB) *Repository {
	return &Repository{
		queries: db.New(dbConn),
	}
}

// HashToken returns the value stored in the database for a reset token.
// Only the hash is persisted, so a database leak does not expose usable links.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Create stores the hash of token for userID.
func (r *Repository) Create(ctx context.Context, userID string, token string, expiresAt time.Time) (db.PasswordResetToken, error) {
	return r.queries.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
		ID:        uuid.NewString(),
		UserID:    userID,
		Token:     HashToken(token),
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	})
}

// Consume deletes the reset token and returns it. The delete makes the token
// single-use even when two requests race with the same link.
func (r *Repository) Consume(ctx context.Context, token string) (db.PasswordResetToken, error) {
	t, err := r.queries.ConsumePasswordResetToken(ctx, HashToken(token))
	if errors.Is(err, sql.ErrNoRows) {
		return db.PasswordResetToken{}, ErrNotFound
	}
	if err != nil {
		return db.PasswordResetToken{}, err
	}
	return t, nil
}

func (r *Repository) DeleteByUserID(ctx context.Context, userID string) error {
	return r.queries.DeletePasswordResetTokensByUserID(ctx, userID)
}
