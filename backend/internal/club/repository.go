package club

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
)

var (
	ErrNotFound = errors.New("club not found")
)

type Repository struct {
	queries *db.Queries
}

func NewRepository(dbConn *sql.DB) *Repository {
	return &Repository{
		queries: db.New(dbConn),
	}
}

type CreateClubParams struct {
	Name    string
	Balance int
	BaseFee int
	UserID  string
}

func (r *Repository) Create(ctx context.Context, params CreateClubParams) (Club, error) {
	now := time.Now().UTC()
	id := uuid.NewString()

	dbClub, err := r.queries.CreateClub(ctx, db.CreateClubParams{
		ID:        id,
		Name:      params.Name,
		Balance:   int32(params.Balance),
		BaseFee:   int32(params.BaseFee),
		UserID:    params.UserID,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return Club{}, err
	}

	return dbClubToClub(dbClub), nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (Club, error) {
	dbClub, err := r.queries.GetClubByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Club{}, ErrNotFound
	}
	if err != nil {
		return Club{}, err
	}

	return dbClubToClub(dbClub), nil
}

func (r *Repository) GetAll(ctx context.Context) ([]Club, error) {
	dbClubs, err := r.queries.GetAllClubs(ctx)
	if err != nil {
		return nil, err
	}

	clubs := make([]Club, len(dbClubs))
	for i, dbClub := range dbClubs {
		clubs[i] = dbClubToClub(dbClub)
	}

	return clubs, nil
}

func (r *Repository) GetByUserID(ctx context.Context, userID string) ([]Club, error) {
	dbClubs, err := r.queries.GetClubsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	clubs := make([]Club, len(dbClubs))
	for i, dbClub := range dbClubs {
		clubs[i] = dbClubToClub(dbClub)
	}

	return clubs, nil
}

type UpdateClubParams struct {
	ID      string
	Name    string
	Balance int
	BaseFee int
}

func (r *Repository) Update(ctx context.Context, params UpdateClubParams) (Club, error) {
	now := time.Now().UTC()

	dbClub, err := r.queries.UpdateClub(ctx, db.UpdateClubParams{
		ID:        params.ID,
		Name:      params.Name,
		Balance:   int32(params.Balance),
		BaseFee:   int32(params.BaseFee),
		UpdatedAt: now,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return Club{}, ErrNotFound
	}
	if err != nil {
		return Club{}, err
	}

	return dbClubToClub(dbClub), nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	// Check if club exists first
	_, err := r.queries.GetClubByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	// Delete the club
	err = r.queries.DeleteClub(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

// dbClubToClub converts a db.Club to a club.Club
func dbClubToClub(dbClub db.Club) Club {
	return Club{
		ID:        dbClub.ID,
		Name:      dbClub.Name,
		Balance:   int(dbClub.Balance),
		BaseFee:   int(dbClub.BaseFee),
		UserID:    dbClub.UserID,
		CreatedAt: dbClub.CreatedAt,
		UpdatedAt: dbClub.UpdatedAt,
	}
}
