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
	Name               string
	Balance            int
	StartBalance       int
	BaseFee            int
	AutoTipEnabled     bool
	CouplesModeEnabled bool
	UserID             string
}

func (r *Repository) Create(ctx context.Context, params CreateClubParams) (Club, error) {
	now := time.Now().UTC()
	id := uuid.NewString()

	dbClub, err := r.queries.CreateClub(ctx, db.CreateClubParams{
		ID:                 id,
		Name:               params.Name,
		Balance:            int32(params.Balance),
		StartBalance:       int32(params.StartBalance),
		BaseFee:            int32(params.BaseFee),
		AutoTipEnabled:     params.AutoTipEnabled,
		CouplesModeEnabled: params.CouplesModeEnabled,
		UserID:             params.UserID,
		CreatedAt:          now,
		UpdatedAt:          now,
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
	ID                 string
	Name               string
	Balance            int
	StartBalance       int
	BaseFee            int
	AutoTipEnabled     bool
	CouplesModeEnabled bool
}

func (r *Repository) Update(ctx context.Context, params UpdateClubParams) (Club, error) {
	now := time.Now().UTC()

	dbClub, err := r.queries.UpdateClub(ctx, db.UpdateClubParams{
		ID:                 params.ID,
		Name:               params.Name,
		Balance:            int32(params.Balance),
		StartBalance:       int32(params.StartBalance),
		BaseFee:            int32(params.BaseFee),
		AutoTipEnabled:     params.AutoTipEnabled,
		CouplesModeEnabled: params.CouplesModeEnabled,
		UpdatedAt:          now,
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

// UpdateOwner sets the club's owner (user_id) to the given new user ID.
func (r *Repository) UpdateOwner(ctx context.Context, clubID, newUserID string) (Club, error) {
	now := time.Now().UTC()
	dbClub, err := r.queries.UpdateClubOwner(ctx, db.UpdateClubOwnerParams{
		UserID:    newUserID,
		UpdatedAt: now,
		ID:        clubID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return Club{}, ErrNotFound
	}
	if err != nil {
		return Club{}, err
	}
	return dbClubToClub(dbClub), nil
}

// dbClubToClub converts a db.Club to a club.Club
func dbClubToClub(dbClub db.Club) Club {
	c := Club{
		ID:                 dbClub.ID,
		Name:               dbClub.Name,
		Balance:            int(dbClub.Balance),
		StartBalance:       int(dbClub.StartBalance),
		BaseFee:            int(dbClub.BaseFee),
		AutoTipEnabled:     dbClub.AutoTipEnabled,
		CouplesModeEnabled: dbClub.CouplesModeEnabled,
		UserID:             dbClub.UserID,
		CreatedAt:          dbClub.CreatedAt,
		UpdatedAt:          dbClub.UpdatedAt,
	}
	if dbClub.SetupDismissedAt.Valid {
		dismissed := dbClub.SetupDismissedAt.Time
		c.SetupDismissedAt = &dismissed
	}
	return c
}

// SetupCounts counts what the Ersteinrichtung of a club needs; OwnPlayers
// refers to the players linked to userID.
func (r *Repository) SetupCounts(ctx context.Context, clubID, userID string) (SetupCounts, error) {
	row, err := r.queries.GetClubSetupCounts(ctx, db.GetClubSetupCountsParams{ClubID: clubID, UserID: &userID})
	if err != nil {
		return SetupCounts{}, err
	}
	return SetupCounts{
		PenaltyTypes:      row.PenaltyTypes,
		Roles:             row.Roles,
		Players:           row.Players,
		PublishedGameDays: row.PublishedGameDays,
		Invitations:       row.Invitations,
		Competitions:      row.Competitions,
		OwnPlayers:        row.OwnPlayers,
	}, nil
}

// DismissSetup hides the Ersteinrichtung card of a club.
func (r *Repository) DismissSetup(ctx context.Context, clubID string) error {
	return r.queries.DismissClubSetup(ctx, db.DismissClubSetupParams{Now: time.Now().UTC(), ID: clubID})
}
