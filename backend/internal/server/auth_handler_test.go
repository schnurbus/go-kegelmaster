package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/schnurbus/go-kegelmaster/backend/internal/handlers"
)

func TestRegister_Success(t *testing.T) {
	srv, mock, _ := newTestServer(t)

	email := "User@Example.com"
	now := time.Now().UTC()
	userID := uuid.NewString()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), strings.ToLower(email), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, strings.ToLower(email), "hash", now, now))

	body := `{"email":"` + email + `","password":"super-safe-123"}`
	req := mustJSONRequest(t, http.MethodPost, "/api/auth/register", body)
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()
	csrfCookie := findCookie(resp.Cookies(), csrfCookieName)
	if csrfCookie == nil || csrfCookie.Value == "" {
		t.Fatalf("expected csrf cookie to be set")
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if want, got := strings.ToLower(email), payload["email"]; got != want {
		t.Fatalf("expected email %s, got %v", want, got)
	}

	cookie := findCookie(resp.Cookies(), "auth_token")
	if cookie == nil || cookie.Value == "" {
		t.Fatalf("expected auth cookie to be set")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	srv, mock, _ := newTestServer(t)

	req := mustJSONRequest(t, http.MethodPost, "/api/auth/register", `{"email":"invalid","password":"123456789"}`)
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	srv, mock, _ := newTestServer(t)

	req := mustJSONRequest(t, http.MethodPost, "/api/auth/register", `{"email":"test@example.com","password":"123"}`)
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRegister_EmailConflict(t *testing.T) {
	srv, mock, _ := newTestServer(t)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), "conflict@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(&pgconn.PgError{Code: "23505"})

	req := mustJSONRequest(t, http.MethodPost, "/api/auth/register", `{"email":"conflict@example.com","password":"pass1234"}`)
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRegister_MissingCSRF(t *testing.T) {
	srv, mock, _ := newTestServer(t)

	req, err := http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(`{"email":"a@b.com","password":"password1"}`))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	email := "login@example.com"
	now := time.Now().UTC()
	userID := uuid.NewString()

	hash, err := authSvc.HashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(strings.ToLower(email)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, strings.ToLower(email), hash, now, now))

	body := `{"email":"` + email + `","password":"password123"}`
	req := mustJSONRequest(t, http.MethodPost, "/api/auth/login", body)
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	cookie := findCookie(resp.Cookies(), "auth_token")
	if cookie == nil || cookie.Value == "" {
		t.Fatalf("expected auth cookie to be set")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestLogin_MissingFields(t *testing.T) {
	srv, mock, _ := newTestServer(t)

	req := mustJSONRequest(t, http.MethodPost, "/api/auth/login", `{"email":"","password":""}`)
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	srv, mock, _ := newTestServer(t)

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs("missing@example.com").
		WillReturnError(sql.ErrNoRows)

	req := mustJSONRequest(t, http.MethodPost, "/api/auth/login", `{"email":"missing@example.com","password":"pass1234"}`)
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	email := "wrongpass@example.com"
	now := time.Now().UTC()
	userID := uuid.NewString()

	hash, err := authSvc.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(strings.ToLower(email)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, strings.ToLower(email), hash, now, now))

	req := mustJSONRequest(t, http.MethodPost, "/api/auth/login", `{"email":"`+email+`","password":"wrong"}`)
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestLogout_ClearsCookie(t *testing.T) {
	srv, _, _ := newTestServer(t)

	req, _ := http.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.Header.Set("X-CSRF-Token", testCSRFToken)
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: testCSRFToken})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.StatusCode)
	}

	cookie := findCookie(resp.Cookies(), "auth_token")
	if cookie == nil {
		t.Fatalf("expected logout to send auth cookie")
	}

	if cookie.Value != "" {
		t.Fatalf("expected cookie value to be empty, got %q", cookie.Value)
	}

	if !cookie.Expires.Equal(time.Unix(0, 0)) {
		t.Fatalf("expected cookie expiry to be Unix epoch, got %s", cookie.Expires)
	}
}

func TestCSRFEndpointSetsCookie(t *testing.T) {
	srv, _, _ := newTestServer(t)

	req, _ := http.NewRequest(http.MethodGet, "/api/auth/csrf-token", nil)
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if findCookie(resp.Cookies(), csrfCookieName) == nil {
		t.Fatalf("expected csrf cookie")
	}
}

func TestCurrentUser_UnauthorizedWithoutCookie(t *testing.T) {
	srv, _, _ := newTestServer(t)

	req, _ := http.NewRequest(http.MethodGet, "/api/auth/me", nil)
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestCurrentUser_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	email := "me@example.com"
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, email, "hash", now, now))

	req, _ := http.NewRequest(http.MethodGet, "/api/auth/me", nil)
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

	if payload.ID != userID {
		t.Fatalf("expected id %s got %s", userID, payload.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
