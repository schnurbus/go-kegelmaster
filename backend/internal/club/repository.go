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
	Name           string
	Balance        int
	StartBalance   int
	BaseFee        int
	AutoTipEnabled bool
	UserID         string
}

func (r *Repository) Create(ctx context.Context, params CreateClubParams) (Club, error) {
	now := time.Now().UTC()
	id := uuid.NewString()

	dbClub, err := r.queries.CreateClub(ctx, db.CreateClubParams{
		ID:             id,
		Name:           params.Name,
		Balance:        int32(params.Balance),
		StartBalance:   int32(params.StartBalance),
		BaseFee:        int32(params.BaseFee),
		AutoTipEnabled: params.AutoTipEnabled,
		UserID:         params.UserID,
		CreatedAt:      now,
		UpdatedAt:      now,
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

func (r *Repository) GetForUser(ctx context.Context, userID string) ([]Club, error) {
	dbClubs, err := r.queries.GetClubsForUser(ctx, userID)
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
	ID             string
	Name           string
	Balance        int
	StartBalance   int
	BaseFee        int
	AutoTipEnabled bool
}

func (r *Repository) Update(ctx context.Context, params UpdateClubParams) (Club, error) {
	now := time.Now().UTC()

	dbClub, err := r.queries.UpdateClub(ctx, db.UpdateClubParams{
		ID:             params.ID,
		Name:           params.Name,
		Balance:        int32(params.Balance),
		StartBalance:   int32(params.StartBalance),
		BaseFee:        int32(params.BaseFee),
		AutoTipEnabled: params.AutoTipEnabled,
		UpdatedAt:      now,
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

// UpdateBalance updates only the balance and updated_at of a club (e.g. after recalculation).
func (r *Repository) UpdateBalance(ctx context.Context, id string, balance int) error {
	now := time.Now().UTC()
	return r.queries.UpdateClubBalance(ctx, db.UpdateClubBalanceParams{
		Balance:   int32(balance),
		UpdatedAt: now,
		ID:        id,
	})
}

// dbClubToClub converts a db.Club to a club.Club
func dbClubToClub(dbClub db.Club) Club {
	return Club{
		ID:             dbClub.ID,
		Name:           dbClub.Name,
		Balance:        int(dbClub.Balance),
		StartBalance:   int(dbClub.StartBalance),
		BaseFee:        int(dbClub.BaseFee),
		AutoTipEnabled: dbClub.AutoTipEnabled,
		UserID:         dbClub.UserID,
		CreatedAt:      dbClub.CreatedAt,
		UpdatedAt:      dbClub.UpdatedAt,
	}
}
