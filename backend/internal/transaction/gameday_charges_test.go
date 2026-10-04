package transaction

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestSaveGameDayWithCharges_BooksEveryBaseFeeInOneTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db, nil, nil, nil)
	now := time.Now().UTC()
	older := now.Add(-time.Hour)
	clubID := uuid.NewString()
	roleID := uuid.NewString()
	playerA := "player-a"
	playerB := "player-b"
	gameDayID := uuid.NewString()
	day := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("GetPlayersByClubIDForUpdate").
		WithArgs(clubID).
		WillReturnRows(playerRows().
			AddRow(playerB, clubID, nil, roleID, "Berta", 0, 0, nil, false, nil, now, now).
			AddRow(playerA, clubID, nil, roleID, "Anna", 0, 0, nil, false, nil, older, now))
	mock.ExpectQuery("GetRolesByClubID").
		WithArgs(clubID).
		WillReturnRows(roleRows().AddRow(roleID, clubID, "Spieler", true, now, now))
	expectClub(mock, clubID, 500, now)
	mock.ExpectQuery("INSERT INTO game_days").
		WillReturnRows(gameDayRows().AddRow(gameDayID, clubID, day, nil, false, now, now))
	mock.ExpectQuery("ListTransactionsByGameDay").
		WithArgs(gameDayID).
		WillReturnRows(gameDayTxRows())
	mock.ExpectQuery("GetGameDayFeesByGameDay").
		WithArgs(gameDayID).
		WillReturnRows(gameDayFeeRows())
	// Newer player first, then the older one. Balances both go from 0 to -500.
	expectBaseFeeInsert(mock, clubID, playerB, gameDayID, -500, 0, -500, 1000)
	expectBaseFeeInsert(mock, clubID, playerA, gameDayID, -500, 0, -500, 1000)
	mock.ExpectExec("UpdatePlayerBalance").
		WithArgs(-500, sqlmock.AnyArg(), playerA).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UpdatePlayerBalance").
		WithArgs(-500, sqlmock.AnyArg(), playerB).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	saved, err := repo.SaveGameDayWithCharges(context.Background(), SaveGameDayWithChargesParams{
		ClubID:     clubID,
		Date:       day,
		Notes:      "",
		IsDraft:    false,
		ChargeFees: true,
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if saved.ID != gameDayID || saved.IsDraft {
		t.Fatalf("saved game day = %+v", saved)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSaveGameDayWithCharges_RollsBackWhenAChargeFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db, nil, nil, nil)
	now := time.Now().UTC()
	clubID := uuid.NewString()
	roleID := uuid.NewString()
	playerID := uuid.NewString()
	gameDayID := uuid.NewString()
	day := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("GetPlayersByClubIDForUpdate").
		WithArgs(clubID).
		WillReturnRows(playerRows().AddRow(playerID, clubID, nil, roleID, "Anna", 0, 0, nil, false, nil, now, now))
	mock.ExpectQuery("GetRolesByClubID").
		WithArgs(clubID).
		WillReturnRows(roleRows().AddRow(roleID, clubID, "Spieler", true, now, now))
	expectClub(mock, clubID, 500, now)
	mock.ExpectQuery("INSERT INTO game_days").
		WillReturnRows(gameDayRows().AddRow(gameDayID, clubID, day, nil, false, now, now))
	mock.ExpectQuery("ListTransactionsByGameDay").
		WithArgs(gameDayID).
		WillReturnRows(gameDayTxRows())
	mock.ExpectQuery("GetGameDayFeesByGameDay").
		WithArgs(gameDayID).
		WillReturnRows(gameDayFeeRows())
	mock.ExpectQuery("INSERT INTO transactions").
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	_, err = repo.SaveGameDayWithCharges(context.Background(), SaveGameDayWithChargesParams{
		ClubID:     clubID,
		Date:       day,
		IsDraft:    false,
		ChargeFees: true,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var chargeErr *ChargeError
	if errors.As(err, &chargeErr) {
		t.Fatalf("database failure surfaced as user error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSaveGameDayWithCharges_SkipsPlayersWhoDoNotPay(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db, nil, nil, nil)
	now := time.Now().UTC()
	clubID := uuid.NewString()
	payingRole := uuid.NewString()
	freeRole := uuid.NewString()
	payerID := "player-payer"
	inactiveID := "player-inactive"
	freeID := "player-free"
	gameDayID := uuid.NewString()
	day := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("GetPlayersByClubIDForUpdate").
		WithArgs(clubID).
		WillReturnRows(playerRows().
			AddRow(inactiveID, clubID, nil, payingRole, "Inaktiv", 0, 0, nil, true, nil, now, now).
			AddRow(freeID, clubID, nil, freeRole, "Gast", 0, 0, nil, false, nil, now, now).
			AddRow(payerID, clubID, nil, payingRole, "Anna", 0, 0, nil, false, nil, now, now))
	mock.ExpectQuery("GetRolesByClubID").
		WithArgs(clubID).
		WillReturnRows(roleRows().
			AddRow(payingRole, clubID, "Aktiv", true, now, now).
			AddRow(freeRole, clubID, "Gast", false, now, now))
	expectClub(mock, clubID, 500, now)
	mock.ExpectQuery("INSERT INTO game_days").
		WillReturnRows(gameDayRows().AddRow(gameDayID, clubID, day, nil, false, now, now))
	mock.ExpectQuery("ListTransactionsByGameDay").
		WithArgs(gameDayID).
		WillReturnRows(gameDayTxRows())
	mock.ExpectQuery("GetGameDayFeesByGameDay").
		WithArgs(gameDayID).
		WillReturnRows(gameDayFeeRows())
	expectBaseFeeInsert(mock, clubID, payerID, gameDayID, -500, 0, -500, 1000)
	mock.ExpectExec("UpdatePlayerBalance").
		WithArgs(-500, sqlmock.AnyArg(), payerID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if _, err := repo.SaveGameDayWithCharges(context.Background(), SaveGameDayWithChargesParams{
		ClubID:     clubID,
		Date:       day,
		IsDraft:    false,
		ChargeFees: true,
	}); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSaveGameDayWithCharges_BooksPenaltyFeesWhenPublishing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db, nil, nil, nil)
	now := time.Now().UTC()
	clubID := uuid.NewString()
	playerID := uuid.NewString()
	roleID := uuid.NewString()
	gameDayID := uuid.NewString()
	feeID := uuid.NewString()
	day := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("GetPlayersByClubIDForUpdate").
		WithArgs(clubID).
		WillReturnRows(playerRows().AddRow(playerID, clubID, nil, roleID, "Anna", -100, 0, nil, false, nil, now, now))
	mock.ExpectQuery("GetRolesByClubID").
		WithArgs(clubID).
		WillReturnRows(roleRows().AddRow(roleID, clubID, "Aktiv", true, now, now))
	expectClub(mock, clubID, 0, now)
	mock.ExpectQuery("UPDATE game_days").
		WillReturnRows(gameDayRows().AddRow(gameDayID, clubID, day, "Notiz", false, now, now))
	mock.ExpectQuery("ListTransactionsByGameDay").
		WithArgs(gameDayID).
		WillReturnRows(gameDayTxRows())
	mock.ExpectQuery("GetGameDayFeesByGameDay").
		WithArgs(gameDayID).
		WillReturnRows(gameDayFeeRows().AddRow(feeID, uuid.NewString(), uuid.NewString(), "Pudel", nil, 100, 2, 1, now, now, playerID, "Anna"))
	mock.ExpectQuery("INSERT INTO transactions").
		WithArgs(sqlmock.AnyArg(), clubID, playerID, "fee", -200, sqlmock.AnyArg(), feeID, gameDayID, -100, -300, 1000, 1000, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(txRow(uuid.NewString(), clubID, playerID, feeID, gameDayID, -200, "Pudel ×2", -100, -300, now))
	mock.ExpectExec("UpdatePlayerBalance").
		WithArgs(-300, sqlmock.AnyArg(), playerID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	saved, err := repo.SaveGameDayWithCharges(context.Background(), SaveGameDayWithChargesParams{
		ID:         gameDayID,
		ClubID:     clubID,
		Date:       day,
		Notes:      "Notiz",
		IsDraft:    false,
		ChargeFees: true,
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if saved.Notes != "Notiz" || saved.IsDraft {
		t.Fatalf("saved game day = %+v", saved)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSaveGameDayWithCharges_MissingRoleRollsBack(t *testing.T) {
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
	day := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery("GetPlayersByClubIDForUpdate").
		WithArgs(clubID).
		WillReturnRows(playerRows().AddRow(playerID, clubID, nil, uuid.NewString(), "Anna", 0, 0, nil, false, nil, now, now))
	mock.ExpectQuery("GetRolesByClubID").
		WithArgs(clubID).
		WillReturnRows(roleRows())
	expectClub(mock, clubID, 500, now)
	mock.ExpectQuery("INSERT INTO game_days").
		WillReturnRows(gameDayRows().AddRow(gameDayID, clubID, day, nil, false, now, now))
	mock.ExpectQuery("ListTransactionsByGameDay").
		WithArgs(gameDayID).
		WillReturnRows(gameDayTxRows())
	mock.ExpectQuery("GetGameDayFeesByGameDay").
		WithArgs(gameDayID).
		WillReturnRows(gameDayFeeRows())
	mock.ExpectRollback()

	_, err = repo.SaveGameDayWithCharges(context.Background(), SaveGameDayWithChargesParams{
		ClubID:     clubID,
		Date:       day,
		IsDraft:    false,
		ChargeFees: true,
	})
	var chargeErr *ChargeError
	if !errors.As(err, &chargeErr) {
		t.Fatalf("expected charge error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func expectClub(mock sqlmock.Sqlmock, clubID string, baseFee int, now time.Time) {
	mock.ExpectQuery("GetClubByID").
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Club", 1000, 0, baseFee, false, false, uuid.NewString(), now, now))
}

func expectBaseFeeInsert(mock sqlmock.Sqlmock, clubID, playerID, gameDayID string, amount, before, after, clubBalance int) {
	mock.ExpectQuery("INSERT INTO transactions").
		WithArgs(sqlmock.AnyArg(), clubID, playerID, "base_fee", amount, sqlmock.AnyArg(), nil, gameDayID, before, after, clubBalance, clubBalance, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(txRow(uuid.NewString(), clubID, playerID, "", gameDayID, amount, "Grundgebühr", before, after, time.Now()))
}

func playerRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"})
}

func roleRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "club_id", "name", "pays_base_fee", "created_at", "updated_at"})
}

func gameDayRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "club_id", "date", "notes", "is_draft", "created_at", "updated_at"})
}

func gameDayTxRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "club_id", "player_id", "transaction_type", "amount", "description", "game_day_fee_id", "game_day_id",
		"player_balance_before", "player_balance_after", "club_balance_before", "club_balance_after",
		"transaction_date", "created_at", "updated_at", "player_name",
	})
}

func gameDayFeeRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "game_day_participant_id", "penalty_type_id", "penalty_type_name", "penalty_type_description",
		"penalty_type_price", "count", "quantity_scale", "created_at", "updated_at", "player_id", "player_name",
	})
}
