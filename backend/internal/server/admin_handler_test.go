package server

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/admin"
	"github.com/schnurbus/go-kegelmaster/backend/internal/handlers"
)

var adminEndpoints = []string{
	"/api/admin/stats",
	"/api/admin/users",
	"/api/admin/users/" + "00000000-0000-0000-0000-000000000001",
	"/api/admin/clubs",
	"/api/admin/clubs/" + "00000000-0000-0000-0000-000000000002",
	"/api/admin/export/memberships.csv",
}

// expectSessionUser expects the user lookup of UserFromCookie.
func expectSessionUser(mock sqlmock.Sqlmock, userID string, isAdmin bool) {
	now := time.Now().UTC().Add(-time.Hour)
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(userColumns).
			AddRow(userID, "op@example.com", "hash", now, now, nil, false, isAdmin, nil, nil, nil))
}

func adminRequest(t *testing.T, token, path string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if token != "" {
		req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	}
	return req
}

func TestAdmin_UnauthenticatedGets404(t *testing.T) {
	srv, mock, _ := newTestServer(t)
	for _, path := range adminEndpoints {
		resp := doRequest(t, srv, adminRequest(t, "", path))
		resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("%s: expected 404, got %d", path, resp.StatusCode)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAdmin_NonAdminGets404(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	token, err := authSvc.GenerateToken(userID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	for _, path := range adminEndpoints {
		expectSessionUser(mock, userID, false)
		resp := doRequest(t, srv, adminRequest(t, token, path))
		resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("%s: expected 404, got %d", path, resp.StatusCode)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// An API token must never grant admin access, not even when a valid admin
// session cookie is sent along. The token is not even looked up.
func TestAdmin_BearerTokenGets404(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	adminID := uuid.NewString()
	token, err := authSvc.GenerateToken(adminID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	for _, path := range adminEndpoints {
		req := adminRequest(t, token, path)
		req.Header.Set("Authorization", "Bearer kgm_doesnotmatter")
		resp := doRequest(t, srv, req)
		resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("%s: expected 404, got %d", path, resp.StatusCode)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func adminSession(t *testing.T) (*Server, sqlmock.Sqlmock, string, string) {
	t.Helper()
	srv, mock, authSvc := newTestServer(t)
	adminID := uuid.NewString()
	token, err := authSvc.GenerateToken(adminID, true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	expectSessionUser(mock, adminID, true)
	return srv, mock, adminID, token
}

func TestAdmin_Stats(t *testing.T) {
	srv, mock, _, token := adminSession(t)

	mock.ExpectQuery(`SELECT\s+\(SELECT COUNT\(\*\) FROM users\)::bigint AS users_total`).
		WillReturnRows(sqlmock.NewRows([]string{
			"users_total", "users_new_7d", "users_new_30d", "users_active_30d", "users_instance_admins", "users_disabled", "users_orphaned",
			"clubs_total", "clubs_active_90d", "clubs_inactive_180d", "players_total", "players_linked",
			"invitations_open", "api_tokens_active", "api_tokens_without_expiry", "api_tokens_unused_90d", "passkeys_total",
		}).AddRow(42, 3, 9, 20, 1, 2, 4, 7, 5, 2, 80, 35, 6, 11, 8, 3, 13))
	mock.ExpectQuery(`FROM users u\s+WHERE u.created_at >= now\(\) - interval '30 days'\s+GROUP BY day`).
		WithArgs(int64(admin.RegistrationSpikeThreshold)).
		WillReturnRows(sqlmock.NewRows([]string{"day", "registrations"}).
			AddRow(time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC), 25))

	resp := doRequest(t, srv, adminRequest(t, token, "/api/admin/stats"))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if cc := resp.Header.Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control: got %q", cc)
	}
	var stats admin.Stats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if stats.Users.Total != 42 || stats.Users.Disabled != 2 || stats.Clubs.Active90d != 5 || stats.Players.Linked != 35 ||
		stats.OpenInvitations != 6 || stats.APITokens.Active != 11 || stats.Passkeys != 13 {
		t.Errorf("unexpected stats: %+v", stats)
	}
	if stats.Anomalies.OrphanedUsers != 4 || stats.Anomalies.InactiveClubs != 2 ||
		stats.Anomalies.APITokensWithoutExpiry != 8 || stats.Anomalies.APITokensUnused90d != 3 {
		t.Errorf("unexpected anomalies: %+v", stats.Anomalies)
	}
	if len(stats.Anomalies.RegistrationSpikes) != 1 || stats.Anomalies.RegistrationSpikes[0].Date != "2026-11-18" {
		t.Errorf("unexpected spikes: %+v", stats.Anomalies.RegistrationSpikes)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAdmin_ListUsers(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	userID := uuid.NewString()
	created := time.Now().UTC().AddDate(0, -2, 0)

	// The search is escaped for ILIKE; filter=orphaned sets only_orphaned.
	mock.ExpectQuery(`SELECT COUNT\(\*\)::bigint\s+FROM users u`).
		WithArgs(`a\_b`, false, false, true).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(30))
	mock.ExpectQuery(`FROM users u\s+WHERE .*ORDER BY`).
		WithArgs(`a\_b`, false, false, true, "email_asc", int32(10), int32(10)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "email", "is_instance_admin", "created_at", "last_login_at", "disabled_at",
			"owned_club_count", "player_club_count", "passkey_count", "api_token_count", "orphaned",
		}).AddRow(userID, "a_b@example.com", false, created, nil, nil, 0, 0, 0, 1, true))

	resp := doRequest(t, srv, adminRequest(t, token, "/api/admin/users?search=a_b&filter=orphaned&sort=email&page=2&limit=10"))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var page admin.Page[admin.UserListItem]
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if page.Total != 30 || page.TotalPages != 3 || page.Page != 2 || page.Limit != 10 || len(page.Data) != 1 {
		t.Fatalf("unexpected page: %+v", page)
	}
	if u := page.Data[0]; u.ID != userID || !u.Orphaned || u.LastLoginAt != nil || u.APITokenCount != 1 {
		t.Errorf("unexpected user: %+v", u)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAdmin_ListUsers_InvalidParams(t *testing.T) {
	for _, query := range []string{"sort=password_hash", "filter=evil", "order=sideways", "limit=1000", "page=0"} {
		srv, mock, _, token := adminSession(t)
		resp := doRequest(t, srv, adminRequest(t, token, "/api/admin/users?"+query))
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", query, resp.StatusCode)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet expectations: %v", err)
		}
	}
}

func TestAdmin_GetUser(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	userID := uuid.NewString()
	clubID := uuid.NewString()
	roleID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`FROM users u\s+WHERE u.id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "is_instance_admin", "created_at", "updated_at", "password_changed_at", "last_login_at", "disabled_at", "sessions_valid_after", "orphaned"}).
			AddRow(userID, "member@example.com", false, now, now, nil, now, now, nil, false))
	mock.ExpectQuery(`FROM clubs c\s+WHERE c.user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at", "player_count"}).
			AddRow(clubID, "KC Test", now, 8))
	mock.ExpectQuery(`FROM players p\s+JOIN clubs c`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "inactive", "club_id", "club_name", "is_owner", "role_id", "role_name"}).
			AddRow(uuid.NewString(), "Max", false, clubID, "KC Test", true, roleID, "Kassierer"))
	mock.ExpectQuery(`SELECT id, name, token_prefix, created_at, expires_at, last_used_at\s+FROM api_tokens`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "token_prefix", "created_at", "expires_at", "last_used_at"}).
			AddRow(uuid.NewString(), "Skript", "kgm_abcd", now, nil, nil))
	mock.ExpectQuery(`FROM webauthn_credentials`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at", "last_used_at"}))

	resp := doRequest(t, srv, adminRequest(t, token, "/api/admin/users/"+userID))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if raw["email"] != "member@example.com" {
		t.Errorf("email: got %v", raw["email"])
	}
	if raw["disabled_at"] == nil || raw["sessions_valid_after"] != nil {
		t.Errorf("disabled_at/sessions_valid_after: got %v / %v", raw["disabled_at"], raw["sessions_valid_after"])
	}
	if pk, ok := raw["passkeys"].([]any); !ok || len(pk) != 0 {
		t.Errorf("passkeys must be an empty array, got %v", raw["passkeys"])
	}
	tokens, _ := raw["api_tokens"].([]any)
	if len(tokens) != 1 {
		t.Fatalf("api_tokens: got %v", raw["api_tokens"])
	}
	for key := range tokens[0].(map[string]any) {
		if strings.Contains(key, "hash") {
			t.Errorf("api token must not expose %q", key)
		}
	}
	players, _ := raw["players"].([]any)
	if len(players) != 1 || players[0].(map[string]any)["role_name"] != "Kassierer" {
		t.Errorf("players: got %v", raw["players"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAdmin_GetUser_NotFound(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	userID := uuid.NewString()
	mock.ExpectQuery(`FROM users u\s+WHERE u.id = \$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	resp := doRequest(t, srv, adminRequest(t, token, "/api/admin/users/"+userID))
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAdmin_ListClubs(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	clubID := uuid.NewString()
	ownerID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`SELECT COUNT\(\*\)::bigint\s+FROM clubs c`).
		WithArgs("", false).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`FROM clubs c\s+JOIN users u .*ORDER BY`).
		WithArgs("", false, "name_asc", int32(0), int32(admin.DefaultPageLimit)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "created_at", "owner_id", "owner_email", "player_count", "linked_player_count", "game_day_count", "last_activity_at",
		}).
			AddRow(clubID, "KC Test", now, ownerID, "owner@example.com", 8, 5, 12, now).
			AddRow(uuid.NewString(), "KC Leer", now, ownerID, "owner@example.com", 0, 0, 0, nil))

	resp := doRequest(t, srv, adminRequest(t, token, "/api/admin/clubs"))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var page admin.Page[admin.ClubListItem]
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(page.Data) != 2 || page.Data[0].Owner.Email != "owner@example.com" || page.Data[0].LastActivityAt == nil || page.Data[1].LastActivityAt != nil {
		t.Errorf("unexpected clubs: %+v", page.Data)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAdmin_GetClub(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	clubID := uuid.NewString()
	ownerID := uuid.NewString()
	linkedUserID := uuid.NewString()
	now := time.Now().UTC()

	mock.ExpectQuery(`FROM clubs c\s+JOIN users u .*WHERE c.id = \$1`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "created_at", "owner_id", "owner_email", "player_count", "linked_player_count", "game_day_count", "transaction_count", "last_activity_at",
		}).AddRow(clubID, "KC Test", now, ownerID, "owner@example.com", 2, 1, 12, 140, now))
	mock.ExpectQuery(`FROM players p\s+LEFT JOIN users u`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "inactive", "created_at", "user_id", "user_email", "role_id", "role_name"}).
			AddRow(uuid.NewString(), "Anna", false, now, linkedUserID, "anna@example.com", nil, nil).
			AddRow(uuid.NewString(), "Bernd", true, now, nil, nil, nil, nil))
	mock.ExpectQuery(`FROM player_invitations i`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "player_id", "player_name", "email", "created_at", "expires_at"}))
	mock.ExpectQuery(`FROM roles r`).
		WithArgs(clubID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "player_count"}).AddRow(uuid.NewString(), "Mitglied", 0))

	resp := doRequest(t, srv, adminRequest(t, token, "/api/admin/clubs/"+clubID))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// No business data of the club.
	for _, key := range []string{"balance", "start_balance", "base_fee", "transactions"} {
		if _, ok := raw[key]; ok {
			t.Errorf("club detail must not contain %q", key)
		}
	}
	players, _ := raw["players"].([]any)
	if len(players) != 2 {
		t.Fatalf("players: got %v", raw["players"])
	}
	if u := players[0].(map[string]any)["user"].(map[string]any); u["email"] != "anna@example.com" {
		t.Errorf("linked user: got %v", u)
	}
	if players[1].(map[string]any)["user"] != nil {
		t.Errorf("unlinked player must have user null, got %v", players[1])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAdmin_ExportMemberships(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	mock.ExpectQuery(`UNION ALL`).
		WillReturnRows(sqlmock.NewRows([]string{"user_email", "club_name", "is_owner", "player_name", "role_name"}).
			AddRow("anna@example.com", "KC Test", false, "Anna", "Mitglied").
			AddRow("owner@example.com", "=HYPERLINK(1)", true, "", ""))

	resp := doRequest(t, srv, adminRequest(t, token, "/api/admin/export/memberships.csv"))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/csv") {
		t.Errorf("Content-Type: got %q", ct)
	}
	if cd := resp.Header.Get("Content-Disposition"); !strings.Contains(cd, "attachment") {
		t.Errorf("Content-Disposition: got %q", cd)
	}
	records, err := csv.NewReader(resp.Body).ReadAll()
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	want := [][]string{
		{"user_email", "club_name", "is_owner", "player_name", "role_name"},
		{"anna@example.com", "KC Test", "false", "Anna", "Mitglied"},
		{"owner@example.com", "'=HYPERLINK(1)", "true", "", ""},
	}
	if len(records) != len(want) {
		t.Fatalf("records: got %v", records)
	}
	for i := range want {
		if strings.Join(records[i], "|") != strings.Join(want[i], "|") {
			t.Errorf("row %d: got %v, want %v", i, records[i], want[i])
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCurrentUser_IncludesInstanceAdminFlag(t *testing.T) {
	srv, mock, _, token := adminSession(t)
	resp := doRequest(t, srv, adminRequest(t, token, "/api/auth/me"))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var payload handlers.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !payload.IsInstanceAdmin {
		t.Errorf("expected is_instance_admin=true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
