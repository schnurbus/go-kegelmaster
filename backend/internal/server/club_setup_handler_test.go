package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

var setupCountCols = []string{"penalty_types", "roles", "players", "published_game_days", "invitations", "competitions", "own_players"}

type setupStep struct {
	ID       string `json:"id"`
	Required bool   `json:"required"`
	Done     bool   `json:"done"`
}

type setupStatus struct {
	Dismissed bool        `json:"dismissed"`
	Complete  bool        `json:"complete"`
	Steps     []setupStep `json:"steps"`
}

// expectClubOwnedBy mocks the user lookup and the club row (owner and setup_dismissed_at).
func expectClubOwnedBy(mock sqlmock.Sqlmock, userID, clubID, ownerID string, dismissedAt any, now time.Time) {
	mock.ExpectQuery(`SELECT id, email, password_hash`).WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(userColumns).AddRow(userID, "user@example.com", "hash", now, now, nil, false, false, nil, nil, nil, now, now))
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee`).WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows(clubCols).AddRow(clubID, "KC Test", 0, 0, 0, true, false, ownerID, now, now, dismissedAt))
}

func getSetupStatus(t *testing.T, srv *Server, token, clubID string) (*http.Response, setupStatus) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, "/api/clubs/"+clubID+"/setup", nil)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	resp := doRequest(t, srv, req)
	var status setupStatus
	if resp.StatusCode == http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			t.Fatalf("decode: %v", err)
		}
	}
	resp.Body.Close()
	return resp, status
}

func stepsByID(steps []setupStep) map[string]setupStep {
	m := make(map[string]setupStep, len(steps))
	for _, s := range steps {
		m[s.ID] = s
	}
	return m
}

func TestClubSetup_EmptyClub(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, clubID := uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	expectClubOwnedBy(mock, userID, clubID, userID, nil, now)
	mock.ExpectQuery(`AS penalty_types`).WithArgs(clubID, userID).
		WillReturnRows(sqlmock.NewRows(setupCountCols).AddRow(0, 0, 0, 0, 0, 0, 0))

	resp, status := getSetupStatus(t, srv, token, clubID)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	wantOrder := []string{"club", "penalty_types", "roles", "players", "first_game_day", "own_player", "invitations", "competitions"}
	if len(status.Steps) != len(wantOrder) {
		t.Fatalf("expected %d steps, got %+v", len(wantOrder), status.Steps)
	}
	for i, id := range wantOrder {
		if status.Steps[i].ID != id {
			t.Fatalf("step %d: expected %s, got %s", i, id, status.Steps[i].ID)
		}
	}
	steps := stepsByID(status.Steps)
	if !steps["club"].Done || !steps["club"].Required {
		t.Fatalf("club step must be required and done: %+v", steps["club"])
	}
	for _, id := range []string{"penalty_types", "roles", "players", "first_game_day"} {
		if !steps[id].Required || steps[id].Done {
			t.Fatalf("%s must be required and open: %+v", id, steps[id])
		}
	}
	for _, id := range []string{"own_player", "invitations", "competitions"} {
		if steps[id].Required || steps[id].Done {
			t.Fatalf("%s must be optional and open: %+v", id, steps[id])
		}
	}
	if status.Complete || status.Dismissed {
		t.Fatalf("expected incomplete, not dismissed: %+v", status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestClubSetup_PartlySetUp(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, clubID := uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	expectClubOwnedBy(mock, userID, clubID, userID, nil, now)
	// Strafentyp und Rolle vorhanden, noch kein Spieler, nur ein Entwurf (zählt nicht).
	mock.ExpectQuery(`AS penalty_types`).WithArgs(clubID, userID).
		WillReturnRows(sqlmock.NewRows(setupCountCols).AddRow(2, 4, 0, 0, 0, 1, 0))

	_, status := getSetupStatus(t, srv, token, clubID)
	steps := stepsByID(status.Steps)
	if !steps["penalty_types"].Done || !steps["roles"].Done || !steps["competitions"].Done {
		t.Fatalf("expected penalty_types, roles, competitions done: %+v", steps)
	}
	if steps["players"].Done || steps["first_game_day"].Done || status.Complete {
		t.Fatalf("expected players and first_game_day open: %+v", status)
	}
}

// A club whose only game day is a draft has not done the step "erster
// Spieltag": the counts query only counts published game days.
func TestClubSetup_DraftGameDayDoesNotCount(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, clubID := uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	expectClubOwnedBy(mock, userID, clubID, userID, nil, now)
	mock.ExpectQuery(`NOT g\.is_draft`).WithArgs(clubID, userID).
		WillReturnRows(sqlmock.NewRows(setupCountCols).AddRow(1, 1, 3, 0, 0, 0, 1))

	resp, status := getSetupStatus(t, srv, token, clubID)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	steps := stepsByID(status.Steps)
	if steps["first_game_day"].Done || status.Complete {
		t.Fatalf("a draft must not complete first_game_day: %+v", status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestClubSetup_CompleteWithoutOwnPlayer(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, clubID := uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	expectClubOwnedBy(mock, userID, clubID, userID, nil, now)
	// Spieler eines anderen Benutzers: zählt für "Spieler", nicht für "Eigener Spieler".
	mock.ExpectQuery(`AS penalty_types`).WithArgs(clubID, userID).
		WillReturnRows(sqlmock.NewRows(setupCountCols).AddRow(1, 1, 3, 1, 2, 0, 0))

	_, status := getSetupStatus(t, srv, token, clubID)
	steps := stepsByID(status.Steps)
	if !status.Complete {
		t.Fatalf("expected complete: %+v", status)
	}
	if !steps["players"].Done || steps["own_player"].Done {
		t.Fatalf("players done, own_player open expected: %+v", steps)
	}
	if !steps["invitations"].Done {
		t.Fatalf("invitations expected done: %+v", steps)
	}
}

func TestClubSetup_DismissedClub(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, clubID := uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	expectClubOwnedBy(mock, userID, clubID, userID, now, now)
	mock.ExpectQuery(`AS penalty_types`).WithArgs(clubID, userID).
		WillReturnRows(sqlmock.NewRows(setupCountCols).AddRow(0, 0, 0, 0, 0, 0, 0))

	_, status := getSetupStatus(t, srv, token, clubID)
	if !status.Dismissed {
		t.Fatalf("expected dismissed: %+v", status)
	}
}

func TestClubSetup_NonOwnerForbidden(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, ownerID, clubID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	expectClubOwnedBy(mock, userID, clubID, ownerID, nil, now)

	resp, _ := getSetupStatus(t, srv, token, clubID)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestClubSetup_UnknownClub(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, clubID := uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	mock.ExpectQuery(`SELECT id, email, password_hash`).WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(userColumns).AddRow(userID, "user@example.com", "hash", now, now, nil, false, false, nil, nil, nil, now, now))
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee`).WithArgs(clubID).WillReturnError(sql.ErrNoRows)

	resp, _ := getSetupStatus(t, srv, token, clubID)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func dismissSetupRequest(t *testing.T, token, clubID string) *http.Request {
	t.Helper()
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/setup/dismiss", "")
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	return req
}

func TestClubSetupDismiss_Owner(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, clubID := uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	expectClubOwnedBy(mock, userID, clubID, userID, nil, now)
	mock.ExpectExec(`UPDATE clubs`).WithArgs(sqlmock.AnyArg(), clubID).WillReturnResult(sqlmock.NewResult(0, 1))

	resp := doRequest(t, srv, dismissSetupRequest(t, token, clubID))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestClubSetupDismiss_NonOwnerForbidden(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, ownerID, clubID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	expectClubOwnedBy(mock, userID, clubID, ownerID, nil, now)

	resp := doRequest(t, srv, dismissSetupRequest(t, token, clubID))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
