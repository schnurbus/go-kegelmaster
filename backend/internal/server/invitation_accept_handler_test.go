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

var invitationCols = []string{"id", "player_id", "email", "token", "expires_at", "accepted_at", "created_at"}

func TestAcceptInvitation_ReturnsClubID(t *testing.T) {
	srv, mock, invMock := newRegistrationModeServer(t, "open")
	authSvc := srv.handlers.AuthSvc
	userID, clubID, roleID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	invID, playerID, token := uuid.NewString(), uuid.NewString(), "seed-token-123"
	now := time.Now().UTC()
	jwt, _ := authSvc.GenerateToken(userID, true)

	mock.ExpectQuery(`SELECT id, email, password_hash`).WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(userColumns).AddRow(userID, "neu@example.com", "hash", now, now, nil, false, false, nil, nil, nil, nil, nil))

	invRow := func() *sqlmock.Rows {
		return sqlmock.NewRows(invitationCols).AddRow(invID, playerID, "neu@example.com", token, now.Add(time.Hour), sql.NullTime{}, now)
	}
	invMock.ExpectQuery(`FROM player_invitations`).WithArgs(token).WillReturnRows(invRow())
	invMock.ExpectBegin()
	invMock.ExpectQuery(`FROM player_invitations`).WithArgs(invID).WillReturnRows(invRow())
	invMock.ExpectQuery(`FROM players`).WithArgs(playerID).
		WillReturnRows(sqlmock.NewRows(playerCols).AddRow(playerID, clubID, nil, roleID, "Neu", 0, 0, nil, false, nil, now, now))
	invMock.ExpectQuery(`FROM players`).WithArgs(userID, clubID).WillReturnError(sql.ErrNoRows)
	invMock.ExpectExec(`UPDATE player_invitations`).WillReturnResult(sqlmock.NewResult(0, 1))
	invMock.ExpectExec(`UPDATE players`).WillReturnResult(sqlmock.NewResult(0, 1))
	invMock.ExpectCommit()

	req := mustJSONRequest(t, http.MethodPost, "/api/invitations/"+token+"/accept", "")
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: jwt})
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body["club_id"] != clubID {
		t.Fatalf("expected club_id %s, got %v", clubID, body)
	}
	if err := invMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
