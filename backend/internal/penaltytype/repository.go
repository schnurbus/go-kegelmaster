package penaltytype

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

type CreatePenaltyTypeParams struct {
	ClubID                string
	Name                  string
	Description           string
	Price                 int
	DisplayOrder          *int  // Optional: if nil, will be set automatically
	AllowsDecimalQuantity bool
}

func (r *Repository) Create(ctx context.Context, params CreatePenaltyTypeParams) (PenaltyType, error) {
	now := time.Now().UTC()
	id := uuid.NewString()

	var description sql.NullString
	if params.Description != "" {
		description = sql.NullString{String: params.Description, Valid: true}
	}

	// Auto-set display_order if not provided
	displayOrder := int32(0)
	if params.DisplayOrder == nil {
		maxOrderResult, err := r.queries.GetMaxDisplayOrderByClubID(ctx, params.ClubID)
		if err != nil {
			return PenaltyType{}, err
		}
		// maxOrderResult is interface{}, but COALESCE returns int64
		maxOrder, ok := maxOrderResult.(int64)
		if !ok {
			maxOrder = 0
		}
		displayOrder = int32(maxOrder) + 1
	} else {
		displayOrder = int32(*params.DisplayOrder)
	}

	dbPenaltyType, err := r.queries.CreatePenaltyType(ctx, db.CreatePenaltyTypeParams{
		ID:                    id,
		ClubID:                params.ClubID,
		Name:                  params.Name,
		Description:           description,
		Price:                 int32(params.Price),
		DisplayOrder:          displayOrder,
		AllowsDecimalQuantity: params.AllowsDecimalQuantity,
		CreatedAt:             now,
		UpdatedAt:             now,
	})
	if err != nil {
		return PenaltyType{}, err
	}

	return dbPenaltyTypeToPenaltyType(dbPenaltyType), nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (PenaltyType, error) {
	dbPenaltyType, err := r.queries.GetPenaltyTypeByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return PenaltyType{}, ErrNotFound
	}
	if err != nil {
		return PenaltyType{}, err
	}

	return dbPenaltyTypeToPenaltyType(dbPenaltyType), nil
}

func (r *Repository) GetByClubID(ctx context.Context, clubID string) ([]PenaltyType, error) {
	dbPenaltyTypes, err := r.queries.GetPenaltyTypesByClubID(ctx, clubID)
	if err != nil {
		return nil, err
	}

	penaltyTypes := make([]PenaltyType, len(dbPenaltyTypes))
	for i, dbPenaltyType := range dbPenaltyTypes {
		penaltyTypes[i] = dbPenaltyTypeToPenaltyType(dbPenaltyType)
	}

	return penaltyTypes, nil
}

type ReplacePenaltyTypeParams struct {
	ID                    string
	Name                  string
	Description           string
	Price                 int
	AllowsDecimalQuantity bool
}

func (r *Repository) Replace(ctx context.Context, params ReplacePenaltyTypeParams) (PenaltyType, error) {
	// Get the old penalty type to verify it exists and get club_id
	oldPenaltyType, err := r.GetByID(ctx, params.ID)
	if err != nil {
		return PenaltyType{}, err
	}

	now := time.Now().UTC()
	newID := uuid.NewString()

	// 1) Soft-delete old row first so (club_id, name) is free for the unique index (WHERE deleted_at IS NULL)
	err = r.queries.MarkPenaltyTypeAsReplaced(ctx, db.MarkPenaltyTypeAsReplacedParams{
		DeletedAt:    sql.NullTime{Time: now, Valid: true},
		ReplacedByID: nil, // set after new row exists (FK)
		ID:            params.ID,
	})
	if err != nil {
		return PenaltyType{}, err
	}

	var description sql.NullString
	if params.Description != "" {
		description = sql.NullString{String: params.Description, Valid: true}
	}

	// 2) Create new penalty type (same name allowed now; old row is excluded from unique index)
	dbPenaltyType, err := r.queries.CreatePenaltyType(ctx, db.CreatePenaltyTypeParams{
		ID:                    newID,
		ClubID:                oldPenaltyType.ClubID,
		Name:                  params.Name,
		Description:           description,
		Price:                 int32(params.Price),
		DisplayOrder:          int32(oldPenaltyType.DisplayOrder),
		AllowsDecimalQuantity: params.AllowsDecimalQuantity,
		CreatedAt:             now,
		UpdatedAt:             now,
	})
	if err != nil {
		return PenaltyType{}, err
	}

	// 3) Set replaced_by_id on old row (FK now valid)
	err = r.queries.SetPenaltyTypeReplacedBy(ctx, db.SetPenaltyTypeReplacedByParams{
		ReplacedByID: &newID,
		ID:           params.ID,
	})
	if err != nil {
		return PenaltyType{}, err
	}

	return dbPenaltyTypeToPenaltyType(dbPenaltyType), nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	// Check if penalty type exists first
	_, err := r.queries.GetPenaltyTypeByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	// Soft delete the penalty type
	now := time.Now().UTC()
	err = r.queries.DeletePenaltyType(ctx, db.DeletePenaltyTypeParams{
		DeletedAt: sql.NullTime{Time: now, Valid: true},
		ID:        id,
	})
	if err != nil {
		return err
	}

	return nil
}

type UpdateDisplayOrderParams struct {
	ID           string
	DisplayOrder int
}

func (r *Repository) UpdateDisplayOrder(ctx context.Context, params UpdateDisplayOrderParams) error {
	// Check if penalty type exists first
	_, err := r.queries.GetPenaltyTypeByID(ctx, params.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	err = r.queries.UpdatePenaltyTypeDisplayOrder(ctx, db.UpdatePenaltyTypeDisplayOrderParams{
		DisplayOrder: int32(params.DisplayOrder),
		UpdatedAt:    now,
		ID:           params.ID,
	})
	if err != nil {
		return err
	}

	return nil
}

// dbPenaltyTypeToPenaltyType converts a db.PenaltyType to a penaltytype.PenaltyType
func dbPenaltyTypeToPenaltyType(dbPenaltyType db.PenaltyType) PenaltyType {
	var description string
	if dbPenaltyType.Description.Valid {
		description = dbPenaltyType.Description.String
	}

	var deletedAt *time.Time
	if dbPenaltyType.DeletedAt.Valid {
		deletedAt = &dbPenaltyType.DeletedAt.Time
	}

	replacedByID := dbPenaltyType.ReplacedByID

	return PenaltyType{
		ID:                    dbPenaltyType.ID,
		ClubID:                dbPenaltyType.ClubID,
		Name:                  dbPenaltyType.Name,
		Description:           description,
		Price:                 int(dbPenaltyType.Price),
		DisplayOrder:          int(dbPenaltyType.DisplayOrder),
		AllowsDecimalQuantity: dbPenaltyType.AllowsDecimalQuantity,
		CreatedAt:             dbPenaltyType.CreatedAt,
		UpdatedAt:             dbPenaltyType.UpdatedAt,
		DeletedAt:             deletedAt,
		ReplacedByID:          replacedByID,
	}
}
