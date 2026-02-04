package player

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

	playerID := uuid.NewString()
	clubID := uuid.NewString()
	userID := uuid.NewString()
	roleID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`INSERT INTO players`).
		WithArgs(sqlmock.AnyArg(), clubID, &userID, &roleID, "Test Player", 1000, 500, sqlmock.AnyArg(), false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "created_at", "updated_at"}).
			AddRow(playerID, clubID, userID, roleID, "Test Player", 1000, 500, nil, false, now, now))

	ctx := context.Background()
	player, err := repo.Create(ctx, CreatePlayerParams{
		ClubID:       clubID,
		UserID:       &userID,
		RoleID:       &roleID,
		Name:         "Test Player",
		Balance:      1000,
		StartBalance: 500,
		Inactive:     false,
	})

	if err != nil {
		t.Fatalf("create player: %v", err)
	}

	if player.ID != playerID {
		t.Fatalf("expected id %s, got %s", playerID, player.ID)
	}
	if player.Name != "Test Player" {
		t.Fatalf("expected name 'Test Player', got %s", player.Name)
	}
	if player.Balance != 1000 {
		t.Fatalf("expected balance 1000, got %d", player.Balance)
	}
	if player.StartBalance != 500 {
		t.Fatalf("expected start_balance 500, got %d", player.StartBalance)
	}
	if player.ClubID != clubID {
		t.Fatalf("expected club_id %s, got %s", clubID, player.ClubID)
	}
	if player.UserID == nil || *player.UserID != userID {
		t.Fatalf("expected user_id %s, got %v", userID, player.UserID)
	}
	if player.RoleID == nil || *player.RoleID != roleID {
		t.Fatalf("expected role_id %s, got %v", roleID, player.RoleID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRepository_Create_NoUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	playerID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`INSERT INTO players`).
		WithArgs(sqlmock.AnyArg(), clubID, nil, nil, "Test Player", 1000, 500, sqlmock.AnyArg(), false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "created_at", "updated_at"}).
			AddRow(playerID, clubID, nil, nil, "Test Player", 1000, 500, nil, false, now, now))

	ctx := context.Background()
	player, err := repo.Create(ctx, CreatePlayerParams{
		ClubID:       clubID,
		UserID:       nil,
		RoleID:       nil,
		Name:         "Test Player",
		Balance:      1000,
		StartBalance: 500,
		Inactive:     false,
	})

	if err != nil {
		t.Fatalf("create player: %v", err)
	}

	if player.UserID != nil {
		t.Fatalf("expected nil user_id, got %v", player.UserID)
	}
	if player.RoleID != nil {
		t.Fatalf("expected nil role_id, got %v", player.RoleID)
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

	playerID := uuid.NewString()
	clubID := uuid.NewString()
	userID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, created_at, updated_at`).
		WithArgs(playerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "created_at", "updated_at"}).
			AddRow(playerID, clubID, userID, nil, "Test Player", 1000, 500, nil, false, now, now))

	ctx := context.Background()
	player, err := repo.GetByID(ctx, playerID)

	if err != nil {
		t.Fatalf("get player: %v", err)
	}

	if player.ID != playerID {
		t.Fatalf("expected id %s, got %s", playerID, player.ID)
	}
	if player.Name != "Test Player" {
		t.Fatalf("expected name 'Test Player', got %s", player.Name)
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

	playerID := uuid.NewString()

	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, created_at, updated_at`).
		WithArgs(playerID).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	_, err = repo.GetByID(ctx, playerID)

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

func TestRepository_GetByClubID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	playerID1 := uuid.NewString()
	playerID2 := uuid.NewString()
	clubID := uuid.NewString()
	userID := uuid.NewString()
	roleID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "created_at", "updated_at"}).
			AddRow(playerID1, clubID, userID, roleID, "Player 1", 1000, 500, nil, false, now, now).
			AddRow(playerID2, clubID, nil, nil, "Player 2", 2000, 600, nil, false, now.Add(time.Hour), now.Add(time.Hour)))

	ctx := context.Background()
	players, err := repo.GetByClubID(ctx, clubID)

	if err != nil {
		t.Fatalf("get players by club id: %v", err)
	}

	if len(players) != 2 {
		t.Fatalf("expected 2 players, got %d", len(players))
	}
	if players[0].ID != playerID1 {
		t.Fatalf("expected first player id %s, got %s", playerID1, players[0].ID)
	}
	if players[1].ID != playerID2 {
		t.Fatalf("expected second player id %s, got %s", playerID2, players[1].ID)
	}
	if players[0].UserID == nil || *players[0].UserID != userID {
		t.Fatalf("expected user_id %s for first player, got %v", userID, players[0].UserID)
	}
	if players[0].RoleID == nil || *players[0].RoleID != roleID {
		t.Fatalf("expected role_id %s for first player, got %v", roleID, players[0].RoleID)
	}
	if players[1].UserID != nil {
		t.Fatalf("expected nil user_id for second player, got %v", players[1].UserID)
	}
	if players[1].RoleID != nil {
		t.Fatalf("expected nil role_id for second player, got %v", players[1].RoleID)
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

	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "created_at", "updated_at"}))

	ctx := context.Background()
	players, err := repo.GetByClubID(ctx, clubID)

	if err != nil {
		t.Fatalf("get players by club id: %v", err)
	}

	if len(players) != 0 {
		t.Fatalf("expected 0 players, got %d", len(players))
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

	playerID := uuid.NewString()
	clubID := uuid.NewString()
	userID := uuid.NewString()
	roleID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`UPDATE players`).
		WithArgs("Updated Player", 2000, 600, &userID, &roleID, sqlmock.AnyArg(), false, sqlmock.AnyArg(), playerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "created_at", "updated_at"}).
			AddRow(playerID, clubID, userID, roleID, "Updated Player", 2000, 600, nil, false, now, now.Add(time.Hour)))

	inactive := false
	ctx := context.Background()
	player, err := repo.Update(ctx, UpdatePlayerParams{
		ID:           playerID,
		Name:         "Updated Player",
		Balance:      2000,
		StartBalance: 600,
		UserID:       &userID,
		RoleID:       &roleID,
		Inactive:     &inactive,
	})

	if err != nil {
		t.Fatalf("update player: %v", err)
	}

	if player.ID != playerID {
		t.Fatalf("expected id %s, got %s", playerID, player.ID)
	}
	if player.Name != "Updated Player" {
		t.Fatalf("expected name 'Updated Player', got %s", player.Name)
	}
	if player.Balance != 2000 {
		t.Fatalf("expected balance 2000, got %d", player.Balance)
	}
	if player.StartBalance != 600 {
		t.Fatalf("expected start_balance 600, got %d", player.StartBalance)
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

	playerID := uuid.NewString()

	// When Inactive is nil, Update loads existing player first; GetPlayerByID returns NotFound
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, created_at, updated_at FROM players`).
		WithArgs(playerID).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	_, err = repo.Update(ctx, UpdatePlayerParams{
		ID:           playerID,
		Name:         "Updated Player",
		Balance:      2000,
		StartBalance: 600,
		UserID:       nil,
		RoleID:       nil,
		Inactive:     nil, // nil triggers load-existing; player not found
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

func TestRepository_Update_PreservesInactiveWhenNil(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	playerID := uuid.NewString()
	clubID := uuid.NewString()
	userID := uuid.NewString()
	roleID := uuid.NewString()
	now := time.Now().UTC()

	// When Inactive is nil, Update loads existing player first (existing has Inactive=true)
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, created_at, updated_at FROM players`).
		WithArgs(playerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "created_at", "updated_at"}).
			AddRow(playerID, clubID, userID, roleID, "Old Name", 1000, 500, nil, true, now, now))

	// UPDATE is called with inactive=true (preserved from existing)
	mock.ExpectQuery(`UPDATE players`).
		WithArgs("Updated Player", 2000, 600, &userID, &roleID, sqlmock.AnyArg(), true, sqlmock.AnyArg(), playerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "created_at", "updated_at"}).
			AddRow(playerID, clubID, userID, roleID, "Updated Player", 2000, 600, nil, true, now, now.Add(time.Hour)))

	ctx := context.Background()
	player, err := repo.Update(ctx, UpdatePlayerParams{
		ID:           playerID,
		Name:         "Updated Player",
		Balance:      2000,
		StartBalance: 600,
		UserID:       &userID,
		RoleID:       &roleID,
		Inactive:     nil, // do not change; existing inactive=true must be preserved
	})
	if err != nil {
		t.Fatalf("update player: %v", err)
	}
	if !player.Inactive {
		t.Fatalf("expected Inactive=true (preserved), got false")
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

	playerID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	// First call: GetPlayerByID to check if player exists
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, created_at, updated_at FROM players`).
		WithArgs(playerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "created_at", "updated_at"}).
			AddRow(playerID, clubID, nil, nil, "Test Player", 1000, 500, nil, false, now, now))

	// Second call: DeletePlayer
	mock.ExpectExec(`DELETE FROM players`).
		WithArgs(playerID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx := context.Background()
	err = repo.Delete(ctx, playerID)

	if err != nil {
		t.Fatalf("delete player: %v", err)
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

	playerID := uuid.NewString()

	// GetPlayerByID returns ErrNoRows
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, created_at, updated_at FROM players`).
		WithArgs(playerID).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	err = repo.Delete(ctx, playerID)

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

	clubID := uuid.NewString()

	mock.ExpectQuery(`INSERT INTO players`).
		WithArgs(sqlmock.AnyArg(), clubID, nil, nil, "Test Player", 1000, 500, sqlmock.AnyArg(), false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	_, err = repo.Create(ctx, CreatePlayerParams{
		ClubID:       clubID,
		UserID:       nil,
		RoleID:       nil,
		Name:         "Test Player",
		Balance:      1000,
		StartBalance: 500,
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

func TestRepository_GetByClubID_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)

	clubID := uuid.NewString()

	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "created_at", "updated_at"}).
			AddRow(uuid.NewString(), clubID, nil, nil, "Player 1", 1000, 500, nil, false, time.Now(), time.Now()).
			RowError(0, sql.ErrConnDone))

	ctx := context.Background()
	_, err = repo.GetByClubID(ctx, clubID)

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

	playerID := uuid.NewString()

	// GetPlayerByID returns a database error
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, created_at, updated_at FROM players`).
		WithArgs(playerID).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	err = repo.Delete(ctx, playerID)

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

	playerID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	// First call: GetPlayerByID succeeds
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, created_at, updated_at FROM players`).
		WithArgs(playerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "created_at", "updated_at"}).
			AddRow(playerID, clubID, nil, nil, "Test Player", 1000, 500, nil, false, now, now))

	// Second call: DeletePlayer returns an error
	mock.ExpectExec(`DELETE FROM players`).
		WithArgs(playerID).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	err = repo.Delete(ctx, playerID)

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

