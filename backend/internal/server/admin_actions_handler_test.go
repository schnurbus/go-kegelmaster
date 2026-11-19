package server

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

var clubColumns = []string{"id", "name", "balance", "start_balance", "base_fee", "auto_tip_enabled", "couples_mode_enabled", "user_id", "created_at", "updated_at"}

// expectTargetUser expects the lookup of the account an admin action targets.
func expectTargetUser(mock sqlmock.Sqlmock, userID, email string, isAdmin bool, disabledAt driver.Value) {
	now := time.Now().UTC().Add(-time.Hour)
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(userColumns).
			AddRow(userID, email, "hash", now, now, nil, false, isAdmin, nil, disabledAt, nil))
}

func expectTargetClub(mock sqlmock.Sqlmock, clubID, name, ownerID string) {
	now := time.Now().UTC()
	mock.ExpectQuery(`SELECT id, name, balance, start_balance, base_fee, auto_tip_enabled, couples_mode_enabled, user_id, created_at, updated_at`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows(clubColumns).AddRow(clubID, name, 1000, 0, 500, true, false, ownerID, now, now))
}

// adminMutation builds a CSRF-protected request with the admin's session cookie.
func adminMutation(t *testing.T, token, method, path, body string) *http.Request {
	t.Helper()
	req := mustJSONRequest(t, method, path, body)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	return req
}

func expectStatus(t *testing.T, srv *Server, req *http.Request, want int) *http.Response {
	t.Helper()
	resp := doRequest(t, srv, req)
	if resp.StatusCode != want {
		t.Fatalf("%s %s: expected %d, got %d (%s)", req.Method, req.URL.Path, want, resp.StatusCode, readMessage(t, resp))
	}
	return resp
}

func assertMockMet(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// --- Disabled accounts and revoked sessions --------------------------------

func TestLogin_DisabledUserIsRejected(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	now := time.Now().UTC()
	hash, err := authSvc.HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs("blocked@example.com").
		WillReturnRows(sqlmock.NewRows(userColumns).
			AddRow(userID, "blocked@example.com", hash, now, now, nil, false, false, nil, now, now))
	// No last_login_at update and no cookie for a disabled account.

	req := mustJSONRequest(t, http.MethodPost, "/api/auth/login", `{"email":"blocked@example.com","password":"password123"}`)
	resp := expectStatus(t, srv, req, http.StatusForbidden)
	defer resp.Body.Close()
	if c := findCookie(resp.Cookies(), authCookieName); c != nil && c.Value != "" {
		t.Errorf("disabled user must not get an auth cookie")
	}
	assertMockMet(t, mock)
}

// The "disabled" hint is only given after a correct password.
func TestLogin_DisabledUserWrongPasswordIsGeneric401(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	now := time.Now().UTC()
	hash, _ := authSvc.HashPassword("password123")
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs("blocked@example.com").
		WillReturnRows(sqlmock.NewRows(userColumns).
			AddRow(uuid.NewString(), "blocked@example.com", hash, now, now, nil, false, false, nil, now, now))

	req := mustJSONRequest(t, http.MethodPost, "/api/auth/login", `{"email":"blocked@example.com","password":"wrong-password"}`)
	resp := expectStatus(t, srv, req, http.StatusUnauthorized)
	resp.Body.Close()
	assertMockMet(t, mock)
}

func TestSession_DisabledUserIsRejected(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	token, _ := authSvc.GenerateToken(userID, true)
	expectTargetUser(mock, userID, "blocked@example.com", false, time.Now().UTC())

	req, _ := http.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	resp := expectStatus(t, srv, req, http.StatusUnauthorized)
	resp.Body.Close()
	assertMockMet(t, mock)
}

func TestSession_RevokedByAdminIsRejected(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	token, _ := authSvc.GenerateToken(userID, true)
	now := time.Now().UTC()
	revokedAt := now.Add(time.Minute) // after the token was issued
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(userColumns).
			AddRow(userID, "u@example.com", "hash", now, now, nil, false, false, nil, nil, revokedAt))

	req, _ := http.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	resp := expectStatus(t, srv, req, http.StatusUnauthorized)
	resp.Body.Close()
	assertMockMet(t, mock)
}

func TestBearerToken_DisabledUserIsRejected(t *testing.T) {
	srv, mock, _ := newTestServer(t)
	userID, tokenID := uuid.NewString(), uuid.NewString()
	expectAPITokenLookup(mock, tokenID, userID, time.Now().UTC().Add(-time.Hour), nil)
	expectTargetUser(mock, userID, "agent@example.com", false, time.Now().UTC())
	// The token is not touched (no last_used_at update).

	resp := expectStatus(t, srv, bearerRequest(http.MethodGet, "/api/auth/me", ""), http.StatusUnauthorized)
	resp.Body.Close()
	assertMockMet(t, mock)
}

// --- Access control of the action endpoints -------------------------------

var adminActionEndpoints = []struct{ method, path string }{
	{http.MethodPost, "/api/admin/users/00000000-0000-0000-0000-000000000001/disable"},
	{http.MethodPost, "/api/admin/users/00000000-0000-0000-0000-000000000001/enable"},
	{http.MethodPost, "/api/admin/users/00000000-0000-0000-0000-000000000001/revoke-sessions"},
	{http.MethodDelete, "/api/admin/users/00000000-0000-0000-0000-000000000001/tokens/00000000-0000-0000-0000-000000000003"},
	{http.MethodDelete, "/api/admin/users/00000000-0000-0000-0000-000000000001"},
	{http.MethodPost, "/api/admin/clubs/00000000-0000-0000-0000-000000000002/transfer-owner"},
	{http.MethodDelete, "/api/admin/clubs/00000000-0000-0000-0000-000000000002"},
}

func TestAdminActions_NonAdminGets404(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	token, _ := authSvc.GenerateToken(userID, true)
	for _, e := range adminActionEndpoints {
		expectSessionUser(mock, userID, false)
		resp := expectStatus(t, srv, adminMutation(t, token, e.method, e.path, `{}`), http.StatusNotFound)
		resp.Body.Close()
	}
	assertMockMet(t, mock)
}

func TestAdminActions_BearerTokenGets404(t *testing.T) {
	srv, mock, _ := newTestServer(t)
	for _, e := range adminActionEndpoints {
		req := mustJSONRequest(t, e.method, e.path, `{}`)
		req.Header.Set("Authorization", "Bearer "+testAPIToken)
		resp := expectStatus(t, srv, req, http.StatusNotFound)
		resp.Body.Close()
	}
	assertMockMet(t, mock)
}

func TestAdminActions_DisabledAdminGets404(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	adminID := uuid.NewString()
	token, _ := authSvc.GenerateToken(adminID, true)
	now := time.Now().UTC().Add(-time.Hour)
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(adminID).
		WillReturnRows(sqlmock.NewRows(userColumns).
			AddRow(adminID, "op@example.com", "hash", now, now, nil, false, true, nil, now, nil))

	resp := expectStatus(t, srv, adminRequest(t, token, "/api/admin/stats"), http.StatusNotFound)
	resp.Body.Close()
	assertMockMet(t, mock)
}

func TestAdminActions_RequireCSRF(t *testing.T) {
	for _, e := range adminActionEndpoints {
		srv, mock, _, token := adminSession(t)
		req, _ := http.NewRequest(e.method, e.path, nil)
		req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
		resp := expectStatus(t, srv, req, http.StatusForbidden)
		resp.Body.Close()
		assertMockMet(t, mock)
	}
}

// --- Disable / enable -------------------------------------------------------

func TestAdminDisableUser_Success(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	targetID := uuid.NewString()
	expectTargetUser(mock, targetID, "member@example.com", false, nil)
	mock.ExpectExec(`UPDATE users\s+SET disabled_at = COALESCE\(disabled_at`).
		WithArgs(sqlmock.AnyArg(), targetID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp := expectStatus(t, srv, adminMutation(t, token, http.MethodPost, "/api/admin/users/"+targetID+"/disable", ""), http.StatusNoContent)
	resp.Body.Close()
	assertMockMet(t, mock)
}

func TestAdminDisableUser_Self(t *testing.T) {
	srv, mock, adminID, token := adminSession(t)
	expectTargetUser(mock, adminID, "op@example.com", true, nil)

	resp := expectStatus(t, srv, adminMutation(t, token, http.MethodPost, "/api/admin/users/"+adminID+"/disable", ""), http.StatusBadRequest)
	resp.Body.Close()
	assertMockMet(t, mock)
}

func TestAdminDisableUser_OtherAdmin(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	otherID := uuid.NewString()
	expectTargetUser(mock, otherID, "op2@example.com", true, nil)

	resp := expectStatus(t, srv, adminMutation(t, token, http.MethodPost, "/api/admin/users/"+otherID+"/disable", ""), http.StatusConflict)
	resp.Body.Close()
	assertMockMet(t, mock)
}

func TestAdminDisableUser_NotFound(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	targetID := uuid.NewString()
	mock.ExpectQuery(`SELECT id, email, password_hash`).WithArgs(targetID).WillReturnError(sql.ErrNoRows)

	resp := expectStatus(t, srv, adminMutation(t, token, http.MethodPost, "/api/admin/users/"+targetID+"/disable", ""), http.StatusNotFound)
	resp.Body.Close()
	assertMockMet(t, mock)
}

func TestAdminEnableUser_Success(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	targetID := uuid.NewString()
	expectTargetUser(mock, targetID, "member@example.com", false, time.Now().UTC())
	mock.ExpectExec(`UPDATE users\s+SET disabled_at = NULL`).
		WithArgs(sqlmock.AnyArg(), targetID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp := expectStatus(t, srv, adminMutation(t, token, http.MethodPost, "/api/admin/users/"+targetID+"/enable", ""), http.StatusNoContent)
	resp.Body.Close()
	assertMockMet(t, mock)
}

// --- Sessions and API tokens ------------------------------------------------

func TestAdminRevokeSessions_Success(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	targetID := uuid.NewString()
	expectTargetUser(mock, targetID, "member@example.com", false, nil)
	mock.ExpectExec(`UPDATE users\s+SET sessions_valid_after`).
		WithArgs(sqlmock.AnyArg(), targetID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp := expectStatus(t, srv, adminMutation(t, token, http.MethodPost, "/api/admin/users/"+targetID+"/revoke-sessions", ""), http.StatusNoContent)
	resp.Body.Close()
	assertMockMet(t, mock)
}

func TestAdminRevokeAPIToken(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	targetID, tokenID := uuid.NewString(), uuid.NewString()
	mock.ExpectExec(`DELETE FROM api_tokens\s+WHERE id = \$1 AND user_id = \$2`).
		WithArgs(tokenID, targetID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp := expectStatus(t, srv, adminMutation(t, token, http.MethodDelete, "/api/admin/users/"+targetID+"/tokens/"+tokenID, ""), http.StatusNoContent)
	resp.Body.Close()
	assertMockMet(t, mock)
}

func TestAdminRevokeAPIToken_NotFound(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	targetID, tokenID := uuid.NewString(), uuid.NewString()
	mock.ExpectExec(`DELETE FROM api_tokens`).
		WithArgs(tokenID, targetID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp := expectStatus(t, srv, adminMutation(t, token, http.MethodDelete, "/api/admin/users/"+targetID+"/tokens/"+tokenID, ""), http.StatusNotFound)
	resp.Body.Close()
	assertMockMet(t, mock)
}

// --- Delete user ------------------------------------------------------------

func TestAdminDeleteUser_Success(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	targetID := uuid.NewString()
	expectTargetUser(mock, targetID, "member@example.com", false, nil)
	mock.ExpectExec(`DELETE FROM users u\s+WHERE u.id = \$1\s+AND NOT EXISTS \(SELECT 1 FROM clubs`).
		WithArgs(targetID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"confirm_email":" Member@Example.com "}`
	resp := expectStatus(t, srv, adminMutation(t, token, http.MethodDelete, "/api/admin/users/"+targetID, body), http.StatusNoContent)
	resp.Body.Close()
	assertMockMet(t, mock)
}

func TestAdminDeleteUser_WrongConfirmation(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	targetID := uuid.NewString()
	expectTargetUser(mock, targetID, "member@example.com", false, nil)

	body := `{"confirm_email":"other@example.com"}`
	resp := expectStatus(t, srv, adminMutation(t, token, http.MethodDelete, "/api/admin/users/"+targetID, body), http.StatusBadRequest)
	resp.Body.Close()
	assertMockMet(t, mock)
}

// An owner cannot be deleted: the cascade would delete the clubs.
func TestAdminDeleteUser_OwnerMustTransferFirst(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	targetID := uuid.NewString()
	expectTargetUser(mock, targetID, "owner@example.com", false, nil)
	mock.ExpectExec(`DELETE FROM users u`).
		WithArgs(targetID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	expectTargetUser(mock, targetID, "owner@example.com", false, nil)

	body := `{"confirm_email":"owner@example.com"}`
	resp := expectStatus(t, srv, adminMutation(t, token, http.MethodDelete, "/api/admin/users/"+targetID, body), http.StatusConflict)
	resp.Body.Close()
	assertMockMet(t, mock)
}

func TestAdminDeleteUser_SelfAndOtherAdmin(t *testing.T) {
	srv, mock, adminID, token := adminSession(t)
	expectTargetUser(mock, adminID, "op@example.com", true, nil)
	resp := expectStatus(t, srv, adminMutation(t, token, http.MethodDelete, "/api/admin/users/"+adminID, `{"confirm_email":"op@example.com"}`), http.StatusBadRequest)
	resp.Body.Close()
	assertMockMet(t, mock)

	srv, mock, _, token = adminSession(t)
	otherID := uuid.NewString()
	expectTargetUser(mock, otherID, "op2@example.com", true, nil)
	resp = expectStatus(t, srv, adminMutation(t, token, http.MethodDelete, "/api/admin/users/"+otherID, `{"confirm_email":"op2@example.com"}`), http.StatusConflict)
	resp.Body.Close()
	assertMockMet(t, mock)
}

// --- Club owner transfer and club deletion ---------------------------------

func TestAdminTransferClubOwner_Success(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	clubID, ownerID, newOwnerID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	expectTargetClub(mock, clubID, "KC Test", ownerID)
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs("new@example.com").
		WillReturnRows(sqlmock.NewRows(userColumns).
			AddRow(newOwnerID, "new@example.com", "hash", now, now, nil, false, false, nil, nil, nil))
	mock.ExpectQuery(`UPDATE clubs`).
		WithArgs(newOwnerID, sqlmock.AnyArg(), clubID).
		WillReturnRows(sqlmock.NewRows(clubColumns).AddRow(clubID, "KC Test", 1000, 0, 500, true, false, newOwnerID, now, now))

	resp := expectStatus(t, srv, adminMutation(t, token, http.MethodPost, "/api/admin/clubs/"+clubID+"/transfer-owner", `{"new_owner_email":"New@Example.com"}`), http.StatusOK)
	defer resp.Body.Close()
	var payload struct {
		ClubID string `json:"club_id"`
		Owner  struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"owner"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.ClubID != clubID || payload.Owner.ID != newOwnerID {
		t.Errorf("unexpected payload: %+v", payload)
	}
	assertMockMet(t, mock)
}

func TestAdminTransferClubOwner_DisabledNewOwner(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	clubID, ownerID := uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	expectTargetClub(mock, clubID, "KC Test", ownerID)
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs("new@example.com").
		WillReturnRows(sqlmock.NewRows(userColumns).
			AddRow(uuid.NewString(), "new@example.com", "hash", now, now, nil, false, false, nil, now, nil))

	resp := expectStatus(t, srv, adminMutation(t, token, http.MethodPost, "/api/admin/clubs/"+clubID+"/transfer-owner", `{"new_owner_email":"new@example.com"}`), http.StatusBadRequest)
	resp.Body.Close()
	assertMockMet(t, mock)
}

func TestAdminTransferClubOwner_InvalidEmail(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	resp := expectStatus(t, srv, adminMutation(t, token, http.MethodPost, "/api/admin/clubs/"+uuid.NewString()+"/transfer-owner", `{"new_owner_email":"nope"}`), http.StatusBadRequest)
	resp.Body.Close()
	assertMockMet(t, mock)
}

func TestAdminDeleteClub_Success(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	clubID := uuid.NewString()
	expectTargetClub(mock, clubID, "KC Test", uuid.NewString())
	// club.Repository.Delete checks existence again, then deletes (cascade).
	expectTargetClub(mock, clubID, "KC Test", uuid.NewString())
	mock.ExpectExec(`DELETE FROM clubs`).WithArgs(clubID).WillReturnResult(sqlmock.NewResult(0, 1))

	resp := expectStatus(t, srv, adminMutation(t, token, http.MethodDelete, "/api/admin/clubs/"+clubID, `{"confirm_name":"KC Test"}`), http.StatusNoContent)
	resp.Body.Close()
	assertMockMet(t, mock)
}

func TestAdminDeleteClub_WrongConfirmation(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	clubID := uuid.NewString()
	expectTargetClub(mock, clubID, "KC Test", uuid.NewString())

	resp := expectStatus(t, srv, adminMutation(t, token, http.MethodDelete, "/api/admin/clubs/"+clubID, `{"confirm_name":"kc test"}`), http.StatusBadRequest)
	resp.Body.Close()
	assertMockMet(t, mock)
}
