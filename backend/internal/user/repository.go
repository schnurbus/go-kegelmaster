package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
)

var (
	ErrNotFound      = errors.New("user not found")
	ErrEmailConflict = errors.New("user email already exists")
)

type Repository struct {
	queries *db.Queries
}

func NewRepository(dbConn *sql.DB) *Repository {
	return &Repository{
		queries: db.New(dbConn),
	}
}

type CreateUserParams struct {
	Email        string
	PasswordHash string
}

func (r *Repository) Create(ctx context.Context, params CreateUserParams) (User, error) {
	now := time.Now().UTC()
	id := uuid.NewString()

	dbUser, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		ID:           id,
		Email:        params.Email,
		PasswordHash: params.PasswordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, ErrEmailConflict
		}
		return User{}, err
	}

	return dbUserToUser(dbUser), nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (User, error) {
	dbUser, err := r.queries.GetUserByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}

	return dbUserToUser(dbUser), nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (User, error) {
	dbUser, err := r.queries.GetUserByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}

	return dbUserToUser(dbUser), nil
}

// dbUserToUser converts a db.User to a user.User
func dbUserToUser(dbUser db.User) User {
	return User{
		ID:           dbUser.ID,
		Email:        dbUser.Email,
		PasswordHash: dbUser.PasswordHash,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
