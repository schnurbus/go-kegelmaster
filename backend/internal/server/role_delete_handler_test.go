package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

var roleCols = []string{"id", "club_id", "name", "pays_base_fee", "created_at", "updated_at"}

func deleteRoleRequest(t *testing.T, token, clubID, roleID string) *http.Request {
	t.Helper()
	req := mustJSONRequest(t, http.MethodDelete, "/api/clubs/"+clubID+"/roles/"+roleID, "")
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	return req
}

func TestHandleDeleteRole_WithPlayersConflict(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, clubID, roleID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	expectOwner(mock, userID, clubID, now)
	mock.ExpectQuery(`SELECT .* FROM roles`).WithArgs(roleID).
		WillReturnRows(sqlmock.NewRows(roleCols).AddRow(roleID, clubID, "Mitglied", true, now, now))
	mock.ExpectQuery(`SELECT role_id, COUNT\(\*\)`).WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"role_id", "count"}).AddRow(roleID, 3))

	resp := doRequest(t, srv, deleteRoleRequest(t, token, clubID, roleID))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	msg, _ := body["message"].(string)
	if !strings.Contains(msg, "3 Spieler") {
		t.Fatalf("expected message to mention 3 Spieler, got %q", msg)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleDeleteRole_EmptyRoleSucceeds(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, clubID, roleID, otherRoleID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	expectOwner(mock, userID, clubID, now)
	mock.ExpectQuery(`SELECT .* FROM roles`).WithArgs(roleID).
		WillReturnRows(sqlmock.NewRows(roleCols).AddRow(roleID, clubID, "Gast", false, now, now))
	mock.ExpectQuery(`SELECT role_id, COUNT\(\*\)`).WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"role_id", "count"}).AddRow(otherRoleID, 5))
	mock.ExpectQuery(`SELECT .* FROM roles`).WithArgs(roleID).
		WillReturnRows(sqlmock.NewRows(roleCols).AddRow(roleID, clubID, "Gast", false, now, now))
	mock.ExpectExec(`DELETE FROM roles`).WithArgs(roleID).WillReturnResult(sqlmock.NewResult(0, 1))

	resp := doRequest(t, srv, deleteRoleRequest(t, token, clubID, roleID))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
