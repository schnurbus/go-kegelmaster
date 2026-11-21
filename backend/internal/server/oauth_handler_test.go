package server

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/auth"
	"github.com/schnurbus/go-kegelmaster/backend/internal/config"
	"github.com/schnurbus/go-kegelmaster/backend/internal/oauth"
)

const (
	testIssuer        = "https://kegel.example.com"
	testRedirectURI   = "https://client.example/callback"
	testCodeVerifier  = "verifier-0123456789-abcdefghijklmnopqrstuvwxyz"
	testOAuthCode     = "kgc_test-code-0123456789abcdefghijklmnopqrstuvwx"
	testRefreshToken  = "kgr_test-refresh-0123456789abcdefghijklmnopqrstu"
	testOAuthAccess   = "kgo_test-access-0123456789abcdefghijklmnopqrstuv"
	testOAuthStateVal = "xyz-state"
)

var (
	oauthClientColumns = []string{"id", "client_name", "redirect_uris", "created_at", "updated_at"}
	oauthCodeColumns   = []string{"id", "code_hash", "client_id", "user_id", "redirect_uri", "code_challenge", "scope", "resource", "expires_at", "used_at", "grant_id", "created_at", "updated_at"}
	oauthGrantColumns  = []string{"id", "client_id", "user_id", "scope", "resource", "last_used_at", "created_at", "updated_at"}
	oauthTokenColumns  = []string{"id", "grant_id", "token_type", "expires_at", "used_at", "client_id", "user_id", "grant_created_at"}
)

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func newOAuthTestServer(t *testing.T) (*Server, sqlmock.Sqlmock, *auth.Service) {
	t.Helper()
	return newTestServerWithConfig(t, func(c *config.Config) { c.BaseURL = testIssuer + "/" })
}

func expectOAuthClient(mock sqlmock.Sqlmock, clientID string) {
	now := time.Now().UTC()
	mock.ExpectQuery(`SELECT .* FROM oauth_clients\s+WHERE id = \$1`).
		WithArgs(clientID).
		WillReturnRows(sqlmock.NewRows(oauthClientColumns).AddRow(clientID, "Claude", "{"+testRedirectURI+"}", now, now))
}

// expectConsumeCode expects the atomic single-use update of a code.
func expectConsumeCode(mock sqlmock.Sqlmock, clientID, userID string, expiresAt time.Time) string {
	codeID := uuid.NewString()
	now := time.Now().UTC()
	mock.ExpectQuery(`UPDATE oauth_authorization_codes\s+SET used_at`).
		WithArgs(sqlmock.AnyArg(), oauth.HashToken(testOAuthCode)).
		WillReturnRows(sqlmock.NewRows(oauthCodeColumns).AddRow(
			codeID, oauth.HashToken(testOAuthCode), clientID, userID, testRedirectURI, pkceChallenge(testCodeVerifier),
			oauth.Scope, testIssuer+"/mcp", expiresAt, now, nil, now, now))
	return codeID
}

func tokenRequest(form url.Values) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func codeForm(clientID string) url.Values {
	return url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {testOAuthCode},
		"redirect_uri":  {testRedirectURI},
		"client_id":     {clientID},
		"code_verifier": {testCodeVerifier},
		"resource":      {testIssuer + "/mcp"},
	}
}

func decodeOAuthError(t *testing.T, resp *http.Response) string {
	t.Helper()
	var body struct {
		Error string `json:"error"`
	}
	raw, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(raw, &body)
	return body.Error
}

func expectMet(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestOAuth_Metadata(t *testing.T) {
	srv, _, _ := newOAuthTestServer(t)

	for _, path := range []string{"/.well-known/oauth-protected-resource", "/.well-known/oauth-protected-resource/mcp"} {
		resp := doRequest(t, srv, httptestGet(path))
		var prm struct {
			Resource             string   `json:"resource"`
			AuthorizationServers []string `json:"authorization_servers"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&prm)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK || prm.Resource != testIssuer+"/mcp" || len(prm.AuthorizationServers) != 1 || prm.AuthorizationServers[0] != testIssuer {
			t.Fatalf("%s: unexpected metadata %d %+v", path, resp.StatusCode, prm)
		}
	}

	resp := doRequest(t, srv, httptestGet("/.well-known/oauth-authorization-server"))
	var as map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&as)
	resp.Body.Close()
	if as["issuer"] != testIssuer || as["token_endpoint"] != testIssuer+"/oauth/token" || as["registration_endpoint"] != testIssuer+"/oauth/register" {
		t.Fatalf("unexpected AS metadata: %v", as)
	}
	if methods, _ := as["code_challenge_methods_supported"].([]any); len(methods) != 1 || methods[0] != "S256" {
		t.Fatalf("expected only S256, got %v", as["code_challenge_methods_supported"])
	}

	resp = doRequest(t, srv, httptestGet("/.well-known/openid-configuration"))
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown well-known path, got %d", resp.StatusCode)
	}
}

func httptestGet(path string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	return req
}

func TestOAuth_RegisterValidatesRedirectURIs(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)

	for name, body := range map[string]string{
		"http non-loopback": `{"redirect_uris":["http://evil.example/cb"]}`,
		"fragment":          `{"redirect_uris":["https://client.example/cb#x"]}`,
		"no uris":           `{"client_name":"x"}`,
		"confidential":      `{"redirect_uris":["https://client.example/cb"],"token_endpoint_auth_method":"client_secret_basic"}`,
		"implicit grant":    `{"redirect_uris":["https://client.example/cb"],"grant_types":["implicit"]}`,
	} {
		req, _ := http.NewRequest(http.MethodPost, "/oauth/register", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := doRequest(t, srv, req)
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", name, resp.StatusCode)
		}
	}

	now := time.Now().UTC()
	mock.ExpectQuery(`INSERT INTO oauth_clients`).
		WithArgs(sqlmock.AnyArg(), "Claude", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(oauthClientColumns).AddRow(uuid.NewString(), "Claude", "{https://claude.ai/api/mcp/auth_callback,http://localhost:3334/cb}", now, now))
	req, _ := http.NewRequest(http.MethodPost, "/oauth/register", strings.NewReader(
		`{"client_name":"Claude","redirect_uris":["https://claude.ai/api/mcp/auth_callback","http://localhost:3334/cb"],"token_endpoint_auth_method":"none","grant_types":["authorization_code","refresh_token"],"response_types":["code"]}`))
	req.Header.Set("Content-Type", "application/json")
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()
	var reg map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&reg)
	if resp.StatusCode != http.StatusCreated || reg["client_id"] == "" || reg["token_endpoint_auth_method"] != "none" {
		t.Fatalf("unexpected registration response %d %v", resp.StatusCode, reg)
	}
	if _, ok := reg["client_secret"]; ok {
		t.Fatal("public clients must not get a secret")
	}
	expectMet(t, mock)
}

func authorizeQuery(clientID, redirectURI string) url.Values {
	return url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"code_challenge":        {pkceChallenge(testCodeVerifier)},
		"code_challenge_method": {"S256"},
		"state":                 {testOAuthStateVal},
		"resource":              {testIssuer + "/mcp"},
	}
}

func TestOAuth_AuthorizeRedirectsToConsent(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)
	clientID := uuid.NewString()
	expectOAuthClient(mock, clientID)

	resp := doRequest(t, srv, httptestGet("/oauth/authorize?"+authorizeQuery(clientID, testRedirectURI).Encode()))
	resp.Body.Close()
	loc := resp.Header.Get("Location")
	if resp.StatusCode != http.StatusFound || !strings.HasPrefix(loc, "/oauth/consent?") || !strings.Contains(loc, "state="+testOAuthStateVal) {
		t.Fatalf("expected redirect to consent, got %d %q", resp.StatusCode, loc)
	}
	expectMet(t, mock)
}

func TestOAuth_AuthorizeRejectsUnregisteredRedirectWithoutRedirecting(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)
	clientID := uuid.NewString()
	expectOAuthClient(mock, clientID)

	resp := doRequest(t, srv, httptestGet("/oauth/authorize?"+authorizeQuery(clientID, "https://evil.example/callback").Encode()))
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest || resp.Header.Get("Location") != "" {
		t.Fatalf("expected 400 without redirect, got %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}
	expectMet(t, mock)
}

func TestOAuth_AuthorizeRequiresS256(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)
	clientID := uuid.NewString()
	expectOAuthClient(mock, clientID)

	q := authorizeQuery(clientID, testRedirectURI)
	q.Set("code_challenge_method", "plain")
	resp := doRequest(t, srv, httptestGet("/oauth/authorize?"+q.Encode()))
	resp.Body.Close()
	loc, _ := url.Parse(resp.Header.Get("Location"))
	if resp.StatusCode != http.StatusFound || loc == nil || !strings.HasPrefix(loc.String(), testRedirectURI) ||
		loc.Query().Get("error") != "invalid_request" || loc.Query().Get("state") != testOAuthStateVal || loc.Query().Get("iss") != testIssuer {
		t.Fatalf("expected error redirect to client, got %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}
	expectMet(t, mock)
}

func consentRequest(t *testing.T, authSvc *auth.Service, userID, clientID string, approve bool) *http.Request {
	t.Helper()
	body := map[string]any{"approve": approve}
	for k, v := range authorizeQuery(clientID, testRedirectURI) {
		body[k] = v[0]
	}
	raw, _ := json.Marshal(body)
	req := mustJSONRequest(t, http.MethodPost, "/api/oauth/authorize", string(raw))
	cookie, err := authSvc.GenerateToken(userID, false)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: cookie})
	return req
}

func TestOAuth_ConsentIssuesCode(t *testing.T) {
	srv, mock, authSvc := newOAuthTestServer(t)
	userID, clientID := uuid.NewString(), uuid.NewString()
	expectUserByID(mock, userID, nil)
	expectOAuthClient(mock, clientID)
	mock.ExpectExec(`DELETE FROM oauth_authorization_codes`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO oauth_authorization_codes`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), clientID, userID, testRedirectURI, pkceChallenge(testCodeVerifier), oauth.Scope, testIssuer+"/mcp", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp := doRequest(t, srv, consentRequest(t, authSvc, userID, clientID, true))
	defer resp.Body.Close()
	var out struct {
		RedirectTo string `json:"redirect_to"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	loc, _ := url.Parse(out.RedirectTo)
	if resp.StatusCode != http.StatusOK || loc == nil || !strings.HasPrefix(loc.Query().Get("code"), oauth.CodePrefix) || loc.Query().Get("state") != testOAuthStateVal {
		t.Fatalf("unexpected consent response %d %q", resp.StatusCode, out.RedirectTo)
	}
	expectMet(t, mock)
}

func TestOAuth_ConsentDeniedAndCSRFAndBearer(t *testing.T) {
	srv, mock, authSvc := newOAuthTestServer(t)
	userID, clientID := uuid.NewString(), uuid.NewString()

	// Denied: redirect with access_denied, no code.
	expectUserByID(mock, userID, nil)
	expectOAuthClient(mock, clientID)
	resp := doRequest(t, srv, consentRequest(t, authSvc, userID, clientID, false))
	var out struct {
		RedirectTo string `json:"redirect_to"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	resp.Body.Close()
	if loc, _ := url.Parse(out.RedirectTo); loc == nil || loc.Query().Get("error") != "access_denied" || loc.Query().Get("code") != "" {
		t.Fatalf("expected access_denied redirect, got %q", out.RedirectTo)
	}

	// Missing CSRF token.
	req := consentRequest(t, authSvc, userID, clientID, true)
	req.Header.Del("X-CSRF-Token")
	resp = doRequest(t, srv, req)
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 without CSRF token, got %d", resp.StatusCode)
	}

	// A bearer token can never approve clients.
	req = consentRequest(t, authSvc, userID, clientID, true)
	req.Header.Set("Authorization", "Bearer "+testAPIToken)
	resp = doRequest(t, srv, req)
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 with bearer token, got %d", resp.StatusCode)
	}
	expectMet(t, mock)
}

func TestOAuth_TokenExchangeIssuesTokens(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)
	userID, clientID := uuid.NewString(), uuid.NewString()
	expectOAuthClient(mock, clientID)
	codeID := expectConsumeCode(mock, clientID, userID, time.Now().UTC().Add(5*time.Minute))
	expectUserByID(mock, userID, nil)
	now := time.Now().UTC()
	grantID := uuid.NewString()
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO oauth_grants`).
		WithArgs(sqlmock.AnyArg(), clientID, userID, oauth.Scope, testIssuer+"/mcp", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(oauthGrantColumns).AddRow(grantID, clientID, userID, oauth.Scope, testIssuer+"/mcp", nil, now, now))
	mock.ExpectExec(`UPDATE oauth_authorization_codes\s+SET grant_id`).WithArgs(grantID, sqlmock.AnyArg(), codeID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO oauth_tokens`).WithArgs(sqlmock.AnyArg(), grantID, "access", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO oauth_tokens`).WithArgs(sqlmock.AnyArg(), grantID, "refresh", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	resp := doRequest(t, srv, tokenRequest(codeForm(clientID)))
	defer resp.Body.Close()
	var tok struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&tok)
	if resp.StatusCode != http.StatusOK || !strings.HasPrefix(tok.AccessToken, oauth.AccessTokenPrefix) ||
		!strings.HasPrefix(tok.RefreshToken, oauth.RefreshTokenPrefix) || tok.TokenType != "Bearer" || tok.ExpiresIn != 3600 {
		t.Fatalf("unexpected token response %d %+v", resp.StatusCode, tok)
	}
	if resp.Header.Get("Cache-Control") != "no-store" {
		t.Errorf("token response must not be cached")
	}
	expectMet(t, mock)
}

func TestOAuth_TokenExchangeRejectsWrongPKCEVerifier(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)
	userID, clientID := uuid.NewString(), uuid.NewString()
	expectOAuthClient(mock, clientID)
	expectConsumeCode(mock, clientID, userID, time.Now().UTC().Add(5*time.Minute))

	form := codeForm(clientID)
	form.Set("code_verifier", strings.Repeat("x", 43))
	resp := doRequest(t, srv, tokenRequest(form))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest || decodeOAuthError(t, resp) != "invalid_grant" {
		t.Fatalf("expected invalid_grant, got %d", resp.StatusCode)
	}
	expectMet(t, mock)
}

func TestOAuth_TokenExchangeRejectsExpiredCode(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)
	userID, clientID := uuid.NewString(), uuid.NewString()
	expectOAuthClient(mock, clientID)
	expectConsumeCode(mock, clientID, userID, time.Now().UTC().Add(-time.Second))

	resp := doRequest(t, srv, tokenRequest(codeForm(clientID)))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest || decodeOAuthError(t, resp) != "invalid_grant" {
		t.Fatalf("expected invalid_grant, got %d", resp.StatusCode)
	}
	expectMet(t, mock)
}

func TestOAuth_TokenExchangeRejectsRedirectURIMismatch(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)
	userID, clientID := uuid.NewString(), uuid.NewString()
	expectOAuthClient(mock, clientID)
	expectConsumeCode(mock, clientID, userID, time.Now().UTC().Add(5*time.Minute))

	form := codeForm(clientID)
	form.Set("redirect_uri", "https://client.example/other")
	resp := doRequest(t, srv, tokenRequest(form))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest || decodeOAuthError(t, resp) != "invalid_grant" {
		t.Fatalf("expected invalid_grant, got %d", resp.StatusCode)
	}
	expectMet(t, mock)
}

func TestOAuth_TokenExchangeRejectsCodeOfOtherClient(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)
	userID, clientID := uuid.NewString(), uuid.NewString()
	expectOAuthClient(mock, clientID)
	expectConsumeCode(mock, uuid.NewString(), userID, time.Now().UTC().Add(5*time.Minute))

	resp := doRequest(t, srv, tokenRequest(codeForm(clientID)))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest || decodeOAuthError(t, resp) != "invalid_grant" {
		t.Fatalf("expected invalid_grant, got %d", resp.StatusCode)
	}
	expectMet(t, mock)
}

func TestOAuth_CodeIsSingleUseAndReuseRevokesGrant(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)
	userID, clientID, grantID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	expectOAuthClient(mock, clientID)
	// The code was already consumed: the atomic update matches no row.
	mock.ExpectQuery(`UPDATE oauth_authorization_codes\s+SET used_at`).WillReturnRows(sqlmock.NewRows(oauthCodeColumns))
	mock.ExpectQuery(`SELECT .* FROM oauth_authorization_codes\s+WHERE code_hash = \$1`).
		WithArgs(oauth.HashToken(testOAuthCode)).
		WillReturnRows(sqlmock.NewRows(oauthCodeColumns).AddRow(
			uuid.NewString(), oauth.HashToken(testOAuthCode), clientID, userID, testRedirectURI, pkceChallenge(testCodeVerifier),
			oauth.Scope, testIssuer+"/mcp", now.Add(5*time.Minute), now, grantID, now, now))
	mock.ExpectExec(`DELETE FROM oauth_grants\s+WHERE id = \$1`).WithArgs(grantID).WillReturnResult(sqlmock.NewResult(0, 1))

	resp := doRequest(t, srv, tokenRequest(codeForm(clientID)))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest || decodeOAuthError(t, resp) != "invalid_grant" {
		t.Fatalf("expected invalid_grant, got %d", resp.StatusCode)
	}
	expectMet(t, mock)
}

func refreshForm(clientID string) url.Values {
	return url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {testRefreshToken},
		"client_id":     {clientID},
	}
}

func TestOAuth_RefreshRotatesTokens(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)
	userID, clientID, grantID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	expectOAuthClient(mock, clientID)
	mock.ExpectQuery(`UPDATE oauth_tokens\s+SET used_at`).
		WithArgs(sqlmock.AnyArg(), oauth.HashToken(testRefreshToken)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "grant_id", "expires_at"}).AddRow(uuid.NewString(), grantID, now.Add(24*time.Hour)))
	mock.ExpectQuery(`SELECT .* FROM oauth_grants\s+WHERE id = \$1`).WithArgs(grantID).
		WillReturnRows(sqlmock.NewRows(oauthGrantColumns).AddRow(grantID, clientID, userID, oauth.Scope, testIssuer+"/mcp", nil, now.Add(-time.Hour), now))
	expectUserByID(mock, userID, nil)
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO oauth_tokens`).WithArgs(sqlmock.AnyArg(), grantID, "access", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO oauth_tokens`).WithArgs(sqlmock.AnyArg(), grantID, "refresh", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM oauth_tokens`).WithArgs(grantID, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	resp := doRequest(t, srv, tokenRequest(refreshForm(clientID)))
	defer resp.Body.Close()
	var tok struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&tok)
	if resp.StatusCode != http.StatusOK || tok.RefreshToken == testRefreshToken || !strings.HasPrefix(tok.RefreshToken, oauth.RefreshTokenPrefix) || !strings.HasPrefix(tok.AccessToken, oauth.AccessTokenPrefix) {
		t.Fatalf("expected rotated tokens, got %d %+v", resp.StatusCode, tok)
	}
	expectMet(t, mock)
}

func TestOAuth_RefreshTokenReuseRevokesGrant(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)
	userID, clientID, grantID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	expectOAuthClient(mock, clientID)
	// Already rotated: the atomic update matches no row, the lookup shows used_at.
	mock.ExpectQuery(`UPDATE oauth_tokens\s+SET used_at`).WillReturnRows(sqlmock.NewRows([]string{"id", "grant_id", "expires_at"}))
	mock.ExpectQuery(`FROM oauth_tokens t\s+JOIN oauth_grants g`).
		WithArgs(oauth.HashToken(testRefreshToken)).
		WillReturnRows(sqlmock.NewRows(oauthTokenColumns).AddRow(uuid.NewString(), grantID, "refresh", now.Add(24*time.Hour), now.Add(-time.Minute), clientID, userID, now.Add(-time.Hour)))
	mock.ExpectExec(`DELETE FROM oauth_grants\s+WHERE id = \$1`).WithArgs(grantID).WillReturnResult(sqlmock.NewResult(0, 1))

	resp := doRequest(t, srv, tokenRequest(refreshForm(clientID)))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest || decodeOAuthError(t, resp) != "invalid_grant" {
		t.Fatalf("expected invalid_grant, got %d", resp.StatusCode)
	}
	expectMet(t, mock)
}

func TestOAuth_RefreshRejectedAfterPasswordChange(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)
	userID, clientID, grantID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	expectOAuthClient(mock, clientID)
	mock.ExpectQuery(`UPDATE oauth_tokens\s+SET used_at`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "grant_id", "expires_at"}).AddRow(uuid.NewString(), grantID, now.Add(24*time.Hour)))
	mock.ExpectQuery(`SELECT .* FROM oauth_grants\s+WHERE id = \$1`).WithArgs(grantID).
		WillReturnRows(sqlmock.NewRows(oauthGrantColumns).AddRow(grantID, clientID, userID, oauth.Scope, testIssuer+"/mcp", nil, now.Add(-time.Hour), now))
	expectUserByID(mock, userID, now.Add(-time.Minute))

	resp := doRequest(t, srv, tokenRequest(refreshForm(clientID)))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest || decodeOAuthError(t, resp) != "invalid_grant" {
		t.Fatalf("expected invalid_grant, got %d", resp.StatusCode)
	}
	expectMet(t, mock)
}

func TestOAuth_TokenEndpointValidation(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)

	// JSON instead of form encoding.
	req, _ := http.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(`{"grant_type":"authorization_code"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := doRequest(t, srv, req)
	if resp.StatusCode != http.StatusBadRequest || decodeOAuthError(t, resp) != "invalid_request" {
		t.Fatalf("expected invalid_request, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Unknown client: no database lookup for a malformed client_id.
	resp = doRequest(t, srv, tokenRequest(url.Values{"grant_type": {"authorization_code"}, "client_id": {"nope"}}))
	if resp.StatusCode != http.StatusUnauthorized || decodeOAuthError(t, resp) != "invalid_client" {
		t.Fatalf("expected invalid_client, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Repeated parameters are rejected.
	resp = doRequest(t, srv, tokenRequest(url.Values{"grant_type": {"refresh_token", "authorization_code"}}))
	if resp.StatusCode != http.StatusBadRequest || decodeOAuthError(t, resp) != "invalid_request" {
		t.Fatalf("expected invalid_request, got %d", resp.StatusCode)
	}
	resp.Body.Close()
	expectMet(t, mock)
}

func TestMCP_UnauthorizedAnnouncesResourceMetadata(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)

	resp := doRequest(t, srv, mcpRequest(""))
	resp.Body.Close()
	want := `Bearer resource_metadata="` + testIssuer + `/.well-known/oauth-protected-resource"`
	if resp.StatusCode != http.StatusUnauthorized || resp.Header.Get("WWW-Authenticate") != want {
		t.Fatalf("expected 401 with %q, got %d %q", want, resp.StatusCode, resp.Header.Get("WWW-Authenticate"))
	}

	// Unknown OAuth token: 401 with the same challenge.
	mock.ExpectQuery(`FROM oauth_tokens t\s+JOIN oauth_grants g`).WillReturnRows(sqlmock.NewRows(oauthTokenColumns))
	resp = doRequest(t, srv, mcpRequest("Bearer "+testOAuthAccess))
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized || resp.Header.Get("WWW-Authenticate") != want {
		t.Fatalf("expected 401 with challenge for invalid token, got %d %q", resp.StatusCode, resp.Header.Get("WWW-Authenticate"))
	}
	expectMet(t, mock)
}

func expectOAuthAccessToken(mock sqlmock.Sqlmock, grantID, clientID, userID string, expiresAt, grantCreatedAt time.Time) {
	mock.ExpectQuery(`FROM oauth_tokens t\s+JOIN oauth_grants g`).
		WithArgs(oauth.HashToken(testOAuthAccess)).
		WillReturnRows(sqlmock.NewRows(oauthTokenColumns).AddRow(uuid.NewString(), grantID, "access", expiresAt, nil, clientID, userID, grantCreatedAt))
}

func TestMCP_ListsToolsWithOAuthToken(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)
	userID, clientID, grantID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()

	expectOAuthAccessToken(mock, grantID, clientID, userID, now.Add(time.Hour), now.Add(-time.Hour))
	expectUserByID(mock, userID, nil)
	mock.ExpectExec(`UPDATE oauth_grants`).WithArgs(sqlmock.AnyArg(), grantID).WillReturnResult(sqlmock.NewResult(0, 1))

	resp := doRequest(t, srv, mcpRequest("Bearer "+testOAuthAccess))
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(raw), `"klubs_auflisten"`) {
		t.Fatalf("expected tools/list with OAuth token, got %d: %s", resp.StatusCode, raw)
	}
	expectMet(t, mock)
}

func TestMCP_ExpiredOAuthTokenRejected(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)
	now := time.Now().UTC()
	expectOAuthAccessToken(mock, uuid.NewString(), uuid.NewString(), uuid.NewString(), now.Add(-time.Second), now.Add(-2*time.Hour))

	resp := doRequest(t, srv, mcpRequest("Bearer "+testOAuthAccess))
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired OAuth token, got %d", resp.StatusCode)
	}
	expectMet(t, mock)
}

// OAuth access tokens are bound to the MCP resource: the REST API (and with
// it the admin area) rejects them unless the MCP server dispatched the request.
func TestOAuthToken_NotAcceptedByRESTAPIDirectly(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)

	for _, path := range []string{"/api/auth/me", "/api/clubs/", "/api/admin/stats"} {
		req, _ := http.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+testOAuthAccess)
		resp := doRequest(t, srv, req)
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusNotFound {
			t.Fatalf("%s: expected 401/404 for OAuth token, got %d", path, resp.StatusCode)
		}
	}
	// No database access: the token is rejected before it is looked up.
	expectMet(t, mock)
}

func TestOAuth_RevokeEndpointRevokesGrant(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)
	userID, clientID, grantID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	expectOAuthClient(mock, clientID)
	expectOAuthAccessToken(mock, grantID, clientID, userID, now.Add(time.Hour), now.Add(-time.Hour))
	mock.ExpectExec(`DELETE FROM oauth_grants\s+WHERE id = \$1`).WithArgs(grantID).WillReturnResult(sqlmock.NewResult(0, 1))

	req, _ := http.NewRequest(http.MethodPost, "/oauth/revoke", strings.NewReader(url.Values{"token": {testOAuthAccess}, "client_id": {clientID}}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp := doRequest(t, srv, req)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	expectMet(t, mock)
}

func TestOAuth_RevokeIgnoresTokenOfOtherClient(t *testing.T) {
	srv, mock, _ := newOAuthTestServer(t)
	clientID := uuid.NewString()
	now := time.Now().UTC()
	expectOAuthClient(mock, clientID)
	expectOAuthAccessToken(mock, uuid.NewString(), uuid.NewString(), uuid.NewString(), now.Add(time.Hour), now.Add(-time.Hour))

	req, _ := http.NewRequest(http.MethodPost, "/oauth/revoke", strings.NewReader(url.Values{"token": {testOAuthAccess}, "client_id": {clientID}}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp := doRequest(t, srv, req)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	expectMet(t, mock) // no DELETE expected
}

func oauthSessionRequest(t *testing.T, authSvc *auth.Service, method, path, userID string) *http.Request {
	t.Helper()
	req := mustJSONRequest(t, method, path, "")
	cookie, err := authSvc.GenerateToken(userID, false)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: cookie})
	return req
}

func TestOAuth_ListAndRevokeConnectedApplications(t *testing.T) {
	srv, mock, authSvc := newOAuthTestServer(t)
	userID, clientID, grantID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()

	expectUserByID(mock, userID, nil)
	mock.ExpectQuery(`FROM oauth_grants g\s+JOIN oauth_clients c`).WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "client_id", "scope", "last_used_at", "created_at", "client_name", "redirect_uris"}).
			AddRow(grantID, clientID, oauth.Scope, nil, now, "Claude", "{https://claude.ai/api/mcp/auth_callback}"))
	resp := doRequest(t, srv, oauthSessionRequest(t, authSvc, http.MethodGet, "/api/oauth/grants", userID))
	var grants []map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&grants)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || len(grants) != 1 || grants[0]["client_name"] != "Claude" || grants[0]["redirect_host"] != "claude.ai" {
		t.Fatalf("unexpected grants %d %v", resp.StatusCode, grants)
	}

	expectUserByID(mock, userID, nil)
	mock.ExpectExec(`DELETE FROM oauth_grants\s+WHERE id = \$1 AND user_id = \$2`).WithArgs(grantID, userID).WillReturnResult(sqlmock.NewResult(0, 1))
	resp = doRequest(t, srv, oauthSessionRequest(t, authSvc, http.MethodDelete, "/api/oauth/grants/"+grantID, userID))
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}

	// Grants of other users are not found.
	expectUserByID(mock, userID, nil)
	mock.ExpectExec(`DELETE FROM oauth_grants\s+WHERE id = \$1 AND user_id = \$2`).WithArgs(grantID, userID).WillReturnResult(sqlmock.NewResult(0, 0))
	resp = doRequest(t, srv, oauthSessionRequest(t, authSvc, http.MethodDelete, "/api/oauth/grants/"+grantID, userID))
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	// Bearer tokens cannot manage connections.
	resp = doRequest(t, srv, bearerRequest(http.MethodGet, "/api/oauth/grants", ""))
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for bearer token, got %d", resp.StatusCode)
	}
	expectMet(t, mock)
}
