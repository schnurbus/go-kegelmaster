package transaction

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

var updateTxColumns = []string{
	"id", "club_id", "player_id", "transaction_type", "amount", "description", "game_day_fee_id", "game_day_id",
	"player_balance_before", "player_balance_after", "club_balance_before", "club_balance_after",
	"transaction_date", "created_at", "updated_at",
}

func storedTxRow(id, clubID string, playerID any, txType string, amount int, gameDayID any, txDate, now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows(updateTxColumns).
		AddRow(id, clubID, playerID, txType, amount, "alt", nil, gameDayID, nil, nil, 0, 0, txDate, now, now)
}

func reloadedTxRow(id, clubID string, playerID any, txType string, amount int, txDate, now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows(append(append([]string{}, updateTxColumns...), "player_name")).
		AddRow(id, clubID, playerID, txType, amount, "neu", nil, nil, nil, nil, 0, 0, txDate, now, now, "Spieler")
}

// Raising a deposit from 5,00 to 8,00 reverses the old booking and applies the
// new one: player -2,00 → -7,00 → +1,00, club +3,00.
func TestUpdate_DepositAmountAdjustsBalances(t *testing.T) {
	repo, mock, done := newTestRepo(t)
	defer done()

	clubID, txID := uuid.NewString(), uuid.NewString()
	playerID := testPlayerA
	now := time.Now().UTC()
	txDate := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	newDate := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("GetTransactionByIDForUpdate").
		WithArgs(txID).
		WillReturnRows(storedTxRow(txID, clubID, playerID, "deposit", 500, nil, txDate, now))
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(playerID).
		WillReturnRows(playerRow(playerID, clubID, -200, now))
	mock.ExpectExec("UpdatePlayerBalance").
		WithArgs(100, sqlmock.AnyArg(), playerID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("IncrementClubBalance").
		WithArgs(300, sqlmock.AnyArg(), clubID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(1300))
	// Snapshots as if booked now: player -700 → 100, club 500 → 1300.
	mock.ExpectQuery("UpdateTransaction").
		WithArgs(txID, playerID, "deposit", 800, "neu", -700, 100, 500, 1300, newDate, sqlmock.AnyArg()).
		WillReturnRows(storedTxRow(txID, clubID, playerID, "deposit", 800, nil, newDate, now))
	mock.ExpectCommit()
	mock.ExpectQuery("GetTransactionByID :one").
		WithArgs(txID).
		WillReturnRows(reloadedTxRow(txID, clubID, playerID, "deposit", 800, newDate, now))

	updated, err := repo.Update(context.Background(), txID, UpdateTransactionParams{
		ClubID:          clubID,
		PlayerID:        &playerID,
		TransactionType: TransactionTypeDeposit,
		Amount:          800,
		Description:     "neu",
		TransactionDate: newDate,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Amount != 800 || updated.PlayerName != "Spieler" {
		t.Errorf("unexpected result: %+v", updated)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Moving a deposit to another player takes it back from the old player and
// books it on the new one; the club only sees the amount difference.
func TestUpdate_DepositToOtherPlayer(t *testing.T) {
	repo, mock, done := newTestRepo(t)
	defer done()

	clubID, txID := uuid.NewString(), uuid.NewString()
	oldPlayer, newPlayer := testPlayerB, testPlayerA
	now := time.Now().UTC()
	txDate := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("GetTransactionByIDForUpdate").
		WithArgs(txID).
		WillReturnRows(storedTxRow(txID, clubID, oldPlayer, "deposit", 500, nil, txDate, now))
	// Locks follow the ID order (A before B).
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(newPlayer).
		WillReturnRows(playerRow(newPlayer, clubID, -300, now))
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(oldPlayer).
		WillReturnRows(playerRow(oldPlayer, clubID, 0, now))
	mock.ExpectExec("UpdatePlayerBalance").
		WithArgs(300, sqlmock.AnyArg(), newPlayer).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UpdatePlayerBalance").
		WithArgs(-500, sqlmock.AnyArg(), oldPlayer).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("IncrementClubBalance").
		WithArgs(100, sqlmock.AnyArg(), clubID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(1100))
	// Date not given: the stored date is kept.
	mock.ExpectQuery("UpdateTransaction").
		WithArgs(txID, newPlayer, "deposit", 600, nil, -300, 300, 500, 1100, txDate, sqlmock.AnyArg()).
		WillReturnRows(storedTxRow(txID, clubID, newPlayer, "deposit", 600, nil, txDate, now))
	mock.ExpectCommit()
	mock.ExpectQuery("GetTransactionByID :one").
		WithArgs(txID).
		WillReturnRows(reloadedTxRow(txID, clubID, newPlayer, "deposit", 600, txDate, now))

	if _, err := repo.Update(context.Background(), txID, UpdateTransactionParams{
		ClubID:          clubID,
		PlayerID:        &newPlayer,
		TransactionType: TransactionTypeDeposit,
		Amount:          600,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Turning a tip into a deposit changes only the player balance; the club
// cash stays the same (zero delta only reads the club).
func TestUpdate_TipToDepositKeepsClubCash(t *testing.T) {
	repo, mock, done := newTestRepo(t)
	defer done()

	clubID, txID := uuid.NewString(), uuid.NewString()
	playerID := testPlayerA
	now := time.Now().UTC()
	txDate := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("GetTransactionByIDForUpdate").
		WithArgs(txID).
		WillReturnRows(storedTxRow(txID, clubID, playerID, "tip", 200, nil, txDate, now))
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(playerID).
		WillReturnRows(playerRow(playerID, clubID, -500, now))
	mock.ExpectExec("UpdatePlayerBalance").
		WithArgs(-300, sqlmock.AnyArg(), playerID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("GetClubByID").
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at", "setup_dismissed_at"}).
			AddRow(clubID, "Club", 2000, 0, 0, true, false, uuid.NewString(), now, now, nil))
	mock.ExpectQuery("UpdateTransaction").
		WithArgs(txID, playerID, "deposit", 200, nil, -500, -300, 1800, 2000, txDate, sqlmock.AnyArg()).
		WillReturnRows(storedTxRow(txID, clubID, playerID, "deposit", 200, nil, txDate, now))
	mock.ExpectCommit()
	mock.ExpectQuery("GetTransactionByID :one").
		WithArgs(txID).
		WillReturnRows(reloadedTxRow(txID, clubID, playerID, "deposit", 200, txDate, now))

	if _, err := repo.Update(context.Background(), txID, UpdateTransactionParams{
		ClubID:          clubID,
		PlayerID:        &playerID,
		TransactionType: TransactionTypeDeposit,
		Amount:          200,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdate_RejectsGameDayFee(t *testing.T) {
	repo, mock, done := newTestRepo(t)
	defer done()

	clubID, txID := uuid.NewString(), uuid.NewString()
	playerID := testPlayerA
	now := time.Now().UTC()

	mock.ExpectBegin()
	mock.ExpectQuery("GetTransactionByIDForUpdate").
		WithArgs(txID).
		WillReturnRows(storedTxRow(txID, clubID, playerID, "fee", -200, uuid.NewString(), now, now))
	mock.ExpectRollback()

	_, err := repo.Update(context.Background(), txID, UpdateTransactionParams{
		ClubID:          clubID,
		PlayerID:        &playerID,
		TransactionType: TransactionTypeDeposit,
		Amount:          200,
	})
	if !errors.Is(err, ErrNotEditable) {
		t.Fatalf("expected ErrNotEditable, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdate_RejectsTransactionOfOtherClub(t *testing.T) {
	repo, mock, done := newTestRepo(t)
	defer done()

	txID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectBegin()
	mock.ExpectQuery("GetTransactionByIDForUpdate").
		WithArgs(txID).
		WillReturnRows(storedTxRow(txID, uuid.NewString(), nil, "expense", -200, nil, now, now))
	mock.ExpectRollback()

	_, err := repo.Update(context.Background(), txID, UpdateTransactionParams{
		ClubID:          uuid.NewString(),
		TransactionType: TransactionTypeExpense,
		Amount:          -300,
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestUpdate_ValidatesBeforeDatabaseAccess(t *testing.T) {
	repo, mock, done := newTestRepo(t)
	defer done()

	playerID := testPlayerA
	cases := []struct {
		name   string
		params UpdateTransactionParams
		want   error
	}{
		{"fee type", UpdateTransactionParams{PlayerID: &playerID, TransactionType: TransactionTypeFee, Amount: -100}, ErrInvalidType},
		{"positive expense", UpdateTransactionParams{TransactionType: TransactionTypeExpense, Amount: 100}, ErrInvalidAmount},
		{"negative deposit", UpdateTransactionParams{PlayerID: &playerID, TransactionType: TransactionTypeDeposit, Amount: -100}, ErrInvalidAmount},
		{"too large", UpdateTransactionParams{PlayerID: &playerID, TransactionType: TransactionTypeTip, Amount: MaxAmount + 1}, ErrInvalidAmount},
		{"deposit without player", UpdateTransactionParams{TransactionType: TransactionTypeDeposit, Amount: 100}, ErrPlayerRequired},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.params.ClubID = uuid.NewString()
			if _, err := repo.Update(context.Background(), uuid.NewString(), tc.params); !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
