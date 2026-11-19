package transaction

import (
	"context"
	"database/sql"
	"errors"
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

// Fixed IDs keep the lock order (sorted by ID) deterministic.
const (
	testPlayerA = "00000000-0000-0000-0000-00000000000a"
	testPlayerB = "00000000-0000-0000-0000-00000000000b"
)

func newTestRepo(t *testing.T) (*Repository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	repo := NewRepository(db, player.NewRepository(db), club.NewRepository(db), nil)
	return repo, mock, func() { db.Close() }
}

func TestCreateDepositCouplesMode_PlayerNotInClub(t *testing.T) {
	repo, mock, done := newTestRepo(t)
	defer done()

	clubID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectBegin()
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(testPlayerA).
		WillReturnRows(playerRow(testPlayerA, clubID, -500, now))
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(testPlayerB).
		WillReturnRows(playerRow(testPlayerB, uuid.NewString(), -500, now))
	mock.ExpectRollback()

	_, err := repo.CreateDepositCouplesMode(context.Background(), clubID, []string{testPlayerA, testPlayerB}, 1000, "Einzahlung", now, true)
	if !errors.Is(err, ErrPlayerNotInClub) {
		t.Fatalf("expected ErrPlayerNotInClub, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateDepositCouplesMode_DuplicatePlayer(t *testing.T) {
	repo, mock, done := newTestRepo(t)
	defer done()

	_, err := repo.CreateDepositCouplesMode(context.Background(), uuid.NewString(), []string{testPlayerA, testPlayerA}, 1000, "Einzahlung", time.Now(), true)
	if !errors.Is(err, ErrDuplicatePlayer) {
		t.Fatalf("expected ErrDuplicatePlayer, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreateDepositCouplesMode_ClubNotFound(t *testing.T) {
	repo, mock, done := newTestRepo(t)
	defer done()

	clubID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectBegin()
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(testPlayerA).
		WillReturnRows(playerRow(testPlayerA, clubID, -500, now))
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(testPlayerB).
		WillReturnRows(playerRow(testPlayerB, clubID, -500, now))
	mock.ExpectQuery("IncrementClubBalance").
		WithArgs(1000, sqlmock.AnyArg(), clubID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err := repo.CreateDepositCouplesMode(context.Background(), clubID, []string{testPlayerA, testPlayerB}, 1000, "Einzahlung", now, true)
	if !errors.Is(err, club.ErrNotFound) {
		t.Fatalf("expected club.ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestCreateDepositCouplesMode_AutoTipFromPoolAllocation verifies pool allocation: total amount
// is applied to debts first; remainder is auto-tip. Example: A -10,40, B -10,00, pay 20,50
// → A +10,40, B +10,00, Auto-Tip +0,10. Players are passed as [B, A] to check that the
// allocation follows the request order while locks follow the ID order.
func TestCreateDepositCouplesMode_AutoTipFromPoolAllocation(t *testing.T) {
	repo, mock, done := newTestRepo(t)
	defer done()

	clubID := uuid.NewString()
	now := time.Now().UTC()
	txDate := time.Date(2025, 2, 20, 12, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(testPlayerA).
		WillReturnRows(playerRow(testPlayerA, clubID, -1000, now))
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(testPlayerB).
		WillReturnRows(playerRow(testPlayerB, clubID, -1040, now))
	mock.ExpectQuery("IncrementClubBalance").
		WithArgs(2050, sqlmock.AnyArg(), clubID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(2050))
	// Deposit 1040 for B (first in request order)
	mock.ExpectQuery("INSERT INTO transactions").
		WithArgs(sqlmock.AnyArg(), clubID, testPlayerB, "deposit", 1040, sqlmock.AnyArg(), nil, nil, -1040, 0, 0, 2050, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(couplesTxRow(clubID, testPlayerB, "deposit", 1040, -1040, 0, txDate, now))
	mock.ExpectExec("UpdatePlayerBalance").
		WithArgs(0, sqlmock.AnyArg(), testPlayerB).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Deposit 1000 for A
	mock.ExpectQuery("INSERT INTO transactions").
		WithArgs(sqlmock.AnyArg(), clubID, testPlayerA, "deposit", 1000, sqlmock.AnyArg(), nil, nil, -1000, 0, 0, 2050, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(couplesTxRow(clubID, testPlayerA, "deposit", 1000, -1000, 0, txDate, now))
	mock.ExpectExec("UpdatePlayerBalance").
		WithArgs(0, sqlmock.AnyArg(), testPlayerA).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Auto-tip 10 on the first player; balance unchanged, so no update
	mock.ExpectQuery("INSERT INTO transactions").
		WithArgs(sqlmock.AnyArg(), clubID, testPlayerB, "tip", 10, sqlmock.AnyArg(), nil, nil, 0, 0, 0, 2050, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(couplesTxRow(clubID, testPlayerB, "tip", 10, 0, 0, txDate, now))
	mock.ExpectCommit()

	results, err := repo.CreateDepositCouplesMode(context.Background(), clubID, []string{testPlayerB, testPlayerA}, 2050, "Einzahlung", txDate, true)
	if err != nil {
		t.Fatalf("CreateDepositCouplesMode: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 transactions, got %d", len(results))
	}
	want := []struct {
		txType string
		amount int
	}{{"deposit", 1040}, {"deposit", 1000}, {"tip", 10}}
	for i, w := range want {
		if string(results[i].TransactionType) != w.txType || results[i].Amount != w.amount {
			t.Errorf("tx %d: expected %s %d, got %s %d", i, w.txType, w.amount, results[i].TransactionType, results[i].Amount)
		}
	}
}

func TestCreate_RejectsPlayerOfOtherClub(t *testing.T) {
	repo, mock, done := newTestRepo(t)
	defer done()

	clubID := uuid.NewString()
	playerID := testPlayerA
	now := time.Now().UTC()

	mock.ExpectBegin()
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(playerID).
		WillReturnRows(playerRow(playerID, uuid.NewString(), -500, now))
	mock.ExpectRollback()

	_, err := repo.Create(context.Background(), CreateTransactionParams{
		ClubID:          clubID,
		PlayerID:        &playerID,
		TransactionType: TransactionTypeDeposit,
		Amount:          500,
	})
	if !errors.Is(err, ErrPlayerNotInClub) {
		t.Fatalf("expected ErrPlayerNotInClub, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestCreate_RejectsTooLargeAmount(t *testing.T) {
	repo, mock, done := newTestRepo(t)
	defer done()

	_, err := repo.Create(context.Background(), CreateTransactionParams{
		ClubID:          uuid.NewString(),
		TransactionType: TransactionTypeExpense,
		Amount:          -(MaxAmount + 1),
	})
	if !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("expected ErrInvalidAmount, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Deposit and auto-tip are booked in one database transaction based on the
// locked balance, and the club balance is changed by atomic increments.
func TestCreateWithAutoTip_SplitsInOneTransaction(t *testing.T) {
	repo, mock, done := newTestRepo(t)
	defer done()

	clubID := uuid.NewString()
	playerID := testPlayerA
	now := time.Now().UTC()
	txDate := time.Date(2025, 2, 20, 12, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(playerID).
		WillReturnRows(playerRow(playerID, clubID, -300, now))
	// Deposit 300
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(playerID).
		WillReturnRows(playerRow(playerID, clubID, -300, now))
	mock.ExpectQuery("IncrementClubBalance").
		WithArgs(300, sqlmock.AnyArg(), clubID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(1300))
	mock.ExpectQuery("INSERT INTO transactions").
		WithArgs(sqlmock.AnyArg(), clubID, playerID, "deposit", 300, sqlmock.AnyArg(), nil, nil, -300, 0, 1000, 1300, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(couplesTxRow(clubID, playerID, "deposit", 300, -300, 0, txDate, now))
	mock.ExpectExec("UpdatePlayerBalance").
		WithArgs(0, sqlmock.AnyArg(), playerID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Tip 200
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(playerID).
		WillReturnRows(playerRow(playerID, clubID, 0, now))
	mock.ExpectQuery("IncrementClubBalance").
		WithArgs(200, sqlmock.AnyArg(), clubID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(1500))
	mock.ExpectQuery("INSERT INTO transactions").
		WithArgs(sqlmock.AnyArg(), clubID, playerID, "tip", 200, sqlmock.AnyArg(), nil, nil, 0, 0, 1300, 1500, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(couplesTxRow(clubID, playerID, "tip", 200, 0, 0, txDate, now))
	mock.ExpectCommit()

	results, err := repo.CreateWithAutoTip(context.Background(), CreateTransactionParams{
		ClubID:          clubID,
		PlayerID:        &playerID,
		TransactionType: TransactionTypeDeposit,
		Amount:          500,
		TransactionDate: txDate,
	}, true)
	if err != nil {
		t.Fatalf("CreateWithAutoTip: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
	if len(results) != 2 || results[0].Amount != 300 || results[1].Amount != 200 {
		t.Fatalf("unexpected results: %+v", results)
	}
}

// Without auto-tip (club setting off or skipped for this deposit) the full
// amount is booked as one deposit, even beyond the player's debt.
func TestCreateWithAutoTip_DisabledBooksFullDeposit(t *testing.T) {
	repo, mock, done := newTestRepo(t)
	defer done()

	clubID := uuid.NewString()
	playerID := testPlayerA
	now := time.Now().UTC()
	txDate := time.Date(2025, 2, 20, 12, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(playerID).
		WillReturnRows(playerRow(playerID, clubID, -300, now))
	mock.ExpectQuery("IncrementClubBalance").
		WithArgs(500, sqlmock.AnyArg(), clubID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(1500))
	mock.ExpectQuery("INSERT INTO transactions").
		WithArgs(sqlmock.AnyArg(), clubID, playerID, "deposit", 500, sqlmock.AnyArg(), nil, nil, -300, 200, 1000, 1500, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(couplesTxRow(clubID, playerID, "deposit", 500, -300, 200, txDate, now))
	mock.ExpectExec("UpdatePlayerBalance").
		WithArgs(200, sqlmock.AnyArg(), playerID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	results, err := repo.CreateWithAutoTip(context.Background(), CreateTransactionParams{
		ClubID:          clubID,
		PlayerID:        &playerID,
		TransactionType: TransactionTypeDeposit,
		Amount:          500,
		TransactionDate: txDate,
	}, false)
	if err != nil {
		t.Fatalf("CreateWithAutoTip: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
	if len(results) != 1 || results[0].TransactionType != TransactionTypeDeposit || results[0].Amount != 500 {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func couplesTxRow(clubID, playerID, txType string, amount, before, after int, txDate, now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "club_id", "player_id", "transaction_type", "amount", "description", "game_day_fee_id", "game_day_id",
		"player_balance_before", "player_balance_after", "club_balance_before", "club_balance_after",
		"transaction_date", "created_at", "updated_at",
	}).AddRow(uuid.NewString(), clubID, playerID, txType, amount, nil, nil, nil, before, after, 0, 0, txDate, now, now)
}
