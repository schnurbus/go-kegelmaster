package club

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

	clubID := uuid.NewString()
	userID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`INSERT INTO clubs`).
		WithArgs(sqlmock.AnyArg(), "Test Club", 1000, 500, true, userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "base_fee", "auto_tip_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 500, true, userID, now, now))

	ctx := context.Background()
	club, err := repo.Create(ctx, CreateClubParams{
		Name:           "Test Club",
		Balance:        1000,
		BaseFee:        500,
		AutoTipEnabled: true,
		UserID:         userID,
	})

	if err != nil {
		t.Fatalf("create club: %v", err)
	}

	if club.ID != clubID {
		t.Fatalf("expected id %s, got %s", clubID, club.ID)
	}
	if club.Name != "Test Club" {
		t.Fatalf("expected name 'Test Club', got %s", club.Name)
	}
	if club.Balance != 1000 {
		t.Fatalf("expected balance 1000, got %d", club.Balance)
	}
	if club.BaseFee != 500 {
		t.Fatalf("expected base_fee 500, got %d", club.BaseFee)
	}
	if club.UserID != userID {
		t.Fatalf("expected user_id %s, got %s", userID, club.UserID)
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

	clubID := uuid.NewString()
	userID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT id, name, balance, base_fee, auto_tip_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "base_fee", "auto_tip_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 500, true, userID, now, now))

	ctx := context.Background()
	club, err := repo.GetByID(ctx, clubID)

	if err != nil {
		t.Fatalf("get club: %v", err)
	}

	if club.ID != clubID {
		t.Fatalf("expected id %s, got %s", clubID, club.ID)
	}
	if club.Name != "Test Club" {
		t.Fatalf("expected name 'Test Club', got %s", club.Name)
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

	clubID := uuid.NewString()

	mock.ExpectQuery(`SELECT id, name, balance, base_fee, auto_tip_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	_, err = repo.GetByID(ctx, clubID)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	clubID1 := uuid.NewString()
	clubID2 := uuid.NewString()
	userID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT id, name, balance, base_fee, auto_tip_enabled, user_id, created_at, updated_at`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "base_fee", "auto_tip_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID1, "Club 1", 1000, 500, true, userID, now, now).
			AddRow(clubID2, "Club 2", 2000, 600, true, userID, now.Add(time.Hour), now.Add(time.Hour)))

	ctx := context.Background()
	clubs, err := repo.GetAll(ctx)

	if err != nil {
		t.Fatalf("get all clubs: %v", err)
	}

	if len(clubs) != 2 {
		t.Fatalf("expected 2 clubs, got %d", len(clubs))
	}
	if clubs[0].ID != clubID1 {
		t.Fatalf("expected first club id %s, got %s", clubID1, clubs[0].ID)
	}
	if clubs[1].ID != clubID2 {
		t.Fatalf("expected second club id %s, got %s", clubID2, clubs[1].ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_GetAll_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	mock.ExpectQuery(`SELECT id, name, balance, base_fee, auto_tip_enabled, user_id, created_at, updated_at`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "base_fee", "auto_tip_enabled", "user_id", "created_at", "updated_at"}))

	ctx := context.Background()
	clubs, err := repo.GetAll(ctx)

	if err != nil {
		t.Fatalf("get all clubs: %v", err)
	}

	if len(clubs) != 0 {
		t.Fatalf("expected 0 clubs, got %d", len(clubs))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_GetByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	clubID1 := uuid.NewString()
	clubID2 := uuid.NewString()
	userID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT id, name, balance, base_fee, auto_tip_enabled, user_id, created_at, updated_at`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "base_fee", "auto_tip_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID1, "Club 1", 1000, 500, true, userID, now, now).
			AddRow(clubID2, "Club 2", 2000, 600, true, userID, now.Add(time.Hour), now.Add(time.Hour)))

	ctx := context.Background()
	clubs, err := repo.GetByUserID(ctx, userID)

	if err != nil {
		t.Fatalf("get clubs by user id: %v", err)
	}

	if len(clubs) != 2 {
		t.Fatalf("expected 2 clubs, got %d", len(clubs))
	}
	if clubs[0].UserID != userID {
		t.Fatalf("expected user_id %s, got %s", userID, clubs[0].UserID)
	}
	if clubs[1].UserID != userID {
		t.Fatalf("expected user_id %s, got %s", userID, clubs[1].UserID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_GetByUserID_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	userID := uuid.NewString()

	mock.ExpectQuery(`SELECT id, name, balance, base_fee, auto_tip_enabled, user_id, created_at, updated_at`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "base_fee", "auto_tip_enabled", "user_id", "created_at", "updated_at"}))

	ctx := context.Background()
	clubs, err := repo.GetByUserID(ctx, userID)

	if err != nil {
		t.Fatalf("get clubs by user id: %v", err)
	}

	if len(clubs) != 0 {
		t.Fatalf("expected 0 clubs, got %d", len(clubs))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_GetForUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	userID := uuid.NewString()
	clubID1 := uuid.NewString()
	otherUserID := uuid.NewString()
	clubID2 := uuid.NewString()
	now := time.Now().UTC()

	// User is owner of club 1; user has linked player in club 2 (owned by other user)
	mock.ExpectQuery(`SELECT id, name, balance, base_fee, auto_tip_enabled, user_id, created_at, updated_at FROM clubs c`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "base_fee", "auto_tip_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID1, "Club A", 1000, 500, true, userID, now, now).
			AddRow(clubID2, "Club B", 2000, 600, true, otherUserID, now.Add(time.Hour), now.Add(time.Hour)))

	ctx := context.Background()
	clubs, err := repo.GetForUser(ctx, userID)

	if err != nil {
		t.Fatalf("get clubs for user: %v", err)
	}

	if len(clubs) != 2 {
		t.Fatalf("expected 2 clubs, got %d", len(clubs))
	}
	if clubs[0].ID != clubID1 {
		t.Fatalf("expected first club id %s, got %s", clubID1, clubs[0].ID)
	}
	if clubs[0].UserID != userID {
		t.Fatalf("expected first club owner %s, got %s", userID, clubs[0].UserID)
	}
	if clubs[1].ID != clubID2 {
		t.Fatalf("expected second club id %s, got %s", clubID2, clubs[1].ID)
	}
	if clubs[1].UserID != otherUserID {
		t.Fatalf("expected second club owner %s, got %s", otherUserID, clubs[1].UserID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_GetForUser_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	userID := uuid.NewString()

	mock.ExpectQuery(`SELECT id, name, balance, base_fee, auto_tip_enabled, user_id, created_at, updated_at FROM clubs c`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "base_fee", "auto_tip_enabled", "user_id", "created_at", "updated_at"}))

	ctx := context.Background()
	clubs, err := repo.GetForUser(ctx, userID)

	if err != nil {
		t.Fatalf("get clubs for user: %v", err)
	}

	if len(clubs) != 0 {
		t.Fatalf("expected 0 clubs, got %d", len(clubs))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	clubID := uuid.NewString()
	userID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`UPDATE clubs`).
		WithArgs("Updated Club", 2000, 600, false, sqlmock.AnyArg(), clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "base_fee", "auto_tip_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Updated Club", 2000, 600, false, userID, now, now.Add(time.Hour)))

	ctx := context.Background()
	club, err := repo.Update(ctx, UpdateClubParams{
		ID:             clubID,
		Name:           "Updated Club",
		Balance:        2000,
		BaseFee:        600,
		AutoTipEnabled: false,
	})

	if err != nil {
		t.Fatalf("update club: %v", err)
	}

	if club.ID != clubID {
		t.Fatalf("expected id %s, got %s", clubID, club.ID)
	}
	if club.Name != "Updated Club" {
		t.Fatalf("expected name 'Updated Club', got %s", club.Name)
	}
	if club.Balance != 2000 {
		t.Fatalf("expected balance 2000, got %d", club.Balance)
	}
	if club.BaseFee != 600 {
		t.Fatalf("expected base_fee 600, got %d", club.BaseFee)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	clubID := uuid.NewString()

	mock.ExpectQuery(`UPDATE clubs`).
		WithArgs("Updated Club", 2000, 600, true, sqlmock.AnyArg(), clubID).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	_, err = repo.Update(ctx, UpdateClubParams{
		ID:             clubID,
		Name:           "Updated Club",
		Balance:        2000,
		BaseFee:        600,
		AutoTipEnabled: true,
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
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

	clubID := uuid.NewString()
	userID := uuid.NewString()
	now := time.Now().UTC()

	// First call: GetClubByID to check if club exists
	mock.ExpectQuery(`SELECT id, name, balance, base_fee, auto_tip_enabled, user_id, created_at, updated_at FROM clubs`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "base_fee", "auto_tip_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 500, true, userID, now, now))

	// Second call: DeleteClub
	mock.ExpectExec(`DELETE FROM clubs`).
		WithArgs(clubID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx := context.Background()
	err = repo.Delete(ctx, clubID)

	if err != nil {
		t.Fatalf("delete club: %v", err)
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

	clubID := uuid.NewString()

	// GetClubByID returns ErrNoRows
	mock.ExpectQuery(`SELECT id, name, balance, base_fee, auto_tip_enabled, user_id, created_at, updated_at FROM clubs`).
		WithArgs(clubID).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	err = repo.Delete(ctx, clubID)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_Create_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	userID := uuid.NewString()

	mock.ExpectQuery(`INSERT INTO clubs`).
		WithArgs(sqlmock.AnyArg(), "Test Club", 1000, 500, true, userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	_, err = repo.Create(ctx, CreateClubParams{
		Name:           "Test Club",
		Balance:        1000,
		BaseFee:        500,
		AutoTipEnabled: true,
		UserID:         userID,
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != sql.ErrConnDone {
		t.Fatalf("expected sql.ErrConnDone, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_GetAll_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	mock.ExpectQuery(`SELECT id, name, balance, base_fee, auto_tip_enabled, user_id, created_at, updated_at`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "base_fee", "auto_tip_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(uuid.NewString(), "Club 1", 1000, 500, true, uuid.NewString(), time.Now(), time.Now()).
			RowError(0, sql.ErrConnDone))

	ctx := context.Background()
	_, err = repo.GetAll(ctx)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_GetByUserID_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	userID := uuid.NewString()

	mock.ExpectQuery(`SELECT id, name, balance, base_fee, auto_tip_enabled, user_id, created_at, updated_at`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "base_fee", "auto_tip_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(uuid.NewString(), "Club 1", 1000, 500, true, userID, time.Now(), time.Now()).
			RowError(0, sql.ErrConnDone))

	ctx := context.Background()
	_, err = repo.GetByUserID(ctx, userID)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_Delete_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	clubID := uuid.NewString()

	// GetClubByID returns a database error
	mock.ExpectQuery(`SELECT id, name, balance, base_fee, auto_tip_enabled, user_id, created_at, updated_at FROM clubs`).
		WithArgs(clubID).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	err = repo.Delete(ctx, clubID)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != sql.ErrConnDone {
		t.Fatalf("expected sql.ErrConnDone, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_Delete_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	clubID := uuid.NewString()
	userID := uuid.NewString()
	now := time.Now().UTC()

	// First call: GetClubByID succeeds
	mock.ExpectQuery(`SELECT id, name, balance, base_fee, auto_tip_enabled, user_id, created_at, updated_at FROM clubs`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "base_fee", "auto_tip_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 500, true, userID, now, now))

	// Second call: DeleteClub returns an error
	mock.ExpectExec(`DELETE FROM clubs`).
		WithArgs(clubID).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	err = repo.Delete(ctx, clubID)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != sql.ErrConnDone {
		t.Fatalf("expected sql.ErrConnDone, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
