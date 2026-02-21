package transaction

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
)

func TestCreateDepositCouplesMode_RequiresAtLeastTwoPlayers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	clubRepo := club.NewRepository(db)
	playerRepo := player.NewRepository(db)
	repo := NewRepository(db, playerRepo, clubRepo, nil)

	ctx := context.Background()
	clubID := uuid.NewString()
	playerID := uuid.NewString()

	_, err = repo.CreateDepositCouplesMode(ctx, clubID, []string{playerID}, 1000, "Einzahlung", time.Now(), true)
	if err == nil {
		t.Fatal("expected error for single player, got nil")
	}
	if err.Error() != "couples mode requires at least 2 players, got 1" {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateDepositCouplesMode_InvalidAmount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	clubRepo := club.NewRepository(db)
	playerRepo := player.NewRepository(db)
	repo := NewRepository(db, playerRepo, clubRepo, nil)

	ctx := context.Background()
	clubID := uuid.NewString()
	player1ID := uuid.NewString()
	player2ID := uuid.NewString()

	// totalAmount <= 0 is checked before any DB access
	_, err = repo.CreateDepositCouplesMode(ctx, clubID, []string{player1ID, player2ID}, 0, "Einzahlung", time.Now(), true)
	if err == nil {
		t.Fatal("expected error for zero amount, got nil")
	}
	if err != ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateDepositCouplesMode_PlayerNotInClub(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	clubRepo := club.NewRepository(db)
	playerRepo := player.NewRepository(db)
	repo := NewRepository(db, playerRepo, clubRepo, nil)

	ctx := context.Background()
	clubID := uuid.NewString()
	otherClubID := uuid.NewString()
	player1ID := uuid.NewString()
	player2ID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Club", 0, 0, 500, true, true, uuid.NewString(), now, now))
	mock.ExpectQuery(`SELECT .* FROM players`).
		WithArgs(player1ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(player1ID, clubID, nil, nil, "Player 1", -500, 0, nil, false, nil, now, now))
	mock.ExpectQuery(`SELECT .* FROM players`).
		WithArgs(player2ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(player2ID, otherClubID, nil, nil, "Player 2", -500, 0, nil, false, nil, now, now))

	_, err = repo.CreateDepositCouplesMode(ctx, clubID, []string{player1ID, player2ID}, 1000, "Einzahlung", time.Now(), true)
	if err == nil {
		t.Fatal("expected error for player not in club, got nil")
	}
	if err.Error() != "player "+player2ID+" does not belong to club" {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateDepositCouplesMode_ClubNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	clubRepo := club.NewRepository(db)
	playerRepo := player.NewRepository(db)
	repo := NewRepository(db, playerRepo, clubRepo, nil)

	ctx := context.Background()
	clubID := uuid.NewString()
	player1ID := uuid.NewString()
	player2ID := uuid.NewString()

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.CreateDepositCouplesMode(ctx, clubID, []string{player1ID, player2ID}, 1000, "Einzahlung", time.Now(), true)
	if err == nil {
		t.Fatal("expected error when club not found, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
