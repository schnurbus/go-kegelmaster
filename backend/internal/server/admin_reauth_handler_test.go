package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-webauthn/webauthn/protocol/webauthncbor"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/auth"
	"github.com/schnurbus/go-kegelmaster/backend/internal/config"
	"github.com/schnurbus/go-kegelmaster/backend/internal/handlers"
)

// userRow builds a users row for userColumns by column name, so the helper
// keeps working when columns are added (new nullable columns stay nil).
func userRow(userID, email, passwordHash string, isAdmin bool) *sqlmock.Rows {
	now := time.Now().UTC().Add(-24 * time.Hour)
	known := map[string]driver.Value{
		"id":                     userID,
		"email":                  email,
		"password_hash":          passwordHash,
		"created_at":             now,
		"updated_at":             now,
		"notify_gameday_settled": false,
		"is_instance_admin":      isAdmin,
	}
	values := make([]driver.Value, len(userColumns))
	for i, col := range userColumns {
		values[i] = known[col]
	}
	return sqlmock.NewRows(userColumns).AddRow(values...)
}

func expectReauthUser(mock sqlmock.Sqlmock, userID, passwordHash string, isAdmin bool) {
	mock.ExpectQuery(`SELECT id, email, password_hash`).
		WithArgs(userID).
		WillReturnRows(userRow(userID, "op@example.com", passwordHash, isAdmin))
}

// sessionToken signs a session JWT like auth.Service does, but with a chosen
// iat and auth_time (nil = cookie from before the auth_time claim existed).
func sessionToken(t *testing.T, userID string, issuedAt time.Time, authTime *time.Time) string {
	t.Helper()
	claims := auth.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(30 * 24 * time.Hour)),
		},
	}
	if authTime != nil {
		claims.AuthTime = jwt.NewNumericDate(*authTime)
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}

func timePtr(t time.Time) *time.Time { return &t }

func assertReauthRequired(t *testing.T, resp *http.Response) {
	t.Helper()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	var body struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error != handlers.ErrorCodeReauthRequired || body.Message == "" {
		t.Fatalf("unexpected body %+v", body)
	}
}

// The admin export needs no further queries, so it is a cheap probe of the
// middleware.
const reauthProbePath = "/api/admin/export/memberships.csv"

func expectMembershipExport(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`UNION ALL`).
		WillReturnRows(sqlmock.NewRows([]string{"user_email", "club_name", "is_owner", "player_name", "role_name"}))
}

func TestAdminReauth_FreshSessionPasses(t *testing.T) {
	srv, mock, _ := newTestServer(t)
	adminID := uuid.NewString()
	now := time.Now().UTC()
	token := sessionToken(t, adminID, now.Add(-2*time.Hour), timePtr(now.Add(-5*time.Minute)))
	expectReauthUser(mock, adminID, "hash", true)
	expectMembershipExport(mock)

	resp := doRequest(t, srv, adminRequest(t, token, reauthProbePath))
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAdminReauth_StaleOrMissingAuthTime(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name     string
		authTime *time.Time
	}{
		{name: "stale auth_time", authTime: timePtr(now.Add(-16 * time.Minute))},
		{name: "missing auth_time (old cookie)", authTime: nil},
		{name: "auth_time far in the future", authTime: timePtr(now.Add(time.Hour))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, mock, _ := newTestServer(t)
			adminID := uuid.NewString()
			// A recently issued cookie (e.g. reissued) does not help: only auth_time counts.
			token := sessionToken(t, adminID, now.Add(-time.Minute), tt.authTime)
			for _, path := range adminEndpoints {
				expectReauthUser(mock, adminID, "hash", true)
				resp := doRequest(t, srv, adminRequest(t, token, path))
				assertReauthRequired(t, resp)
				if cc := resp.Header.Get("Cache-Control"); cc != "no-store" {
					t.Errorf("%s: Cache-Control %q", path, cc)
				}
				resp.Body.Close()
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		})
	}
}

func TestAdminReauth_ConfiguredWindow(t *testing.T) {
	srv, mock, _ := newTestServerWithConfig(t, func(cfg *config.Config) { cfg.AdminReauthMinutes = 60 })
	adminID := uuid.NewString()
	now := time.Now().UTC()
	token := sessionToken(t, adminID, now.Add(-time.Hour), timePtr(now.Add(-30*time.Minute)))
	expectReauthUser(mock, adminID, "hash", true)
	expectMembershipExport(mock)

	resp := doRequest(t, srv, adminRequest(t, token, reauthProbePath))
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 within a 60 minute window, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// Non-admins and bearer requests keep getting 404, whatever their auth_time:
// the reauth answer must not reveal that the admin area exists.
func TestAdminReauth_NonAdminStillGets404(t *testing.T) {
	srv, mock, _ := newTestServer(t)
	userID := uuid.NewString()
	token := sessionToken(t, userID, time.Now().Add(-time.Hour), nil)
	for _, path := range adminEndpoints {
		expectReauthUser(mock, userID, "hash", false)
		resp := doRequest(t, srv, adminRequest(t, token, path))
		resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("%s: expected 404, got %d", path, resp.StatusCode)
		}
	}
	for _, path := range adminEndpoints {
		req := adminRequest(t, token, path)
		req.Header.Set("Authorization", "Bearer kgm_doesnotmatter")
		resp := doRequest(t, srv, req)
		resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("%s with bearer: expected 404, got %d", path, resp.StatusCode)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func reauthCookieAuthTime(t *testing.T, authSvc *auth.Service, resp *http.Response) (auth.Claims, time.Time) {
	t.Helper()
	cookie := findCookie(resp.Cookies(), authCookieName)
	if cookie == nil || cookie.Value == "" {
		t.Fatalf("expected a reissued auth cookie")
	}
	claims, err := authSvc.ParseToken(cookie.Value)
	if err != nil {
		t.Fatalf("parse reissued token: %v", err)
	}
	at, ok := claims.AuthenticatedAt()
	if !ok {
		t.Fatalf("reissued token has no auth_time")
	}
	return claims, at
}

func TestReauthPassword_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	adminID := uuid.NewString()
	hash, err := authSvc.HashPassword("geheim-123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	now := time.Now().UTC()
	// A remember-me session whose last authentication is stale.
	stale := sessionToken(t, adminID, now.Add(-3*time.Hour), timePtr(now.Add(-3*time.Hour)))
	expectReauthUser(mock, adminID, hash, true)

	resp := doRequest(t, srv, sessionRequest(t, http.MethodPost, "/api/auth/reauth", stale, `{"password":"geheim-123"}`))
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	claims, at := reauthCookieAuthTime(t, authSvc, resp)
	if claims.UserID != adminID {
		t.Fatalf("cookie issued for %s, want %s", claims.UserID, adminID)
	}
	if time.Since(at) > time.Minute {
		t.Fatalf("auth_time not fresh: %v", at)
	}
	if !authSvc.IsRememberMe(claims) {
		t.Errorf("remember-me choice of the session must be kept")
	}
	if findCookie(resp.Cookies(), csrfCookieName) == nil {
		t.Errorf("expected the CSRF cookie to be rotated")
	}

	// The reissued cookie opens the admin area again.
	expectReauthUser(mock, adminID, hash, true)
	expectMembershipExport(mock)
	cookie := findCookie(resp.Cookies(), authCookieName)
	resp = doRequest(t, srv, adminRequest(t, cookie.Value, reauthProbePath))
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 after reauth, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestReauthPassword_WrongPassword(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	adminID := uuid.NewString()
	hash, _ := authSvc.HashPassword("geheim-123")
	now := time.Now().UTC()
	stale := sessionToken(t, adminID, now.Add(-time.Hour), timePtr(now.Add(-time.Hour)))
	expectReauthUser(mock, adminID, hash, true)

	resp := doRequest(t, srv, sessionRequest(t, http.MethodPost, "/api/auth/reauth", stale, `{"password":"falsch"}`))
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if findCookie(resp.Cookies(), authCookieName) != nil {
		t.Fatalf("must not reissue the cookie")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestReauthPassword_Rejections(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		srv, mock, _ := newTestServer(t)
		resp := doRequest(t, srv, mustJSONRequest(t, http.MethodPost, "/api/auth/reauth", `{"password":"geheim-123"}`))
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unexpected queries: %v", err)
		}
	})
	t.Run("without CSRF", func(t *testing.T) {
		srv, mock, authSvc := newTestServer(t)
		token, _ := authSvc.GenerateToken(uuid.NewString(), false)
		req, _ := http.NewRequest(http.MethodPost, "/api/auth/reauth", nil)
		req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
		resp := doRequest(t, srv, req)
		resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", resp.StatusCode)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unexpected queries: %v", err)
		}
	})
	t.Run("API token", func(t *testing.T) {
		srv, mock, _ := newTestServer(t)
		req := mustJSONRequest(t, http.MethodPost, "/api/auth/reauth", `{"password":"geheim-123"}`)
		req.Header.Set("Authorization", "Bearer kgm_doesnotmatter")
		resp := doRequest(t, srv, req)
		resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", resp.StatusCode)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unexpected queries: %v", err)
		}
	})
	t.Run("empty password", func(t *testing.T) {
		srv, mock, authSvc := newTestServer(t)
		userID := uuid.NewString()
		token, _ := authSvc.GenerateToken(userID, false)
		expectReauthUser(mock, userID, "hash", false)
		resp := doRequest(t, srv, sessionRequest(t, http.MethodPost, "/api/auth/reauth", token, `{"password":"  "}`))
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet expectations: %v", err)
		}
	})
}

// --- Passkey re-authentication ---

const (
	reauthRPID   = "localhost"
	reauthOrigin = "http://localhost:5173"
)

type valueCapture struct{ value driver.Value }

func (c *valueCapture) Match(v driver.Value) bool {
	c.value = v
	return true
}

func b64url(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

// testPasskey is a minimal software passkey (ES256) for the assertion ceremony.
type testPasskey struct {
	key    *ecdsa.PrivateKey
	credID []byte
	owner  uuid.UUID
}

func newTestPasskey(t *testing.T, owner string) *testPasskey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	credID := make([]byte, 32)
	_, _ = rand.Read(credID)
	return &testPasskey{key: key, credID: credID, owner: uuid.MustParse(owner)}
}

const testPasskeyFlags = 0x01 | 0x04 // user present, user verified

// row is the stored webauthn_credentials row of this passkey.
func (p *testPasskey) row(t *testing.T, id string) *sqlmock.Rows {
	t.Helper()
	cose, err := webauthncbor.Marshal(map[int]any{
		1:  2,  // kty: EC2
		3:  -7, // alg: ES256
		-1: 1,  // crv: P-256
		-2: p.key.X.FillBytes(make([]byte, 32)),
		-3: p.key.Y.FillBytes(make([]byte, 32)),
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	return sqlmock.NewRows(passkeyColumns).AddRow(id, p.owner.String(), "Laptop", p.credID, cose, "none",
		make([]byte, 16), int64(0), "internal", int64(testPasskeyFlags), nil, now, now)
}

// assertion answers navigator.credentials.get for the given challenge.
func (p *testPasskey) assertion(t *testing.T, challenge string) string {
	t.Helper()
	rpIDHash := sha256.Sum256([]byte(reauthRPID))
	authData := append([]byte{}, rpIDHash[:]...)
	authData = append(authData, testPasskeyFlags)
	authData = binary.BigEndian.AppendUint32(authData, 0)
	clientData, _ := json.Marshal(map[string]any{"type": "webauthn.get", "challenge": challenge, "origin": reauthOrigin, "crossOrigin": false})
	clientHash := sha256.Sum256(clientData)
	digest := sha256.Sum256(append(append([]byte{}, authData...), clientHash[:]...))
	sig, err := ecdsa.SignASN1(rand.Reader, p.key, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{
		"credential": map[string]any{
			"id":    b64url(p.credID),
			"rawId": b64url(p.credID),
			"type":  "public-key",
			"response": map[string]any{
				"clientDataJSON":    b64url(clientData),
				"authenticatorData": b64url(authData),
				"signature":         b64url(sig),
				"userHandle":        b64url(p.owner[:]),
			},
			"clientExtensionResults": map[string]any{},
		},
	})
	return string(body)
}

// beginPasskeyCeremony runs POST /api/auth/passkeys/login/begin and returns
// the challenge, the ceremony ID and the stored session data.
func beginPasskeyCeremony(t *testing.T, srv *Server, mock sqlmock.Sqlmock) (string, string, driver.Value) {
	t.Helper()
	mock.ExpectExec(`DELETE FROM webauthn_sessions\s+WHERE expires_at`).WillReturnResult(sqlmock.NewResult(0, 0))
	data := &valueCapture{}
	mock.ExpectExec(`INSERT INTO webauthn_sessions`).
		WithArgs(sqlmock.AnyArg(), "login", nil, data, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	resp := doRequest(t, srv, mustJSONRequest(t, http.MethodPost, "/api/auth/passkeys/login/begin", ""))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("begin: expected 200, got %d", resp.StatusCode)
	}
	var options struct {
		PublicKey struct {
			Challenge string `json:"challenge"`
		} `json:"publicKey"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&options); err != nil {
		t.Fatalf("decode options: %v", err)
	}
	cookie := findCookie(resp.Cookies(), passkeyCeremonyCookie)
	if cookie == nil {
		t.Fatalf("expected ceremony cookie")
	}
	return options.PublicKey.Challenge, cookie.Value, data.value
}

func passkeyReauthRequest(t *testing.T, token, ceremonyID, body string) *http.Request {
	t.Helper()
	req := sessionRequest(t, http.MethodPost, "/api/auth/passkeys/reauth", token, body)
	if ceremonyID != "" {
		req.AddCookie(&http.Cookie{Name: passkeyCeremonyCookie, Value: ceremonyID})
	}
	return req
}

func expectPasskeyFinish(mock sqlmock.Sqlmock, t *testing.T, pk *testPasskey, ceremonyID string, data driver.Value, credRowID string) {
	now := time.Now().UTC()
	mock.ExpectQuery(`DELETE FROM webauthn_sessions`).WithArgs(ceremonyID, "login").
		WillReturnRows(sqlmock.NewRows([]string{"id", "ceremony", "user_id", "session_data", "expires_at", "created_at", "updated_at"}).
			AddRow(ceremonyID, "login", nil, data, now.Add(5*time.Minute), now, now))
	mock.ExpectQuery(`SELECT .* FROM webauthn_credentials\s+WHERE credential_id = \$1`).WithArgs(pk.credID).
		WillReturnRows(pk.row(t, credRowID))
	mock.ExpectExec(`UPDATE webauthn_credentials`).WithArgs(credRowID, int64(0), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

func TestReauthPasskey_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	adminID := uuid.NewString()
	pk := newTestPasskey(t, adminID)
	now := time.Now().UTC()
	stale := sessionToken(t, adminID, now.Add(-time.Hour), timePtr(now.Add(-time.Hour)))

	challenge, ceremonyID, data := beginPasskeyCeremony(t, srv, mock)
	expectReauthUser(mock, adminID, "hash", true)
	expectPasskeyFinish(mock, t, pk, ceremonyID, data, uuid.NewString())

	resp := doRequest(t, srv, passkeyReauthRequest(t, stale, ceremonyID, pk.assertion(t, challenge)))
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	claims, at := reauthCookieAuthTime(t, authSvc, resp)
	if claims.UserID != adminID || time.Since(at) > time.Minute {
		t.Fatalf("unexpected reissued session: uid=%s auth_time=%v", claims.UserID, at)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// A valid passkey of another account must not refresh the current session
// (nor switch it to the other account).
func TestReauthPasskey_PasskeyOfAnotherUser(t *testing.T) {
	srv, mock, _ := newTestServer(t)
	adminID := uuid.NewString()
	pk := newTestPasskey(t, uuid.NewString())
	now := time.Now().UTC()
	stale := sessionToken(t, adminID, now.Add(-time.Hour), timePtr(now.Add(-time.Hour)))

	challenge, ceremonyID, data := beginPasskeyCeremony(t, srv, mock)
	expectReauthUser(mock, adminID, "hash", true)
	expectPasskeyFinish(mock, t, pk, ceremonyID, data, uuid.NewString())

	resp := doRequest(t, srv, passkeyReauthRequest(t, stale, ceremonyID, pk.assertion(t, challenge)))
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if c := findCookie(resp.Cookies(), authCookieName); c != nil {
		t.Fatalf("must not reissue the auth cookie, got %+v", c)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestReauthPasskey_Rejections(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		srv, mock, _ := newTestServer(t)
		resp := doRequest(t, srv, passkeyReauthRequest(t, "", uuid.NewString(), `{"credential":{"id":"x"}}`))
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unexpected queries: %v", err)
		}
	})
	t.Run("API token", func(t *testing.T) {
		srv, mock, _ := newTestServer(t)
		req := passkeyReauthRequest(t, "", uuid.NewString(), `{"credential":{"id":"x"}}`)
		req.Header.Set("Authorization", "Bearer kgm_doesnotmatter")
		resp := doRequest(t, srv, req)
		resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", resp.StatusCode)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unexpected queries: %v", err)
		}
	})
	t.Run("without ceremony", func(t *testing.T) {
		srv, mock, authSvc := newTestServer(t)
		userID := uuid.NewString()
		token, _ := authSvc.GenerateToken(userID, false)
		expectReauthUser(mock, userID, "hash", true)
		resp := doRequest(t, srv, passkeyReauthRequest(t, token, "", `{"credential":{"id":"x"}}`))
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
		if findCookie(resp.Cookies(), authCookieName) != nil {
			t.Fatalf("must not reissue the cookie")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet expectations: %v", err)
		}
	})
}
