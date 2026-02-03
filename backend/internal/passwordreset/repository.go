package passwordreset

import (
	"context"
	"database/sql"
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

func (r *Repository) Create(ctx context.Context, userID string, token string, expiresAt time.Time) (db.PasswordResetToken, error) {
	return r.queries.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
		ID:        uuid.NewString(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	})
}

func (r *Repository) GetByToken(ctx context.Context, token string) (db.PasswordResetToken, error) {
	t, err := r.queries.GetPasswordResetTokenByToken(ctx, token)
	if errors.Is(err, sql.ErrNoRows) {
		return db.PasswordResetToken{}, ErrNotFound
	}
	if err != nil {
		return db.PasswordResetToken{}, err
	}
	return t, nil
}

func (r *Repository) DeleteByToken(ctx context.Context, token string) error {
	return r.queries.DeletePasswordResetToken(ctx, token)
}

func (r *Repository) DeleteByUserID(ctx context.Context, userID string) error {
	return r.queries.DeletePasswordResetTokensByUserID(ctx, userID)
}
