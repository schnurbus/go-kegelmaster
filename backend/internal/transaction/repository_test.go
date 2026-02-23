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

// TestCreateDepositCouplesMode_AutoTipFromPoolAllocation verifies pool allocation: total amount
// is applied to debts first; remainder is auto-tip. Example: A -10,40, B -10,00, pay 20,50
// → A +10,40, B +10,00, Auto-Tip +0,10.
func TestCreateDepositCouplesMode_AutoTipFromPoolAllocation(t *testing.T) {
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
	player1ID := uuid.NewString() // A: debt 1040
	player2ID := uuid.NewString()  // B: debt 1000
	now := time.Now().UTC()
	txDate := time.Date(2025, 2, 20, 12, 0, 0, 0, time.UTC)

	// Club
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Club", 0, 0, 500, true, true, uuid.NewString(), now, now))
	// Player 1: balance -1040 (10,40 € debt)
	mock.ExpectQuery(`SELECT .* FROM players`).
		WithArgs(player1ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(player1ID, clubID, nil, nil, "Player A", -1040, 0, nil, false, nil, now, now))
	// Player 2: balance -1000 (10,00 € debt)
	mock.ExpectQuery(`SELECT .* FROM players`).
		WithArgs(player2ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(player2ID, clubID, nil, nil, "Player B", -1000, 0, nil, false, nil, now, now))

	txCols := []string{"id", "club_id", "player_id", "transaction_type", "amount", "description", "game_day_fee_id", "game_day_id", "player_balance_before", "player_balance_after", "club_balance_before", "club_balance_after", "transaction_date", "created_at", "updated_at"}
	addTxRow := func(amount int32, txType string, playerID string, balanceBefore, balanceAfter int) *sqlmock.Rows {
		return sqlmock.NewRows(txCols).
			AddRow(uuid.NewString(), clubID, playerID, txType, amount, nil, nil, nil, balanceBefore, balanceAfter, 0, 2050, txDate, now, now)
	}

	mock.ExpectBegin()
	// Deposit 1040 for player 1
	mock.ExpectQuery(`INSERT INTO transactions`).
		WithArgs(sqlmock.AnyArg(), clubID, player1ID, "deposit", 1040, sqlmock.AnyArg(), nil, nil, -1040, 0, 0, 2050, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(addTxRow(1040, "deposit", player1ID, -1040, 0))
	mock.ExpectExec(`UPDATE players`).
		WithArgs(0, sqlmock.AnyArg(), player1ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Deposit 1000 for player 2
	mock.ExpectQuery(`INSERT INTO transactions`).
		WithArgs(sqlmock.AnyArg(), clubID, player2ID, "deposit", 1000, sqlmock.AnyArg(), nil, nil, -1000, 0, 0, 2050, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(addTxRow(1000, "deposit", player2ID, -1000, 0))
	mock.ExpectExec(`UPDATE players`).
		WithArgs(0, sqlmock.AnyArg(), player2ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Auto-tip 10
	mock.ExpectQuery(`INSERT INTO transactions`).
		WithArgs(sqlmock.AnyArg(), clubID, player1ID, "tip", 10, sqlmock.AnyArg(), nil, nil, 0, 0, 0, 2050, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(addTxRow(10, "tip", player1ID, 0, 0))
	mock.ExpectQuery(`UPDATE clubs`).
		WithArgs("Club", 2050, 0, 500, true, true, sqlmock.AnyArg(), clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Club", 2050, 0, 500, true, true, uuid.NewString(), now, now))
	mock.ExpectCommit()

	results, err := repo.CreateDepositCouplesMode(ctx, clubID, []string{player1ID, player2ID}, 2050, "Einzahlung", txDate, true)
	if err != nil {
		t.Fatalf("CreateDepositCouplesMode: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}

	// Expect 3 transactions: deposit 1040, deposit 1000, tip 10
	if len(results) != 3 {
		t.Fatalf("expected 3 transactions, got %d", len(results))
	}
	amounts := make([]int, len(results))
	types := make([]string, len(results))
	for i, tx := range results {
		amounts[i] = tx.Amount
		types[i] = string(tx.TransactionType)
	}
	if amounts[0] != 1040 || types[0] != "deposit" {
		t.Errorf("first tx: expected deposit 1040, got %s %d", types[0], amounts[0])
	}
	if amounts[1] != 1000 || types[1] != "deposit" {
		t.Errorf("second tx: expected deposit 1000, got %s %d", types[1], amounts[1])
	}
	if amounts[2] != 10 || types[2] != "tip" {
		t.Errorf("third tx: expected tip 10, got %s %d", types[2], amounts[2])
	}
}
