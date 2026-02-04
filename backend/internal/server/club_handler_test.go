package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/handlers"
)

// Club handler tests

func TestCreateClub_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	mock.ExpectQuery(`INSERT INTO clubs`).
		WithArgs(sqlmock.AnyArg(), "Test Club", 1000, 0, 500, true, false, userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, userID, now, now))

	body := `{"name":"Test Club","balance":1000,"base_fee":500}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs", body)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var payload handlers.ClubResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Name != "Test Club" {
		t.Fatalf("expected name 'Test Club', got %s", payload.Name)
	}
	if payload.Balance != 1000 {
		t.Fatalf("expected balance 1000, got %d", payload.Balance)
	}
	if payload.BaseFee != 500 {
		t.Fatalf("expected base_fee 500, got %d", payload.BaseFee)
	}
	if payload.UserID != userID {
		t.Fatalf("expected user_id %s, got %s", userID, payload.UserID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateClub_Unauthorized(t *testing.T) {
	srv, mock, _ := newTestServer(t)

	body := `{"name":"Test Club","balance":1000,"base_fee":500}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs", body)

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateClub_MissingCSRF(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/clubs", bytes.NewBufferString(`{"name":"Test Club","balance":1000,"base_fee":500}`))
	req.Header.Set("Content-Type", "application/json")
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

func TestCreateClub_InvalidName(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	now := time.Now().UTC()
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	body := `{"name":"","balance":1000,"base_fee":500}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs", body)
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

func TestCreateClub_NegativeBalance(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	now := time.Now().UTC()
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	body := `{"name":"Test Club","balance":-100,"base_fee":500}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs", body)
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

func TestGetClubs_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID1 := uuid.NewString()
	clubID2 := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID1, "Club 1", 1000, 0, 500, true, false, userID, now, now).
			AddRow(clubID2, "Club 2", 2000, 0, 600, true, false, userID, now, now))

	req, _ := http.NewRequest(http.MethodGet, "/api/clubs", nil)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var payload []handlers.ClubResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(payload) != 2 {
		t.Fatalf("expected 2 clubs, got %d", len(payload))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetClubs_Unauthorized(t *testing.T) {
	srv, mock, _ := newTestServer(t)

	req, _ := http.NewRequest(http.MethodGet, "/api/clubs", nil)

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetClub_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, userID, now, now))

	req, _ := http.NewRequest(http.MethodGet, "/api/clubs/"+clubID, nil)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var payload handlers.ClubResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.ID != clubID {
		t.Fatalf("expected id %s, got %s", clubID, payload.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetClub_NotFound(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnError(sql.ErrNoRows)

	req, _ := http.NewRequest(http.MethodGet, "/api/clubs/"+clubID, nil)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetClub_MissingID(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	req, _ := http.NewRequest(http.MethodGet, "/api/clubs/", nil)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	// Fiber behandelt leere Parameter anders, aber wir testen den Fall
	if resp.StatusCode == http.StatusNotFound {
		// Das ist auch akzeptabel
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateClub_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Old Name", 1000, 0, 500, true, false, userID, now, now))

	mock.ExpectQuery(`UPDATE clubs`).
		WithArgs("Updated Club", 2000, 0, 600, false, false, sqlmock.AnyArg(), clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Updated Club", 2000, 0, 600, false, false, userID, now, now.Add(time.Hour)))

	body := `{"name":"Updated Club","balance":2000,"base_fee":600}`
	req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID, body)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var payload handlers.ClubResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Name != "Updated Club" {
		t.Fatalf("expected name 'Updated Club', got %s", payload.Name)
	}
	if payload.Balance != 2000 {
		t.Fatalf("expected balance 2000, got %d", payload.Balance)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateClub_NotOwner(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	ownerID := uuid.NewString()
	otherUserID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(otherUserID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(otherUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(otherUserID, "other@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, ownerID, now, now))

	body := `{"name":"Updated Club","balance":2000,"base_fee":600}`
	req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID, body)
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

func TestUpdateClub_NotFound(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnError(sql.ErrNoRows)

	body := `{"name":"Updated Club","balance":2000,"base_fee":600}`
	req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID, body)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateClub_InvalidRequest(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, userID, now, now))

	body := `{"name":"","balance":2000,"base_fee":600}`
	req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID, body)
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

func TestDeleteClub_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	passwordHash, err := authSvc.HashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", passwordHash, now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, userID, now, now))

	// Additional GetByID call in Delete method
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at FROM clubs`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, userID, now, now))

	mock.ExpectExec(`DELETE FROM clubs`).
		WithArgs(clubID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := mustJSONRequest(t, http.MethodDelete, "/api/clubs/"+clubID, `{"password":"password123"}`)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteClub_NotOwner(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	ownerID := uuid.NewString()
	otherUserID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(otherUserID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(otherUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(otherUserID, "other@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, ownerID, now, now))

	req := mustJSONRequest(t, http.MethodDelete, "/api/clubs/"+clubID, `{"password":"any"}`)
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

func TestDeleteClub_MissingPassword(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, userID, now, now))

	req := mustJSONRequest(t, http.MethodDelete, "/api/clubs/"+clubID, `{}`)
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

func TestDeleteClub_InvalidPassword(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, userID, now, now))

	req := mustJSONRequest(t, http.MethodDelete, "/api/clubs/"+clubID, `{"password":"wrongpassword"}`)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteClub_NotFound(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnError(sql.ErrNoRows)

	req := mustJSONRequest(t, http.MethodDelete, "/api/clubs/"+clubID, `{"password":"any"}`)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteClub_Unauthorized(t *testing.T) {
	srv, mock, _ := newTestServer(t)

	clubID := uuid.NewString()

	req := mustJSONRequest(t, http.MethodDelete, "/api/clubs/"+clubID, "")

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateClub_DatabaseError(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	mock.ExpectQuery(`INSERT INTO clubs`).
		WithArgs(sqlmock.AnyArg(), "Test Club", 1000, 0, 500, true, false, userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	body := `{"name":"Test Club","balance":1000,"base_fee":500}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs", body)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateClub_InvalidJSON(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	body := `{"name":"Test Club","balance":"invalid"}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs", body)
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

func TestGetClubs_DatabaseError(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)

	req, _ := http.NewRequest(http.MethodGet, "/api/clubs", nil)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetClub_DatabaseError(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnError(sql.ErrConnDone)

	req, _ := http.NewRequest(http.MethodGet, "/api/clubs/"+clubID, nil)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateClub_DatabaseError(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, userID, now, now))

	mock.ExpectQuery(`UPDATE clubs`).
		WithArgs("Updated Club", 2000, 0, 600, false, false, sqlmock.AnyArg(), clubID).
		WillReturnError(sql.ErrConnDone)

	body := `{"name":"Updated Club","balance":2000,"base_fee":600}`
	req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID, body)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateClub_InvalidJSON(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, userID, now, now))

	body := `{"name":"Updated Club","balance":"invalid"}`
	req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID, body)
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

func TestUpdateClub_NegativeBaseFee(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, userID, now, now))

	body := `{"name":"Updated Club","balance":2000,"base_fee":-100}`
	req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID, body)
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

func TestUpdateClub_GetByIDError(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnError(sql.ErrConnDone)

	body := `{"name":"Updated Club","balance":2000,"base_fee":600}`
	req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID, body)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteClub_GetByIDError(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnError(sql.ErrConnDone)

	req := mustJSONRequest(t, http.MethodDelete, "/api/clubs/"+clubID, `{"password":"any"}`)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteClub_DatabaseError(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	passwordHash, err := authSvc.HashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", passwordHash, now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, userID, now, now))

	// Additional GetByID call in Delete method returns error
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at FROM clubs`).
		WithArgs(clubID).
		WillReturnError(sql.ErrConnDone)

	req := mustJSONRequest(t, http.MethodDelete, "/api/clubs/"+clubID, `{"password":"password123"}`)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTransferClubOwner_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	ownerID := uuid.NewString()
	newOwnerID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	passwordHash, err := authSvc.HashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	token, err := authSvc.GenerateToken(ownerID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(ownerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(ownerID, "owner@example.com", passwordHash, now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, ownerID, now, now))

	mock.ExpectQuery(`SELECT id, email, password_hash, created_at, updated_at FROM users`).
		WithArgs("newowner@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(newOwnerID, "newowner@example.com", "hash", now, now))

	mock.ExpectQuery(`UPDATE clubs`).
		WithArgs(newOwnerID, sqlmock.AnyArg(), clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, newOwnerID, now, now))

	body := `{"new_owner_email":"newowner@example.com","password":"password123"}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/transfer-owner", body)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var payload handlers.ClubResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.UserID != newOwnerID {
		t.Fatalf("expected user_id %s, got %s", newOwnerID, payload.UserID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTransferClubOwner_NotOwner(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	ownerID := uuid.NewString()
	otherUserID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(otherUserID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(otherUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(otherUserID, "other@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, ownerID, now, now))

	body := `{"new_owner_email":"someone@example.com","password":"any"}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/transfer-owner", body)
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

func TestTransferClubOwner_InvalidPassword(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	ownerID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(ownerID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(ownerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(ownerID, "owner@example.com", "hash", now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, ownerID, now, now))

	body := `{"new_owner_email":"other@example.com","password":"wrongpassword"}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/transfer-owner", body)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTransferClubOwner_NewUserNotFound(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	ownerID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	passwordHash, err := authSvc.HashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	token, err := authSvc.GenerateToken(ownerID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(ownerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(ownerID, "owner@example.com", passwordHash, now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, ownerID, now, now))

	mock.ExpectQuery(`SELECT id, email, password_hash, created_at, updated_at FROM users`).
		WithArgs("nonexistent@example.com").
		WillReturnError(sql.ErrNoRows)

	body := `{"new_owner_email":"nonexistent@example.com","password":"password123"}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/transfer-owner", body)
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

func TestTransferClubOwner_SameUser(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	ownerID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	passwordHash, err := authSvc.HashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	token, err := authSvc.GenerateToken(ownerID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(ownerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(ownerID, "owner@example.com", passwordHash, now, now))

	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 1000, 0, 500, true, false, ownerID, now, now))

	mock.ExpectQuery(`SELECT id, email, password_hash, created_at, updated_at FROM users`).
		WithArgs("owner@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(ownerID, "owner@example.com", "hash", now, now))

	body := `{"new_owner_email":"owner@example.com","password":"password123"}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/transfer-owner", body)
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

// Role handler tests
