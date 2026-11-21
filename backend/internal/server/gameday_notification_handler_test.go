package server

import (
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/handlers"
)

type settledCall struct {
	clubID, gameDayID string
	date              time.Time
}

// recordingNotifier records GameDaySettled calls instead of sending e-mails.
type recordingNotifier struct {
	mu    sync.Mutex
	calls []settledCall
}

func (n *recordingNotifier) GameDaySettled(clubID, gameDayID string, date time.Time) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.calls = append(n.calls, settledCall{clubID: clubID, gameDayID: gameDayID, date: date})
}

func (n *recordingNotifier) Calls() []settledCall {
	n.mu.Lock()
	defer n.mu.Unlock()
	return append([]settledCall(nil), n.calls...)
}

func newNotifyingTestServer(t *testing.T) (*Server, sqlmock.Sqlmock, string, string, *recordingNotifier) {
	t.Helper()
	notifier := &recordingNotifier{}
	srv, mock, authSvc := newTestServer(t, func(d *Dependencies) { d.GameDayNotifier = notifier })
	userID := uuid.NewString()
	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return srv, mock, token, userID, notifier
}

var (
	userCols    = []string{"id", "email", "password_hash", "created_at", "updated_at", "password_changed_at", "notify_gameday_settled", "is_instance_admin", "last_login_at", "disabled_at", "sessions_valid_after", "tour_seen_at", "setup_hint_seen_at"}
	clubCols    = []string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at", "setup_dismissed_at"}
	gameDayCols = []string{"id", "club_id", "date", "notes", "is_draft", "created_at", "updated_at"}
)

// expectOwner mocks the user lookup and the owner permission check.
func expectOwner(mock sqlmock.Sqlmock, userID, clubID string, now time.Time) {
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(userCols).AddRow(userID, "user@example.com", "hash", now, now, nil, false, false, nil, nil, nil, nil, nil))
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows(clubCols).AddRow(clubID, "Test Club", 0, 0, 0, true, false, userID, now, now, nil))
}

// expectPublishWithoutCharges mocks SaveGameDayWithCharges for a club without
// players, so no transaction is booked. saveQuery is the INSERT or UPDATE.
func expectPublishWithoutCharges(mock sqlmock.Sqlmock, userID, clubID, gameDayID, saveQuery string, day, now time.Time) {
	mock.ExpectBegin()
	mock.ExpectQuery("GetPlayersByClubIDForUpdate").WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}))
	mock.ExpectQuery("GetRolesByClubID").WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "pays_base_fee", "created_at", "updated_at"}))
	mock.ExpectQuery("GetClubByID").WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows(clubCols).AddRow(clubID, "Test Club", 0, 0, 0, true, false, userID, now, now, nil))
	mock.ExpectQuery(saveQuery).
		WillReturnRows(sqlmock.NewRows(gameDayCols).AddRow(gameDayID, clubID, day, nil, false, now, now))
	mock.ExpectQuery("ListTransactionsByGameDay").WithArgs(gameDayID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "club_id", "player_id", "transaction_type", "amount", "description", "game_day_fee_id", "game_day_id",
			"player_balance_before", "player_balance_after", "club_balance_before", "club_balance_after",
			"transaction_date", "created_at", "updated_at", "player_name",
		}))
	mock.ExpectQuery("GetGameDayFeesByGameDay").WithArgs(gameDayID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "game_day_participant_id", "penalty_type_id", "penalty_type_name", "penalty_type_description",
			"penalty_type_price", "count", "quantity_scale", "created_at", "updated_at", "player_id", "player_name",
		}))
	mock.ExpectCommit()
}

func TestUpdateGameDay_PublishingDraftNotifies(t *testing.T) {
	srv, mock, token, userID, notifier := newNotifyingTestServer(t)
	clubID, gameDayID := uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	day := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)

	expectOwner(mock, userID, clubID, now)
	mock.ExpectQuery(`FROM game_days WHERE id`).WithArgs(gameDayID).
		WillReturnRows(sqlmock.NewRows(gameDayCols).AddRow(gameDayID, clubID, day, nil, true, now, now))
	expectPublishWithoutCharges(mock, userID, clubID, gameDayID, "UPDATE game_days", day, now)

	req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID+"/gamedays/"+gameDayID, `{"date":"2026-11-14","is_draft":false}`)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	calls := notifier.Calls()
	if len(calls) != 1 || calls[0].clubID != clubID || calls[0].gameDayID != gameDayID || !calls[0].date.Equal(day) {
		t.Fatalf("notifier calls = %+v", calls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateGameDay_EditingPublishedDayDoesNotNotify(t *testing.T) {
	srv, mock, token, userID, notifier := newNotifyingTestServer(t)
	clubID, gameDayID := uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	day := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)

	expectOwner(mock, userID, clubID, now)
	mock.ExpectQuery(`FROM game_days WHERE id`).WithArgs(gameDayID).
		WillReturnRows(sqlmock.NewRows(gameDayCols).AddRow(gameDayID, clubID, day, nil, false, now, now))
	mock.ExpectQuery(`UPDATE game_days`).
		WillReturnRows(sqlmock.NewRows(gameDayCols).AddRow(gameDayID, clubID, day, "Neue Notiz", false, now, now))

	req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID+"/gamedays/"+gameDayID, `{"date":"2026-11-14","notes":"Neue Notiz","is_draft":false}`)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if calls := notifier.Calls(); len(calls) != 0 {
		t.Fatalf("expected no notification, got %+v", calls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateGameDay_SavingDraftDoesNotNotify(t *testing.T) {
	srv, mock, token, userID, notifier := newNotifyingTestServer(t)
	clubID, gameDayID := uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	day := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)

	expectOwner(mock, userID, clubID, now)
	mock.ExpectQuery(`FROM game_days WHERE id`).WithArgs(gameDayID).
		WillReturnRows(sqlmock.NewRows(gameDayCols).AddRow(gameDayID, clubID, day, nil, true, now, now))
	mock.ExpectQuery(`UPDATE game_days`).
		WillReturnRows(sqlmock.NewRows(gameDayCols).AddRow(gameDayID, clubID, day, nil, true, now, now))

	req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID+"/gamedays/"+gameDayID, `{"date":"2026-11-14"}`)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if calls := notifier.Calls(); len(calls) != 0 {
		t.Fatalf("expected no notification, got %+v", calls)
	}
}

func TestCreateGameDay_PublishedNotifies(t *testing.T) {
	srv, mock, token, userID, notifier := newNotifyingTestServer(t)
	clubID, gameDayID := uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	day := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)

	expectOwner(mock, userID, clubID, now)
	expectPublishWithoutCharges(mock, userID, clubID, gameDayID, "INSERT INTO game_days", day, now)

	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/gamedays", `{"date":"2026-11-14","is_draft":false}`)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	if calls := notifier.Calls(); len(calls) != 1 || calls[0].gameDayID != gameDayID {
		t.Fatalf("notifier calls = %+v", calls)
	}
}

func TestCreateGameDay_DraftDoesNotNotify(t *testing.T) {
	srv, mock, token, userID, notifier := newNotifyingTestServer(t)
	clubID, gameDayID := uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	day := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)

	expectOwner(mock, userID, clubID, now)
	mock.ExpectQuery(`INSERT INTO game_days`).
		WillReturnRows(sqlmock.NewRows(gameDayCols).AddRow(gameDayID, clubID, day, nil, true, now, now))

	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/gamedays", `{"date":"2026-11-14"}`)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	if calls := notifier.Calls(); len(calls) != 0 {
		t.Fatalf("expected no notification, got %+v", calls)
	}
}

func TestUpdateCurrentUser_EnablesGameDayNotification(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	now := time.Now().UTC()
	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(userCols).AddRow(userID, "user@example.com", "hash", now, now, nil, false, false, nil, nil, nil, nil, nil))
	mock.ExpectQuery(`UPDATE users\s+SET notify_gameday_settled`).
		WithArgs(userID, true, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(userCols).AddRow(userID, "user@example.com", "hash", now, now, nil, true, false, nil, nil, nil, nil, nil))

	req := mustJSONRequest(t, http.MethodPatch, "/api/auth/me", `{"notify_gameday_settled":true}`)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var payload handlers.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !payload.NotifyGameDaySettled {
		t.Fatalf("expected notify_gameday_settled=true, got %+v", payload)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateCurrentUser_UnchangedSkipsWrite(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	now := time.Now().UTC()
	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(userCols).AddRow(userID, "user@example.com", "hash", now, now, nil, false, false, nil, nil, nil, nil, nil))

	req := mustJSONRequest(t, http.MethodPatch, "/api/auth/me", `{"notify_gameday_settled":false}`)
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

func TestUpdateCurrentUser_RequiresCSRF(t *testing.T) {
	srv, _, authSvc := newTestServer(t)
	token, err := authSvc.GenerateToken(uuid.NewString(), true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	req := mustJSONRequest(t, http.MethodPatch, "/api/auth/me", `{"notify_gameday_settled":true}`)
	req.Header.Del("X-CSRF-Token")
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}
