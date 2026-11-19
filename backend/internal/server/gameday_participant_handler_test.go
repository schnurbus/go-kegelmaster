package server

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// expectOwnerWithGameDay mocks the user lookup, the owner permission check and
// the game day lookup shared by the participant handlers.
func expectOwnerWithGameDay(mock sqlmock.Sqlmock, userID, clubID, gameDayID string, now time.Time) {
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at", "password_changed_at", "notify_gameday_settled", "is_instance_admin", "last_login_at", "disabled_at", "sessions_valid_after"}).
			AddRow(userID, "user@example.com", "hash", now, now, nil, false, false, nil, nil, nil))
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 0, 0, 0, true, false, userID, now, now))
	mock.ExpectQuery(`SELECT id, club_id, date, notes, is_draft, created_at, updated_at FROM game_days WHERE id`).
		WithArgs(gameDayID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "date", "notes", "is_draft", "created_at", "updated_at"}).
			AddRow(gameDayID, clubID, now, "", false, now, now))
}

func TestHandleSetParticipantAttendance_AbsentDeletesCompetitionValues(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID, clubID, gameDayID, playerID, participantID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	expectOwnerWithGameDay(mock, userID, clubID, gameDayID, now)
	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE game_day_participants`).
		WithArgs(gameDayID, playerID, false).
		WillReturnRows(sqlmock.NewRows([]string{"id", "game_day_id", "player_id", "created_at", "attended"}).
			AddRow(participantID, gameDayID, playerID, now, false))
	mock.ExpectExec(`DELETE FROM game_day_competition_values WHERE game_day_participant_id`).
		WithArgs(participantID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	req := mustJSONRequest(t, http.MethodPatch, "/api/clubs/"+clubID+"/gamedays/"+gameDayID+"/participants/"+playerID, `{"attended":false}`)
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
	if payload["attended"] != false {
		t.Fatalf("expected attended=false, got %v", payload["attended"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleSetParticipantAttendance_PresentKeepsCompetitionValues(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID, clubID, gameDayID, playerID, participantID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	expectOwnerWithGameDay(mock, userID, clubID, gameDayID, now)
	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE game_day_participants`).
		WithArgs(gameDayID, playerID, true).
		WillReturnRows(sqlmock.NewRows([]string{"id", "game_day_id", "player_id", "created_at", "attended"}).
			AddRow(participantID, gameDayID, playerID, now, true))
	mock.ExpectCommit()

	req := mustJSONRequest(t, http.MethodPatch, "/api/clubs/"+clubID+"/gamedays/"+gameDayID+"/participants/"+playerID, `{"attended":true}`)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleSetParticipantAttendance_MissingFieldRejected(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID, clubID, gameDayID, playerID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at", "password_changed_at", "notify_gameday_settled", "is_instance_admin", "last_login_at", "disabled_at", "sessions_valid_after"}).
			AddRow(userID, "user@example.com", "hash", now, now, nil, false, false, nil, nil, nil))
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 0, 0, 0, true, false, userID, now, now))

	req := mustJSONRequest(t, http.MethodPatch, "/api/clubs/"+clubID+"/gamedays/"+gameDayID+"/participants/"+playerID, `{}`)
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
