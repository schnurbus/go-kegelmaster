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

func TestHandleCreateRole_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	roleID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	// Mock user lookup from cookie
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	// Mock club lookup for owner check
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 0, 0, 0, true, false, userID, now, now))

	// Mock role creation
	mock.ExpectQuery(`INSERT INTO roles`).
		WithArgs(sqlmock.AnyArg(), clubID, "Test Role", false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "pays_base_fee", "created_at", "updated_at"}).
			AddRow(roleID, clubID, "Test Role", false, now, now))

	body := `{"name":"Test Role","pays_base_fee":false}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/roles", body)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if want, got := "Test Role", payload["name"]; got != want {
		t.Fatalf("expected name %s, got %v", want, got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleCreateRole_NotOwner(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	otherUserID := uuid.NewString()
	clubID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	// Mock user lookup from cookie
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	// Mock club lookup - user is not owner (HasPermission -> IsClubOwner uses GetClubByID)
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 0, 0, 0, true, false, otherUserID, now, now))

	// HasPermission then checks player; no player in club -> no permission
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, partner_id, created_at, updated_at FROM players`).
		WithArgs(userID, clubID).
		WillReturnError(sql.ErrNoRows)

	body := `{"name":"Test Role","pays_base_fee":false}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/roles", body)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleGetRoles_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	roleID1 := uuid.NewString()
	roleID2 := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	// Mock user lookup from cookie
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	// Mock club lookup for owner check
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 0, 0, 0, true, false, userID, now, now))

	// Mock roles query
	mock.ExpectQuery(`SELECT id, club_id, name, pays_base_fee, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "pays_base_fee", "created_at", "updated_at"}).
			AddRow(roleID1, clubID, "Role 1", false, now, now).
			AddRow(roleID2, clubID, "Role 2", true, now, now))

	// Mock player count per role for club
	mock.ExpectQuery(`SELECT role_id, COUNT`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"role_id", "count"}).
			AddRow(roleID1, 2).
			AddRow(roleID2, 0))

	// Mock permissions queries for each role
	mock.ExpectQuery(`SELECT id, role_id, entity_type, permission_type, created_at`).
		WithArgs(roleID1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role_id", "entity_type", "permission_type", "created_at"}))

	mock.ExpectQuery(`SELECT id, role_id, entity_type, permission_type, created_at`).
		WithArgs(roleID2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role_id", "entity_type", "permission_type", "created_at"}))

	req := mustJSONRequest(t, http.MethodGet, "/api/clubs/"+clubID+"/roles", "")
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var payload []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(payload) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(payload))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleAddPermission_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	roleID := uuid.NewString()
	permID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	// Mock user lookup from cookie
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	// Mock club lookup for owner check
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 0, 0, 0, true, false, userID, now, now))

	// Mock role lookup
	mock.ExpectQuery(`SELECT id, club_id, name, pays_base_fee, created_at, updated_at`).
		WithArgs(roleID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "pays_base_fee", "created_at", "updated_at"}).
			AddRow(roleID, clubID, "Test Role", false, now, now))

	// Mock permission creation
	mock.ExpectQuery(`INSERT INTO role_permissions`).
		WithArgs(sqlmock.AnyArg(), roleID, "roles", "create", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "role_id", "entity_type", "permission_type", "created_at"}).
			AddRow(permID, roleID, "roles", "create", now))

	body := `{"entity_type":"roles","permission_type":"create"}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/roles/"+roleID+"/permissions", body)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if want, got := "roles", payload["entity_type"]; got != want {
		t.Fatalf("expected entity_type %s, got %v", want, got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleAddPermission_InvalidEntityType(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	roleID := uuid.NewString()
	now := time.Now().UTC()

	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	// Mock user lookup from cookie
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "user@example.com", "hash", now, now))

	// Mock club lookup for owner check
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 0, 0, 0, true, false, userID, now, now))

	// Mock role lookup
	mock.ExpectQuery(`SELECT id, club_id, name, pays_base_fee, created_at, updated_at`).
		WithArgs(roleID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "name", "pays_base_fee", "created_at", "updated_at"}).
			AddRow(roleID, clubID, "Test Role", false, now, now))

	body := `{"entity_type":"invalid","permission_type":"create"}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/roles/"+roleID+"/permissions", body)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
