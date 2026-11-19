package server

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/auth"
	"github.com/schnurbus/go-kegelmaster/backend/internal/handlers"
)

func expectUserWithHash(mock sqlmock.Sqlmock, userID, hash string, created time.Time) {
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(userColumns).
			AddRow(userID, "me@example.com", hash, created, created, nil, false, false, nil, nil, nil))
}

func changePasswordRequest(t *testing.T, token, body string) *http.Request {
	t.Helper()
	req := mustJSONRequest(t, http.MethodPost, "/api/auth/change-password", body)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	}
	return req
}

func hashAndToken(t *testing.T, authSvc *auth.Service, userID, password string, remember bool) (string, string) {
	t.Helper()
	hash, err := authSvc.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	token, err := authSvc.GenerateToken(userID, remember)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return hash, token
}

func TestChangePassword_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	hash, token := hashAndToken(t, authSvc, userID, "old-password", true)
	created := time.Now().UTC().Add(-48 * time.Hour)

	expectUserWithHash(mock, userID, hash, created)
	mock.ExpectExec(`UPDATE users`).
		WithArgs(userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM api_tokens`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	resp := doRequest(t, srv, changePasswordRequest(t, token, `{"current_password":"old-password","new_password":"new-password-123"}`))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var payload handlers.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.PasswordChangedAt == nil {
		t.Fatalf("expected password_changed_at to be set")
	}

	cookie := findCookie(resp.Cookies(), authCookieName)
	if cookie == nil || cookie.Value == "" {
		t.Fatalf("expected auth cookie to be reissued")
	}
	if cookie.Expires.IsZero() {
		t.Fatalf("expected remember-me cookie to keep its expiry")
	}
	claims, err := authSvc.ParseToken(cookie.Value)
	if err != nil {
		t.Fatalf("parse reissued token: %v", err)
	}
	if claims.UserID != userID {
		t.Fatalf("expected token for %s, got %s", userID, claims.UserID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestChangePassword_KeepsSessionCookie(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	hash, token := hashAndToken(t, authSvc, userID, "old-password", false)

	expectUserWithHash(mock, userID, hash, time.Now().UTC())
	mock.ExpectExec(`UPDATE users`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM api_tokens`).WillReturnResult(sqlmock.NewResult(0, 0))

	resp := doRequest(t, srv, changePasswordRequest(t, token, `{"current_password":"old-password","new_password":"new-password-123"}`))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	cookie := findCookie(resp.Cookies(), authCookieName)
	if cookie == nil || !cookie.Expires.IsZero() {
		t.Fatalf("expected a session cookie without expiry, got %+v", cookie)
	}
}

func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	hash, token := hashAndToken(t, authSvc, userID, "old-password", true)

	expectUserWithHash(mock, userID, hash, time.Now().UTC())

	resp := doRequest(t, srv, changePasswordRequest(t, token, `{"current_password":"wrong-password","new_password":"new-password-123"}`))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if findCookie(resp.Cookies(), authCookieName) != nil {
		t.Fatalf("expected no auth cookie on failure")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestChangePassword_Validation(t *testing.T) {
	cases := map[string]string{
		"short new password": `{"current_password":"old-password","new_password":"short"}`,
		"missing current":    `{"current_password":"","new_password":"new-password-123"}`,
		"same password":      `{"current_password":"old-password","new_password":"old-password"}`,
		"invalid json":       `{`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			srv, mock, authSvc := newTestServer(t)
			_, token := hashAndToken(t, authSvc, uuid.NewString(), "old-password", true)

			resp := doRequest(t, srv, changePasswordRequest(t, token, body))
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

func TestChangePassword_Unauthenticated(t *testing.T) {
	srv, _, _ := newTestServer(t)

	resp := doRequest(t, srv, changePasswordRequest(t, "", `{"current_password":"old-password","new_password":"new-password-123"}`))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestChangePassword_MissingCSRF(t *testing.T) {
	srv, _, authSvc := newTestServer(t)
	_, token := hashAndToken(t, authSvc, uuid.NewString(), "old-password", true)

	req, _ := http.NewRequest(http.MethodPost, "/api/auth/change-password", nil)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

func TestChangePassword_RejectsAPIToken(t *testing.T) {
	srv, _, _ := newTestServer(t)

	req := mustJSONRequest(t, http.MethodPost, "/api/auth/change-password", `{"current_password":"old-password","new_password":"new-password-123"}`)
	req.Header.Set("Authorization", "Bearer kgm_whatever")
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}
