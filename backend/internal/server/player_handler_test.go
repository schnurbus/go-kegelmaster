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

func TestHandleCreatePlayer_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	playerID := uuid.NewString()
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

	// Mock club lookup for owner check (permission check)
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 0, 0, 0, true, false, userID, now, now))

	// Mock player creation
	mock.ExpectQuery(`INSERT INTO players`).
		WithArgs(sqlmock.AnyArg(), clubID, nil, &roleID, "Test Player", 1000, 500, sqlmock.AnyArg(), false, nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(playerID, clubID, nil, roleID, "Test Player", 1000, 500, nil, false, nil, now, now))

	body := `{"name":"Test Player","balance":1000,"start_balance":500,"role_id":"` + roleID + `"}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/players", body)
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

	if want, got := "Test Player", payload["name"]; got != want {
		t.Fatalf("expected name %s, got %v", want, got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleCreatePlayer_WithUserID(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	playerID := uuid.NewString()
	playerUserID := uuid.NewString()
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

	// Mock player creation
	mock.ExpectQuery(`INSERT INTO players`).
		WithArgs(sqlmock.AnyArg(), clubID, &playerUserID, &roleID, "Test Player", 1000, 500, sqlmock.AnyArg(), false, nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(playerID, clubID, playerUserID, roleID, "Test Player", 1000, 500, nil, false, nil, now, now))

	body := `{"name":"Test Player","balance":1000,"start_balance":500,"user_id":"` + playerUserID + `","role_id":"` + roleID + `"}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/players", body)
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

	if want, got := "Test Player", payload["name"]; got != want {
		t.Fatalf("expected name %s, got %v", want, got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleCreatePlayer_NotOwnerNoPermission(t *testing.T) {
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

	// Mock club lookup - user is not owner
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 0, 0, 0, true, false, otherUserID, now, now))

	// Permission check: user has no player in this club
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, partner_id, created_at, updated_at FROM players`).
		WithArgs(userID, clubID).
		WillReturnError(sql.ErrNoRows)

	body := `{"name":"Test Player","balance":1000,"start_balance":500}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/players", body)
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

func TestHandleCreatePlayer_MissingName(t *testing.T) {
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

	body := `{"name":"","balance":1000,"start_balance":500,"role_id":"` + roleID + `"}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/players", body)
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

func TestHandleCreatePlayer_MissingRole(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
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

	// Mock club lookup for owner check
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 0, 0, 0, true, false, userID, now, now))

	body := `{"name":"Test Player","balance":1000,"start_balance":500}`
	req := mustJSONRequest(t, http.MethodPost, "/api/clubs/"+clubID+"/players", body)
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

func TestHandleGetPlayers_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	playerID1 := uuid.NewString()
	playerID2 := uuid.NewString()
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

	// Mock players query
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, partner_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(playerID1, clubID, nil, nil, "Player 1", 1000, 500, nil, false, nil, now, now).
			AddRow(playerID2, clubID, nil, nil, "Player 2", 2000, 1000, nil, false, nil, now, now))

	req := mustJSONRequest(t, http.MethodGet, "/api/clubs/"+clubID+"/players", "")
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
		t.Fatalf("expected 2 players, got %d", len(payload))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleGetPlayers_NotOwnerNoPermission(t *testing.T) {
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

	// Mock club lookup - user is not owner
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 0, 0, 0, true, false, otherUserID, now, now))

	// Permission check: user has no player in this club
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, partner_id, created_at, updated_at FROM players`).
		WithArgs(userID, clubID).
		WillReturnError(sql.ErrNoRows)

	req := mustJSONRequest(t, http.MethodGet, "/api/clubs/"+clubID+"/players", "")
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

func TestHandleGetPlayer_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	playerID := uuid.NewString()
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

	// Mock player query (handler loads player first)
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, partner_id, created_at, updated_at`).
		WithArgs(playerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(playerID, clubID, nil, nil, "Test Player", 1000, 500, nil, false, nil, now, now))

	// Permission check: club lookup (user is owner, so has permission)
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 0, 0, 0, true, false, userID, now, now))

	req := mustJSONRequest(t, http.MethodGet, "/api/clubs/"+clubID+"/players/"+playerID, "")
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if want, got := "Test Player", payload["name"]; got != want {
		t.Fatalf("expected name %s, got %v", want, got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleGetPlayer_NotOwnerNoPermission(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	otherUserID := uuid.NewString()
	clubID := uuid.NewString()
	playerID := uuid.NewString()
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

	// Mock player query (handler loads player first)
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, partner_id, created_at, updated_at`).
		WithArgs(playerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(playerID, clubID, nil, nil, "Other Player", 1000, 500, nil, false, nil, now, now))

	// Permission check: club lookup (user is not owner)
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 0, 0, 0, true, false, otherUserID, now, now))

	// Permission check: user has no player in this club
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, partner_id, created_at, updated_at FROM players`).
		WithArgs(userID, clubID).
		WillReturnError(sql.ErrNoRows)

	req := mustJSONRequest(t, http.MethodGet, "/api/clubs/"+clubID+"/players/"+playerID, "")
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

func TestHandleGetPlayer_WrongClub(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	wrongClubID := uuid.NewString()
	playerID := uuid.NewString()
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

	// Mock player query - player belongs to different club (handler returns 404 before permission check)
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, partner_id, created_at, updated_at`).
		WithArgs(playerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(playerID, wrongClubID, nil, nil, "Test Player", 1000, 500, nil, false, nil, now, now))

	req := mustJSONRequest(t, http.MethodGet, "/api/clubs/"+clubID+"/players/"+playerID, "")
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleUpdatePlayer_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	playerID := uuid.NewString()
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

	// Mock player query to verify existence and club membership
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, partner_id, created_at, updated_at`).
		WithArgs(playerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(playerID, clubID, nil, roleID, "Old Name", 1000, 500, nil, false, nil, now, now))

	// Mock player update
	mock.ExpectQuery(`UPDATE players`).
		WithArgs("Updated Player", 2000, 1000, nil, &roleID, sqlmock.AnyArg(), false, nil, sqlmock.AnyArg(), playerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(playerID, clubID, nil, roleID, "Updated Player", 2000, 1000, nil, false, nil, now, now.Add(time.Hour)))

	body := `{"name":"Updated Player","balance":2000,"start_balance":1000,"role_id":"` + roleID + `"}`
	req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID+"/players/"+playerID, body)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if want, got := "Updated Player", payload["name"]; got != want {
		t.Fatalf("expected name %s, got %v", want, got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleUpdatePlayer_MissingRole(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	playerID := uuid.NewString()
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

	// Mock player query to verify existence and club membership
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, partner_id, created_at, updated_at`).
		WithArgs(playerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(playerID, clubID, nil, roleID, "Test Player", 1000, 500, nil, false, nil, now, now))

	body := `{"name":"Updated Player","balance":2000,"start_balance":1000}`
	req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID+"/players/"+playerID, body)
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

func TestHandleUpdatePlayer_NotOwnerNoPermission(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	otherUserID := uuid.NewString()
	clubID := uuid.NewString()
	playerID := uuid.NewString()
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

	// Mock club lookup - user is not owner
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 0, 0, 0, true, false, otherUserID, now, now))

	// Permission check: user has no player in this club
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, partner_id, created_at, updated_at FROM players`).
		WithArgs(userID, clubID).
		WillReturnError(sql.ErrNoRows)

	body := `{"name":"Updated Player","balance":2000,"start_balance":1000}`
	req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID+"/players/"+playerID, body)
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

func TestHandleUpdatePlayer_WrongClub(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	wrongClubID := uuid.NewString()
	playerID := uuid.NewString()
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

	// Mock player query - player belongs to different club
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, partner_id, created_at, updated_at`).
		WithArgs(playerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(playerID, wrongClubID, nil, nil, "Test Player", 1000, 500, nil, false, nil, now, now))

	body := `{"name":"Updated Player","balance":2000,"start_balance":1000}`
	req := mustJSONRequest(t, http.MethodPut, "/api/clubs/"+clubID+"/players/"+playerID, body)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleDeletePlayer_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	playerID := uuid.NewString()
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

	// Mock player query to verify existence and club membership
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, partner_id, created_at, updated_at`).
		WithArgs(playerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(playerID, clubID, nil, nil, "Test Player", 1000, 500, nil, false, nil, now, now))

	// Additional GetByID call in Delete method
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, partner_id, created_at, updated_at FROM players`).
		WithArgs(playerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(playerID, clubID, nil, nil, "Test Player", 1000, 500, nil, false, nil, now, now))

	// Mock player deletion
	mock.ExpectExec(`DELETE FROM players`).
		WithArgs(playerID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := mustJSONRequest(t, http.MethodDelete, "/api/clubs/"+clubID+"/players/"+playerID, "")
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleDeletePlayer_NotOwnerNoPermission(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	otherUserID := uuid.NewString()
	clubID := uuid.NewString()
	playerID := uuid.NewString()
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

	// Mock club lookup - user is not owner
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}).
			AddRow(clubID, "Test Club", 0, 0, 0, true, false, otherUserID, now, now))

	// Permission check: user has no player in this club
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, partner_id, created_at, updated_at FROM players`).
		WithArgs(userID, clubID).
		WillReturnError(sql.ErrNoRows)

	req := mustJSONRequest(t, http.MethodDelete, "/api/clubs/"+clubID+"/players/"+playerID, "")
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

func TestHandleDeletePlayer_WrongClub(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)

	userID := uuid.NewString()
	clubID := uuid.NewString()
	wrongClubID := uuid.NewString()
	playerID := uuid.NewString()
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

	// Mock player query - player belongs to different club
	mock.ExpectQuery(`SELECT id, club_id, user_id, role_id, name, balance, start_balance, gender, inactive, partner_id, created_at, updated_at`).
		WithArgs(playerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "club_id", "user_id", "role_id", "name", "balance", "start_balance", "gender", "inactive", "partner_id", "created_at", "updated_at"}).
			AddRow(playerID, wrongClubID, nil, nil, "Test Player", 1000, 500, nil, false, nil, now, now))

	req := mustJSONRequest(t, http.MethodDelete, "/api/clubs/"+clubID+"/players/"+playerID, "")
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	resp := doRequest(t, srv, req)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

