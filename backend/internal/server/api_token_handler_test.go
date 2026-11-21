package server

import (
	"database/sql/driver"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/apitoken"
	"github.com/schnurbus/go-kegelmaster/backend/internal/handlers"
)

const testAPIToken = "kgm_test-token-0123456789abcdefghijklmnopqrstuvw"

var apiTokenColumns = []string{"id", "user_id", "name", "token_hash", "token_prefix", "expires_at", "last_used_at", "created_at", "updated_at"}
var userColumns = []string{"id", "email", "password_hash", "created_at", "updated_at", "password_changed_at", "notify_gameday_settled", "is_instance_admin", "last_login_at", "disabled_at", "sessions_valid_after", "tour_seen_at", "setup_hint_seen_at"}

func expectAPITokenLookup(mock sqlmock.Sqlmock, tokenID, userID string, createdAt time.Time, expiresAt driver.Value) {
	mock.ExpectQuery(`SELECT .* FROM api_tokens\s+WHERE token_hash = \$1`).
		WithArgs(apitoken.HashToken(testAPIToken)).
		WillReturnRows(sqlmock.NewRows(apiTokenColumns).
			AddRow(tokenID, userID, "agent", apitoken.HashToken(testAPIToken), testAPIToken[:12], expiresAt, nil, createdAt, createdAt))
}

func expectUserByID(mock sqlmock.Sqlmock, userID string, passwordChangedAt driver.Value) {
	now := time.Now().UTC()
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(userColumns).AddRow(userID, "agent@example.com", "hash", now, now, passwordChangedAt, false, false, nil, nil, nil, nil, nil))
}

func bearerRequest(method, path, body string) *http.Request {
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testAPIToken)
	return req
}

func readMessage(t *testing.T, resp *http.Response) string {
	t.Helper()
	raw, _ := io.ReadAll(resp.Body)
	var body map[string]string
	_ = json.Unmarshal(raw, &body)
	return body["message"]
}

func TestBearerToken_AuthenticatesUser(t *testing.T) {
	srv, mock, _ := newTestServer(t)
	userID, tokenID := uuid.NewString(), uuid.NewString()

	expectAPITokenLookup(mock, tokenID, userID, time.Now().UTC().Add(-time.Hour), nil)
	expectUserByID(mock, userID, nil)
	mock.ExpectExec(`UPDATE api_tokens`).WithArgs(sqlmock.AnyArg(), tokenID).WillReturnResult(sqlmock.NewResult(0, 1))

	resp := doRequest(t, srv, bearerRequest(http.MethodGet, "/api/auth/me", ""))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var payload handlers.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.ID != userID {
		t.Fatalf("expected user %s, got %s", userID, payload.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBearerToken_UnknownTokenDoesNotFallBackToCookie(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	cookieToken, err := authSvc.GenerateToken(userID, false)
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(`SELECT .* FROM api_tokens`).WillReturnRows(sqlmock.NewRows(apiTokenColumns))

	req := bearerRequest(http.MethodGet, "/api/auth/me", "")
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: cookieToken})
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBearerToken_MalformedTokenSkipsDatabase(t *testing.T) {
	srv, mock, _ := newTestServer(t)

	req, _ := http.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer not-a-kegelmaster-token")
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected queries: %v", err)
	}
}

func TestBearerToken_Expired(t *testing.T) {
	srv, mock, _ := newTestServer(t)
	expectAPITokenLookup(mock, uuid.NewString(), uuid.NewString(), time.Now().UTC().Add(-48*time.Hour), time.Now().UTC().Add(-time.Minute))

	resp := doRequest(t, srv, bearerRequest(http.MethodGet, "/api/auth/me", ""))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestBearerToken_RevokedByPasswordChange(t *testing.T) {
	srv, mock, _ := newTestServer(t)
	userID := uuid.NewString()
	created := time.Now().UTC().Add(-48 * time.Hour)
	expectAPITokenLookup(mock, uuid.NewString(), userID, created, nil)
	expectUserByID(mock, userID, time.Now().UTC().Add(-time.Hour))

	resp := doRequest(t, srv, bearerRequest(http.MethodGet, "/api/auth/me", ""))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestBearerToken_MutationNeedsNoCSRF(t *testing.T) {
	srv, mock, _ := newTestServer(t)

	// Logout is a mutation guarded by CSRF; with a bearer header it must pass.
	resp := doRequest(t, srv, bearerRequest(http.MethodPost, "/api/auth/logout", ""))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected queries: %v", err)
	}
}

func TestAPITokenManagement_RejectsBearerAuth(t *testing.T) {
	srv, _, _ := newTestServer(t)

	for _, req := range []*http.Request{
		bearerRequest(http.MethodGet, "/api/auth/tokens", ""),
		bearerRequest(http.MethodPost, "/api/auth/tokens", `{"name":"x"}`),
		bearerRequest(http.MethodDelete, "/api/auth/tokens/"+uuid.NewString(), ""),
	} {
		resp := doRequest(t, srv, req)
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("%s %s: expected 403, got %d", req.Method, req.URL.Path, resp.StatusCode)
		}
		if msg := readMessage(t, resp); !strings.Contains(msg, "Weboberfläche") {
			t.Errorf("%s %s: unexpected message %q", req.Method, req.URL.Path, msg)
		}
		resp.Body.Close()
	}
}

func cookieRequest(t *testing.T, authSvc interface {
	GenerateToken(string, bool) (string, error)
}, userID, method, path, body string) *http.Request {
	t.Helper()
	token, err := authSvc.GenerateToken(userID, false)
	if err != nil {
		t.Fatal(err)
	}
	req := mustJSONRequest(t, method, path, body)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	return req
}

func TestCreateAPIToken_ReturnsSecretOnce(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	now := time.Now().UTC()

	expectUserByID(mock, userID, nil)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM api_tokens`).WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(`INSERT INTO api_tokens`).
		WithArgs(sqlmock.AnyArg(), userID, "Mein Agent", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(apiTokenColumns).AddRow(uuid.NewString(), userID, "Mein Agent", "h", "kgm_abcdefgh", now.AddDate(0, 0, 30), nil, now, now))

	resp := doRequest(t, srv, cookieRequest(t, authSvc, userID, http.MethodPost, "/api/auth/tokens", `{"name":" Mein Agent ","expires_in_days":30}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", resp.StatusCode, readMessage(t, resp))
	}
	var payload handlers.CreatedAPITokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !strings.HasPrefix(payload.Token, apitoken.Prefix) || len(payload.Token) < 40 {
		t.Fatalf("unexpected token %q", payload.Token)
	}
	if payload.ExpiresAt == nil {
		t.Fatal("expected expires_at")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateAPIToken_Validation(t *testing.T) {
	cases := map[string]string{
		"missing name":   `{"name":"  "}`,
		"too long":       `{"name":"` + strings.Repeat("x", 101) + `"}`,
		"expiry too big": `{"name":"a","expires_in_days":366}`,
		"expiry zero":    `{"name":"a","expires_in_days":0}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			srv, mock, authSvc := newTestServer(t)
			userID := uuid.NewString()
			expectUserByID(mock, userID, nil)
			resp := doRequest(t, srv, cookieRequest(t, authSvc, userID, http.MethodPost, "/api/auth/tokens", body))
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", resp.StatusCode)
			}
		})
	}
}

func TestCreateAPIToken_LimitReached(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	expectUserByID(mock, userID, nil)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM api_tokens`).WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(20))

	resp := doRequest(t, srv, cookieRequest(t, authSvc, userID, http.MethodPost, "/api/auth/tokens", `{"name":"a"}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestCreateAPIToken_RequiresCSRF(t *testing.T) {
	srv, _, authSvc := newTestServer(t)
	token, _ := authSvc.GenerateToken(uuid.NewString(), false)
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/tokens", strings.NewReader(`{"name":"a"}`))
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

func TestListAPITokens_OmitsSecret(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	now := time.Now().UTC()
	expectUserByID(mock, userID, nil)
	mock.ExpectQuery(`SELECT .* FROM api_tokens\s+WHERE user_id = \$1`).WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(apiTokenColumns).AddRow(uuid.NewString(), userID, "agent", "secret-hash", "kgm_abcdefgh", nil, now, now, now))

	resp := doRequest(t, srv, cookieRequest(t, authSvc, userID, http.MethodGet, "/api/auth/tokens", ""))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	raw, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(raw), "secret-hash") || strings.Contains(string(raw), `"token"`) {
		t.Fatalf("response leaks secret material: %s", raw)
	}
	var payload []handlers.APITokenResponse
	if err := json.Unmarshal(raw, &payload); err != nil || len(payload) != 1 || payload[0].TokenPrefix != "kgm_abcdefgh" {
		t.Fatalf("unexpected payload %s (%v)", raw, err)
	}
}

func TestDeleteAPIToken_OtherUsersTokenNotFound(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, tokenID := uuid.NewString(), uuid.NewString()
	expectUserByID(mock, userID, nil)
	mock.ExpectExec(`DELETE FROM api_tokens\s+WHERE id = \$1 AND user_id = \$2`).WithArgs(tokenID, userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp := doRequest(t, srv, cookieRequest(t, authSvc, userID, http.MethodDelete, "/api/auth/tokens/"+tokenID, ""))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteAPIToken_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, tokenID := uuid.NewString(), uuid.NewString()
	expectUserByID(mock, userID, nil)
	mock.ExpectExec(`DELETE FROM api_tokens`).WithArgs(tokenID, userID).WillReturnResult(sqlmock.NewResult(0, 1))

	resp := doRequest(t, srv, cookieRequest(t, authSvc, userID, http.MethodDelete, "/api/auth/tokens/"+tokenID, ""))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}
