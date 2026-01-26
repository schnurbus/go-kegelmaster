package player

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
)

var (
	ErrNotFound = errors.New("player not found")
)

type Repository struct {
	queries *db.Queries
}

func NewRepository(dbConn *sql.DB) *Repository {
	return &Repository{
		queries: db.New(dbConn),
	}
}

type CreatePlayerParams struct {
	ClubID       string
	UserID       *string
	RoleID       *string
	Name         string
	Balance      int
	StartBalance int
}

func (r *Repository) Create(ctx context.Context, params CreatePlayerParams) (Player, error) {
	now := time.Now().UTC()
	id := uuid.NewString()

	dbPlayer, err := r.queries.CreatePlayer(ctx, db.CreatePlayerParams{
		ID:           id,
		ClubID:       params.ClubID,
		UserID:       params.UserID,
		RoleID:       params.RoleID,
		Name:         params.Name,
		Balance:      int32(params.Balance),
		StartBalance: int32(params.StartBalance),
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return Player{}, err
	}

	return dbPlayerToPlayer(dbPlayer), nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (Player, error) {
	dbPlayer, err := r.queries.GetPlayerByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Player{}, ErrNotFound
	}
	if err != nil {
		return Player{}, err
	}

	return dbPlayerToPlayer(dbPlayer), nil
}

func (r *Repository) GetByClubID(ctx context.Context, clubID string) ([]Player, error) {
	dbPlayers, err := r.queries.GetPlayersByClubID(ctx, clubID)
	if err != nil {
		return nil, err
	}

	players := make([]Player, len(dbPlayers))
	for i, dbPlayer := range dbPlayers {
		players[i] = dbPlayerToPlayer(dbPlayer)
	}

	return players, nil
}

func (r *Repository) GetByUserIDAndClubID(ctx context.Context, userID, clubID string) (Player, error) {
	dbPlayer, err := r.queries.GetPlayerByUserIDAndClubID(ctx, db.GetPlayerByUserIDAndClubIDParams{
		UserID: &userID,
		ClubID: clubID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return Player{}, ErrNotFound
	}
	if err != nil {
		return Player{}, err
	}

	return dbPlayerToPlayer(dbPlayer), nil
}

// dbPlayerToPlayer converts a db.Player to a player.Player
func dbPlayerToPlayer(dbPlayer db.Player) Player {
	return Player{
		ID:           dbPlayer.ID,
		ClubID:       dbPlayer.ClubID,
		UserID:       dbPlayer.UserID,
		RoleID:       dbPlayer.RoleID,
		Name:         dbPlayer.Name,
		Balance:      int(dbPlayer.Balance),
		StartBalance: int(dbPlayer.StartBalance),
		CreatedAt:    dbPlayer.CreatedAt,
		UpdatedAt:    dbPlayer.UpdatedAt,
	}
}

type UpdatePlayerParams struct {
	ID           string
	Name         string
	Balance      int
	StartBalance int
	UserID       *string
	RoleID       *string
}

func (r *Repository) Update(ctx context.Context, params UpdatePlayerParams) (Player, error) {
	now := time.Now().UTC()

	dbPlayer, err := r.queries.UpdatePlayer(ctx, db.UpdatePlayerParams{
		ID:           params.ID,
		Name:         params.Name,
		Balance:      int32(params.Balance),
		StartBalance: int32(params.StartBalance),
		UserID:       params.UserID,
		RoleID:       params.RoleID,
		UpdatedAt:    now,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return Player{}, ErrNotFound
	}
	if err != nil {
		return Player{}, err
	}

	return dbPlayerToPlayer(dbPlayer), nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	// Check if player exists first
	_, err := r.queries.GetPlayerByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	// Delete the player
	err = r.queries.DeletePlayer(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

