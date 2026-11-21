package server

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

var txColumns = []string{
	"id", "club_id", "player_id", "transaction_type", "amount", "description", "game_day_fee_id", "game_day_id",
	"player_balance_before", "player_balance_after", "club_balance_before", "club_balance_after",
	"transaction_date", "created_at", "updated_at",
}

// expectUserAndClub mocks the user lookup and the club lookup of the
// permission check; ownerID decides whether the user owns the club.
func expectUserAndClub(mock sqlmock.Sqlmock, userID, clubID, ownerID string, now time.Time) {
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at", "password_changed_at", "notify_gameday_settled", "is_instance_admin", "last_login_at", "disabled_at", "sessions_valid_after"}).
			AddRow(userID, "user@example.com", "hash", now, now, nil, false, false, nil, nil, nil))
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 0, 0, 0, true, false, ownerID, now, now))
}

func TestHandleUpdateTransaction_OwnerUpdatesDeposit(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID, clubID, txID, playerID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	txDate := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	expectUserAndClub(mock, userID, clubID, userID, now)
	mock.ExpectBegin()
	mock.ExpectQuery(`GetTransactionByIDForUpdate`).
		WithArgs(txID).
		WillReturnRows(sqlmock.NewRows(txColumns).
			AddRow(txID, clubID, playerID, "deposit", 500, "", nil, nil, -500, 0, 0, 500, txDate, now, now))
	mock.ExpectQuery(`GetPlayerByIDForUpdate`).
		WithArgs(playerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(playerID, clubID, nil, "00000000-0000-0000-0000-0000000000f1", "Anna", 0, 0, nil, false, nil, now, now))
	mock.ExpectExec(`UpdatePlayerBalance`).
		WithArgs(-200, sqlmock.AnyArg(), playerID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`IncrementClubBalance`).
		WithArgs(-200, sqlmock.AnyArg(), clubID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(300))
	mock.ExpectQuery(`UpdateTransaction`).
		WithArgs(txID, playerID, "deposit", 300, "Korrektur", -500, -200, 0, 300, txDate, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(txColumns).
			AddRow(txID, clubID, playerID, "deposit", 300, "Korrektur", nil, nil, -500, -200, 0, 300, txDate, now, now))
	mock.ExpectCommit()
	mock.ExpectQuery(`GetTransactionByID :one`).
		WithArgs(txID).
		WillReturnRows(sqlmock.NewRows(append(append([]string{}, txColumns...), "player_name")).
			AddRow(txID, clubID, playerID, "deposit", 300, "Korrektur", nil, nil, -500, -200, 0, 300, txDate, now, now, "Anna"))

	body := `{"transaction_type":"deposit","player_id":"` + playerID + `","amount":300,"description":" Korrektur ","transaction_date":"2026-11-03"}`
	req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID+"/transactions/"+txID, body)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["amount"] != float64(300) || payload["player_name"] != "Anna" {
		t.Fatalf("unexpected response: %v", payload)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleUpdateTransaction_ForbiddenWithoutUpdatePermission(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID, clubID, txID, playerID, roleID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	expectUserAndClub(mock, userID, clubID, uuid.NewString(), now)
	mock.ExpectQuery(`GetPlayerByUserIDAndClubID`).
		WithArgs(userID, clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(playerID, clubID, userID, roleID, "Kassenwart", 0, 0, nil, false, nil, now, now))
	mock.ExpectQuery(`HasRolePermission`).
		WithArgs(roleID, "transactions", "update").
		WillReturnRows(sqlmock.NewRows([]string{"?column?"}).AddRow(false))

	req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID+"/transactions/"+txID,
		`{"transaction_type":"deposit","player_id":"`+playerID+`","amount":300}`)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleUpdateTransaction_RejectsGameDayFee(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID, clubID, txID, playerID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	expectUserAndClub(mock, userID, clubID, userID, now)
	mock.ExpectBegin()
	mock.ExpectQuery(`GetTransactionByIDForUpdate`).
		WithArgs(txID).
		WillReturnRows(sqlmock.NewRows(txColumns).
			AddRow(txID, clubID, playerID, "fee", -200, "Pumpe", uuid.NewString(), uuid.NewString(), 0, -200, 0, 0, now, now, now))
	mock.ExpectRollback()

	req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID+"/transactions/"+txID,
		`{"transaction_type":"deposit","player_id":"`+playerID+`","amount":200}`)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleUpdateTransaction_ValidationErrors(t *testing.T) {
	cases := map[string]string{
		"fee type":               `{"transaction_type":"fee","player_id":"p","amount":200}`,
		"zero amount":            `{"transaction_type":"expense","amount":0}`,
		"deposit without player": `{"transaction_type":"deposit","amount":200}`,
		"bad date":               `{"transaction_type":"expense","amount":200,"transaction_date":"03.11.2026"}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			srv, mock, authSvc := newTestServer(t)
			userID, clubID := uuid.NewString(), uuid.NewString()
			now := time.Now().UTC()
			token, err := authSvc.GenerateToken(userID, true)
			if err != nil {
				t.Fatalf("generate token: %v", err)
			}
			expectUserAndClub(mock, userID, clubID, userID, now)

			req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID+"/transactions/"+uuid.NewString(), body)
			req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
			resp := doRequest(t, srv, req)
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", resp.StatusCode)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		})
	}
}
