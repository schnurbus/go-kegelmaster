package penaltytype

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	penaltyTypeID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT COALESCE\(MAX\(display_order\)`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"max_display_order"}).AddRow(int64(0)))

	mock.ExpectQuery(`INSERT INTO penalty_types`).
		WithArgs(sqlmock.AnyArg(), clubID, "Test Penalty", sqlmock.AnyArg(), 500, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "description", "price", "display_order", "allows_decimal_quantity", "created_at", "updated_at", "deleted_at", "replaced_by_id"}).
			AddRow(penaltyTypeID, clubID, "Test Penalty", "Test Description", 500, 1, false, now, now, nil, nil))

	ctx := context.Background()
	penaltyType, err := repo.Create(ctx, CreatePenaltyTypeParams{
		ClubID:      clubID,
		Name:        "Test Penalty",
		Description: "Test Description",
		Price:       500,
	})

	if err != nil {
		t.Fatalf("create penalty type: %v", err)
	}

	if penaltyType.ID != penaltyTypeID {
		t.Fatalf("expected id %s, got %s", penaltyTypeID, penaltyType.ID)
	}
	if penaltyType.Name != "Test Penalty" {
		t.Fatalf("expected name 'Test Penalty', got %s", penaltyType.Name)
	}
	if penaltyType.Description != "Test Description" {
		t.Fatalf("expected description 'Test Description', got %s", penaltyType.Description)
	}
	if penaltyType.Price != 500 {
		t.Fatalf("expected price 500, got %d", penaltyType.Price)
	}
	if penaltyType.DisplayOrder != 1 {
		t.Fatalf("expected display_order 1, got %d", penaltyType.DisplayOrder)
	}
	if penaltyType.ClubID != clubID {
		t.Fatalf("expected club_id %s, got %s", clubID, penaltyType.ClubID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_Create_NoDescription(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	penaltyTypeID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT COALESCE\(MAX\(display_order\)`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"max_display_order"}).AddRow(int64(0)))

	mock.ExpectQuery(`INSERT INTO penalty_types`).
		WithArgs(sqlmock.AnyArg(), clubID, "Test Penalty", sqlmock.AnyArg(), 500, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "description", "price", "display_order", "allows_decimal_quantity", "created_at", "updated_at", "deleted_at", "replaced_by_id"}).
			AddRow(penaltyTypeID, clubID, "Test Penalty", nil, 500, 1, false, now, now, nil, nil))

	ctx := context.Background()
	penaltyType, err := repo.Create(ctx, CreatePenaltyTypeParams{
		ClubID:      clubID,
		Name:        "Test Penalty",
		Description: "",
		Price:       500,
	})

	if err != nil {
		t.Fatalf("create penalty type: %v", err)
	}

	if penaltyType.Description != "" {
		t.Fatalf("expected empty description, got %s", penaltyType.Description)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	penaltyTypeID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`.*SELECT id, club_id, name, description, price, display_order, allows_decimal_quantity, created_at, updated_at, deleted_at, replaced_by_id FROM penalty_types.*`).
		WithArgs(penaltyTypeID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "description", "price", "display_order", "allows_decimal_quantity", "created_at", "updated_at", "deleted_at", "replaced_by_id"}).
			AddRow(penaltyTypeID, clubID, "Test Penalty", "Test Description", 500, 1, false, now, now, nil, nil))

	ctx := context.Background()
	penaltyType, err := repo.GetByID(ctx, penaltyTypeID)

	if err != nil {
		t.Fatalf("get penalty type: %v", err)
	}

	if penaltyType.ID != penaltyTypeID {
		t.Fatalf("expected id %s, got %s", penaltyTypeID, penaltyType.ID)
	}
	if penaltyType.Name != "Test Penalty" {
		t.Fatalf("expected name 'Test Penalty', got %s", penaltyType.Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	penaltyTypeID := uuid.NewString()

	mock.ExpectQuery(`.*SELECT id, club_id, name, description, price, display_order, allows_decimal_quantity, created_at, updated_at, deleted_at, replaced_by_id FROM penalty_types.*`).
		WithArgs(penaltyTypeID).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	_, err = repo.GetByID(ctx, penaltyTypeID)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_GetByClubID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	clubID := uuid.NewString()
	penaltyTypeID1 := uuid.NewString()
	penaltyTypeID2 := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT id, club_id, name, description, price, display_order, allows_decimal_quantity, created_at, updated_at, deleted_at, replaced_by_id FROM penalty_types`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "description", "price", "display_order", "allows_decimal_quantity", "created_at", "updated_at", "deleted_at", "replaced_by_id"}).
			AddRow(penaltyTypeID1, clubID, "Penalty 1", "Description 1", 500, 1, false, now, now, nil, nil).
			AddRow(penaltyTypeID2, clubID, "Penalty 2", "Description 2", 1000, 2, false, now, now, nil, nil))

	ctx := context.Background()
	penaltyTypes, err := repo.GetByClubID(ctx, clubID)

	if err != nil {
		t.Fatalf("get penalty types: %v", err)
	}

	if len(penaltyTypes) != 2 {
		t.Fatalf("expected 2 penalty types, got %d", len(penaltyTypes))
	}
	if penaltyTypes[0].ID != penaltyTypeID1 {
		t.Fatalf("expected first id %s, got %s", penaltyTypeID1, penaltyTypes[0].ID)
	}
	if penaltyTypes[1].ID != penaltyTypeID2 {
		t.Fatalf("expected second id %s, got %s", penaltyTypeID2, penaltyTypes[1].ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_GetByClubID_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	clubID := uuid.NewString()

	mock.ExpectQuery(`SELECT id, club_id, name, description, price, display_order, allows_decimal_quantity, created_at, updated_at, deleted_at, replaced_by_id FROM penalty_types`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "description", "price", "display_order", "allows_decimal_quantity", "created_at", "updated_at", "deleted_at", "replaced_by_id"}))

	ctx := context.Background()
	penaltyTypes, err := repo.GetByClubID(ctx, clubID)

	if err != nil {
		t.Fatalf("get penalty types: %v", err)
	}

	if len(penaltyTypes) != 0 {
		t.Fatalf("expected 0 penalty types, got %d", len(penaltyTypes))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_Replace(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	oldID := uuid.NewString()
	newID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	// First call: GetByID to get old penalty type
	mock.ExpectQuery(`.*SELECT id, club_id, name, description, price, display_order, allows_decimal_quantity, created_at, updated_at, deleted_at, replaced_by_id FROM penalty_types.*`).
		WithArgs(oldID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "description", "price", "display_order", "allows_decimal_quantity", "created_at", "updated_at", "deleted_at", "replaced_by_id"}).
			AddRow(oldID, clubID, "Old Penalty", "Old Description", 500, 1, false, now, now, nil, nil))

	// Second call: MarkPenaltyTypeAsReplaced
	mock.ExpectExec(`UPDATE penalty_types`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), oldID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Third call: CreatePenaltyType for new version
	mock.ExpectQuery(`INSERT INTO penalty_types`).
		WithArgs(sqlmock.AnyArg(), clubID, "New Penalty", sqlmock.AnyArg(), 1000, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "description", "price", "display_order", "allows_decimal_quantity", "created_at", "updated_at", "deleted_at", "replaced_by_id"}).
			AddRow(newID, clubID, "New Penalty", "New Description", 1000, 1, false, now, now, nil, nil))

	ctx := context.Background()
	newPenaltyType, err := repo.Replace(ctx, ReplacePenaltyTypeParams{
		ID:          oldID,
		Name:        "New Penalty",
		Description: "New Description",
		Price:       1000,
	})

	if err != nil {
		t.Fatalf("replace penalty type: %v", err)
	}

	if newPenaltyType.ID != newID {
		t.Fatalf("expected new id %s, got %s", newID, newPenaltyType.ID)
	}
	if newPenaltyType.Name != "New Penalty" {
		t.Fatalf("expected name 'New Penalty', got %s", newPenaltyType.Name)
	}
	if newPenaltyType.Price != 1000 {
		t.Fatalf("expected price 1000, got %d", newPenaltyType.Price)
	}
	if newPenaltyType.DisplayOrder != 1 {
		t.Fatalf("expected display_order 1 (kept from old), got %d", newPenaltyType.DisplayOrder)
	}
	if newPenaltyType.ClubID != clubID {
		t.Fatalf("expected club_id %s, got %s", clubID, newPenaltyType.ClubID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_Replace_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	oldID := uuid.NewString()

	// GetByID returns not found
	mock.ExpectQuery(`SELECT id, club_id, name, description, price, display_order, allows_decimal_quantity, created_at, updated_at, deleted_at, replaced_by_id FROM penalty_types`).
		WithArgs(oldID).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	_, err = repo.Replace(ctx, ReplacePenaltyTypeParams{
		ID:          oldID,
		Name:        "New Penalty",
		Description: "New Description",
		Price:       1000,
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	penaltyTypeID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	// First call: GetByID to verify exists
	mock.ExpectQuery(`.*SELECT id, club_id, name, description, price, display_order, allows_decimal_quantity, created_at, updated_at, deleted_at, replaced_by_id FROM penalty_types.*`).
		WithArgs(penaltyTypeID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "description", "price", "display_order", "allows_decimal_quantity", "created_at", "updated_at", "deleted_at", "replaced_by_id"}).
			AddRow(penaltyTypeID, clubID, "Test Penalty", "Test Description", 500, 1, false, now, now, nil, nil))

	// Second call: DeletePenaltyType (soft delete)
	mock.ExpectExec(`UPDATE penalty_types`).
		WithArgs(sqlmock.AnyArg(), penaltyTypeID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err = repo.Delete(ctx, penaltyTypeID)

	if err != nil {
		t.Fatalf("delete penalty type: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	penaltyTypeID := uuid.NewString()

	// GetByID returns not found
	mock.ExpectQuery(`SELECT id, club_id, name, description, price, display_order, allows_decimal_quantity, created_at, updated_at, deleted_at, replaced_by_id FROM penalty_types`).
		WithArgs(penaltyTypeID).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	err = repo.Delete(ctx, penaltyTypeID)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

