package competition

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
)

type Repository struct {
	queries *db.Queries
}

func NewRepository(dbConn *sql.DB) *Repository {
	return &Repository{
		queries: db.New(dbConn),
	}
}

type CreateCompetitionParams struct {
	ClubID           string
	Name             string
	ScoringType      string // winner, loser, both
	IsGenderSpecific bool
	DisplayOrder     *int // Optional: if nil, will be set automatically
}

func (r *Repository) Create(ctx context.Context, params CreateCompetitionParams) (Competition, error) {
	now := time.Now().UTC()
	id := uuid.NewString()

	displayOrder := int32(0)
	if params.DisplayOrder == nil {
		maxOrder, err := r.queries.GetMaxCompetitionDisplayOrderByClubID(ctx, params.ClubID)
		if err != nil {
			return Competition{}, err
		}
		displayOrder = maxOrder + 1
	} else {
		displayOrder = int32(*params.DisplayOrder)
	}

	dbCompetition, err := r.queries.CreateCompetition(ctx, db.CreateCompetitionParams{
		ID:               id,
		ClubID:           params.ClubID,
		Name:             params.Name,
		ScoringType:      params.ScoringType,
		IsGenderSpecific: params.IsGenderSpecific,
		DisplayOrder:     displayOrder,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		return Competition{}, err
	}

	return dbCompetitionToCompetition(dbCompetition), nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (Competition, error) {
	dbCompetition, err := r.queries.GetCompetitionByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Competition{}, ErrNotFound
	}
	if err != nil {
		return Competition{}, err
	}

	return dbCompetitionToCompetition(dbCompetition), nil
}

func (r *Repository) GetByClubID(ctx context.Context, clubID string) ([]Competition, error) {
	dbCompetitions, err := r.queries.GetCompetitionsByClubID(ctx, clubID)
	if err != nil {
		return nil, err
	}

	competitions := make([]Competition, len(dbCompetitions))
	for i, dbCompetition := range dbCompetitions {
		competitions[i] = dbCompetitionToCompetition(dbCompetition)
	}

	return competitions, nil
}

type UpdateCompetitionParams struct {
	ID               string
	Name             string
	ScoringType      string
	IsGenderSpecific bool
	DisplayOrder     int
}

func (r *Repository) Update(ctx context.Context, params UpdateCompetitionParams) (Competition, error) {
	now := time.Now().UTC()

	dbCompetition, err := r.queries.UpdateCompetition(ctx, db.UpdateCompetitionParams{
		Name:             params.Name,
		ScoringType:      params.ScoringType,
		IsGenderSpecific: params.IsGenderSpecific,
		DisplayOrder:     int32(params.DisplayOrder),
		UpdatedAt:       now,
		ID:               params.ID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return Competition{}, ErrNotFound
	}
	if err != nil {
		return Competition{}, err
	}

	return dbCompetitionToCompetition(dbCompetition), nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.queries.GetCompetitionByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	return r.queries.DeleteCompetition(ctx, id)
}

func dbCompetitionToCompetition(dbCompetition db.Competition) Competition {
	return Competition{
		ID:               dbCompetition.ID,
		ClubID:           dbCompetition.ClubID,
		Name:             dbCompetition.Name,
		ScoringType:      dbCompetition.ScoringType,
		IsGenderSpecific: dbCompetition.IsGenderSpecific,
		DisplayOrder:     int(dbCompetition.DisplayOrder),
		CreatedAt:        dbCompetition.CreatedAt,
		UpdatedAt:        dbCompetition.UpdatedAt,
	}
}
