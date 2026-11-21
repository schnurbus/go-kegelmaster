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

var playerCols = []string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}

func createPlayerRequest(t *testing.T, token, clubID, body string) *http.Request {
	t.Helper()
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/players", body)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	return req
}

func TestCreatePlayer_LinkMeLinksCaller(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, clubID, roleID, playerID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	expectClubOwnedBy(mock, userID, clubID, userID, nil, now)
	mock.ExpectQuery(`SELECT id, club_id, name, pays_base_fee`).WithArgs(roleID).
		WillReturnRows(sqlmock.NewRows(roleCols).AddRow(roleID, clubID, "Präsident", true, now, now))
	mock.ExpectQuery(`FROM players`).WithArgs(userID, clubID).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO players`).
		WithArgs(sqlmock.AnyArg(), clubID, userID, roleID, "Hans", 0, 0, sqlmock.AnyArg(), false, nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(playerCols).AddRow(playerID, clubID, userID, roleID, "Hans", 0, 0, nil, false, nil, now, now))

	resp := doRequest(t, srv, createPlayerRequest(t, token, clubID, `{"name":"Hans","role_id":"`+roleID+`","link_me":true}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body["user_id"] != userID {
		t.Fatalf("expected player linked to %s, got %v", userID, body["user_id"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreatePlayer_LinkMeConflictWhenCallerHasPlayer(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, clubID, roleID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	expectClubOwnedBy(mock, userID, clubID, userID, nil, now)
	mock.ExpectQuery(`SELECT id, club_id, name, pays_base_fee`).WithArgs(roleID).
		WillReturnRows(sqlmock.NewRows(roleCols).AddRow(roleID, clubID, "Präsident", true, now, now))
	mock.ExpectQuery(`FROM players`).WithArgs(userID, clubID).
		WillReturnRows(sqlmock.NewRows(playerCols).AddRow(uuid.NewString(), clubID, userID, roleID, "Ich", 0, 0, nil, false, nil, now, now))

	resp := doRequest(t, srv, createPlayerRequest(t, token, clubID, `{"name":"Hans","role_id":"`+roleID+`","link_me":true}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
