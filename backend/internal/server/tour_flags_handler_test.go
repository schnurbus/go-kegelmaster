package server

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func meRequest(t *testing.T, token, method, body string) *http.Request {
	t.Helper()
	var req *http.Request
	if method == http.MethodGet {
		req, _ = http.NewRequest(http.MethodGet, "/api/auth/me", nil)
	} else {
		req = mustJSONRequest(t, method, "/api/auth/me", body)
	}
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	return req
}

func decodeMap(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return m
}

func TestCurrentUser_ReturnsTourFlags(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	now := time.Now().UTC()
	seen := now.Add(-24 * time.Hour)
	token, _ := authSvc.GenerateToken(userID, true)

	mock.ExpectQuery(`SELECT id, email, password_hash`).WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(userColumns).
			AddRow(userID, "me@example.com", "hash", now, now, nil, false, false, nil, nil, nil, seen, nil))

	resp := doRequest(t, srv, meRequest(t, token, http.MethodGet, ""))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeMap(t, resp)
	if body["tour_seen_at"] == nil {
		t.Fatalf("expected tour_seen_at to be set, got %v", body)
	}
	if v, ok := body["setup_hint_seen_at"]; !ok || v != nil {
		t.Fatalf("expected setup_hint_seen_at null, got %v (present=%v)", v, ok)
	}
}

func TestUpdateCurrentUser_MarksTourSeen(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	mock.ExpectQuery(`SELECT id, email, password_hash`).WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(userColumns).
			AddRow(userID, "me@example.com", "hash", now, now, nil, false, false, nil, nil, nil, nil, nil))
	mock.ExpectQuery(`UPDATE users`).WithArgs(true, sqlmock.AnyArg(), false, userID).
		WillReturnRows(sqlmock.NewRows(userColumns).
			AddRow(userID, "me@example.com", "hash", now, now, nil, false, false, nil, nil, nil, now, nil))

	resp := doRequest(t, srv, meRequest(t, token, http.MethodPatch, `{"tour_seen":true}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeMap(t, resp)
	if body["tour_seen_at"] == nil {
		t.Fatalf("expected tour_seen_at to be set, got %v", body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateCurrentUser_MarksSetupHintSeen(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	now := time.Now().UTC()
	token, _ := authSvc.GenerateToken(userID, true)

	mock.ExpectQuery(`SELECT id, email, password_hash`).WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(userColumns).
			AddRow(userID, "me@example.com", "hash", now, now, nil, false, false, nil, nil, nil, now, nil))
	mock.ExpectQuery(`UPDATE users`).WithArgs(false, sqlmock.AnyArg(), true, userID).
		WillReturnRows(sqlmock.NewRows(userColumns).
			AddRow(userID, "me@example.com", "hash", now, now, nil, false, false, nil, nil, nil, now, now))

	resp := doRequest(t, srv, meRequest(t, token, http.MethodPatch, `{"setup_hint_seen":true}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if body := decodeMap(t, resp); body["setup_hint_seen_at"] == nil {
		t.Fatalf("expected setup_hint_seen_at to be set, got %v", body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
