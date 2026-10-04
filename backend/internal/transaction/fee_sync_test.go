package transaction

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestFeeAmountAndDescription(t *testing.T) {
	if got := FeeAmount(100, 2, 1); got != -200 {
		t.Fatalf("amount = %d, want -200", got)
	}
	if got := FeeAmount(50, 250, 100); got != -125 {
		t.Fatalf("decimal amount = %d, want -125", got)
	}
	if got := FeeAmount(100, 0, 1); got != 0 {
		t.Fatalf("zero count amount = %d, want 0", got)
	}
	if got := FeeTransactionDescription("Pudel", 2, 1); got != "Pudel ×2" {
		t.Fatalf("description = %q", got)
	}
	if got := FeeTransactionDescription("Pudel", 250, 100); got != "Pudel ×2.50" {
		t.Fatalf("decimal description = %q", got)
	}
}

func TestSyncParticipantFees_UnchangedSkipsWrites(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db, nil, nil, nil)
	now := time.Now().UTC()
	clubID := uuid.NewString()
	playerID := uuid.NewString()
	gameDayID := uuid.NewString()
	participantID := uuid.NewString()
	feeID := uuid.NewString()
	penaltyTypeID := uuid.NewString()

	mock.ExpectBegin()
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(playerID).
		WillReturnRows(playerRow(playerID, clubID, -200, now))
	mock.ExpectQuery("GetGameDayFeesByParticipantForUpdate").
		WithArgs(participantID).
		WillReturnRows(feeRows().AddRow(feeID, participantID, penaltyTypeID, "Pudel", nil, 100, 2, 1, now, now))
	mock.ExpectQuery("GetTransactionByGameDayFee").
		WithArgs(feeID).
		WillReturnRows(txRow(uuid.NewString(), clubID, playerID, feeID, gameDayID, -200, "Pudel ×2", -200, -200, now))
	mock.ExpectCommit()

	err = repo.SyncParticipantFees(context.Background(), SyncParticipantFeesParams{
		ClubID:             clubID,
		PlayerID:           playerID,
		GameDayID:          gameDayID,
		ParticipantID:      participantID,
		TransactionDate:    now,
		RecordTransactions: true,
		Fees: []ParticipantFeeChange{{
			PenaltyTypeID:    penaltyTypeID,
			PenaltyTypeName:  "Pudel",
			PenaltyTypePrice: 100,
			Count:            2,
			QuantityScale:    1,
		}},
	})
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSyncParticipantFees_CreatesFeeAndTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db, nil, nil, nil)
	now := time.Now().UTC()
	clubID := uuid.NewString()
	playerID := uuid.NewString()
	gameDayID := uuid.NewString()
	participantID := uuid.NewString()
	penaltyTypeID := uuid.NewString()
	createdFeeID := uuid.NewString()

	mock.ExpectBegin()
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(playerID).
		WillReturnRows(playerRow(playerID, clubID, 0, now))
	mock.ExpectQuery("GetGameDayFeesByParticipantForUpdate").
		WithArgs(participantID).
		WillReturnRows(feeRows())
	mock.ExpectQuery("INSERT INTO game_day_fees").
		WillReturnRows(feeRows().AddRow(createdFeeID, participantID, penaltyTypeID, "Pudel", nil, 100, 2, 1, now, now))
	mock.ExpectQuery("GetClubByID").
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Club", 5000, 0, 0, false, false, uuid.NewString(), now, now))
	mock.ExpectQuery("INSERT INTO transactions").
		WithArgs(sqlmock.AnyArg(), clubID, playerID, "fee", -200, sqlmock.AnyArg(), createdFeeID, gameDayID, 0, -200, 5000, 5000, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(txRow(uuid.NewString(), clubID, playerID, createdFeeID, gameDayID, -200, "Pudel ×2", 0, -200, now))
	mock.ExpectExec("UpdatePlayerBalance").
		WithArgs(-200, sqlmock.AnyArg(), playerID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = repo.SyncParticipantFees(context.Background(), SyncParticipantFeesParams{
		ClubID:             clubID,
		PlayerID:           playerID,
		GameDayID:          gameDayID,
		ParticipantID:      participantID,
		TransactionDate:    now,
		RecordTransactions: true,
		Fees: []ParticipantFeeChange{{
			PenaltyTypeID:    penaltyTypeID,
			PenaltyTypeName:  "Pudel",
			PenaltyTypePrice: 100,
			Count:            2,
			QuantityScale:    1,
		}},
	})
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSyncParticipantFees_ReplacesFeeInOneBalanceUpdate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db, nil, nil, nil)
	now := time.Now().UTC()
	clubID := uuid.NewString()
	playerID := uuid.NewString()
	gameDayID := uuid.NewString()
	participantID := uuid.NewString()
	penaltyTypeID := uuid.NewString()
	feeID := uuid.NewString()
	oldTxID := uuid.NewString()

	mock.ExpectBegin()
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(playerID).
		WillReturnRows(playerRow(playerID, clubID, -100, now))
	mock.ExpectQuery("GetGameDayFeesByParticipantForUpdate").
		WithArgs(participantID).
		WillReturnRows(feeRows().AddRow(feeID, participantID, penaltyTypeID, "Pudel", nil, 100, 1, 1, now, now))
	mock.ExpectQuery("GetTransactionByGameDayFee").
		WithArgs(feeID).
		WillReturnRows(txRow(oldTxID, clubID, playerID, feeID, gameDayID, -100, "Pudel ×1", -100, -100, now))
	mock.ExpectQuery("INSERT INTO game_day_fees").
		WillReturnRows(feeRows().AddRow(feeID, participantID, penaltyTypeID, "Pudel", nil, 100, 3, 1, now, now))
	mock.ExpectExec("DeleteTransaction").
		WithArgs(oldTxID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("GetClubByID").
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Club", 5000, 0, 0, false, false, uuid.NewString(), now, now))
	mock.ExpectQuery("INSERT INTO transactions").
		WithArgs(sqlmock.AnyArg(), clubID, playerID, "fee", -300, sqlmock.AnyArg(), feeID, gameDayID, 0, -300, 5000, 5000, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(txRow(uuid.NewString(), clubID, playerID, feeID, gameDayID, -300, "Pudel ×3", 0, -300, now))
	mock.ExpectExec("UpdatePlayerBalance").
		WithArgs(-300, sqlmock.AnyArg(), playerID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = repo.SyncParticipantFees(context.Background(), SyncParticipantFeesParams{
		ClubID:             clubID,
		PlayerID:           playerID,
		GameDayID:          gameDayID,
		ParticipantID:      participantID,
		TransactionDate:    now,
		RecordTransactions: true,
		Fees: []ParticipantFeeChange{{
			PenaltyTypeID:    penaltyTypeID,
			PenaltyTypeName:  "Pudel",
			PenaltyTypePrice: 100,
			Count:            3,
			QuantityScale:    1,
		}},
	})
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSyncParticipantFees_DeletesFeeAndRevertsBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db, nil, nil, nil)
	now := time.Now().UTC()
	clubID := uuid.NewString()
	playerID := uuid.NewString()
	gameDayID := uuid.NewString()
	participantID := uuid.NewString()
	penaltyTypeID := uuid.NewString()
	feeID := uuid.NewString()
	oldTxID := uuid.NewString()

	mock.ExpectBegin()
	mock.ExpectQuery("GetPlayerByIDForUpdate").
		WithArgs(playerID).
		WillReturnRows(playerRow(playerID, clubID, -200, now))
	mock.ExpectQuery("GetGameDayFeesByParticipantForUpdate").
		WithArgs(participantID).
		WillReturnRows(feeRows().AddRow(feeID, participantID, penaltyTypeID, "Pudel", nil, 100, 2, 1, now, now))
	mock.ExpectQuery("GetTransactionByGameDayFee").
		WithArgs(feeID).
		WillReturnRows(txRow(oldTxID, clubID, playerID, feeID, gameDayID, -200, "Pudel ×2", -200, -200, now))
	mock.ExpectExec("DeleteTransaction").
		WithArgs(oldTxID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DeleteGameDayFeeByParticipantAndType").
		WithArgs(participantID, penaltyTypeID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UpdatePlayerBalance").
		WithArgs(0, sqlmock.AnyArg(), playerID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = repo.SyncParticipantFees(context.Background(), SyncParticipantFeesParams{
		ClubID:             clubID,
		PlayerID:           playerID,
		GameDayID:          gameDayID,
		ParticipantID:      participantID,
		TransactionDate:    now,
		RecordTransactions: true,
		Fees: []ParticipantFeeChange{{
			PenaltyTypeID: penaltyTypeID,
			Count:         0,
		}},
	})
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSyncParticipantFees_DraftDoesNotTouchTransactions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db, nil, nil, nil)
	now := time.Now().UTC()
	participantID := uuid.NewString()
	penaltyTypeID := uuid.NewString()
	feeID := uuid.NewString()

	mock.ExpectBegin()
	mock.ExpectQuery("GetGameDayFeesByParticipantForUpdate").
		WithArgs(participantID).
		WillReturnRows(feeRows())
	mock.ExpectQuery("INSERT INTO game_day_fees").
		WillReturnRows(feeRows().AddRow(feeID, participantID, penaltyTypeID, "Pudel", nil, 100, 1, 1, now, now))
	mock.ExpectCommit()

	err = repo.SyncParticipantFees(context.Background(), SyncParticipantFeesParams{
		ClubID:             uuid.NewString(),
		PlayerID:           uuid.NewString(),
		GameDayID:          uuid.NewString(),
		ParticipantID:      participantID,
		RecordTransactions: false,
		Fees: []ParticipantFeeChange{{
			PenaltyTypeID:    penaltyTypeID,
			PenaltyTypeName:  "Pudel",
			PenaltyTypePrice: 100,
			Count:            1,
			QuantityScale:    1,
		}},
	})
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func playerRow(playerID, clubID string, balance int, now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
		AddRow(playerID, clubID, nil, nil, "Spieler", balance, 0, nil, false, nil, now, now)
}

func feeRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "game_day_participant_id", "penalty_type_id", "penalty_type_name", "penalty_type_description",
		"penalty_type_price", "count", "quantity_scale", "created_at", "updated_at",
	})
}

func txRow(id, clubID, playerID, feeID, gameDayID string, amount int, description string, before, after int, now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "club_id", "player_id", "transaction_type", "amount", "description", "game_day_fee_id", "game_day_id",
		"player_balance_before", "player_balance_after", "club_balance_before", "club_balance_after",
		"transaction_date", "created_at", "updated_at",
	}).AddRow(id, clubID, playerID, "fee", amount, description, feeID, gameDayID, before, after, 5000, 5000, now, now, now)
}
