package server

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/passwordreset"
)

const passkeyCeremonyCookie = "passkey_ceremony"

var passkeyColumns = []string{"id", "user_id", "name", "credential_id", "public_key", "attestation_type", "aaguid", "sign_count", "transports", "flags", "last_used_at", "created_at", "updated_at"}

func sessionRequest(t *testing.T, method, path, token, body string) *http.Request {
	t.Helper()
	req := mustJSONRequest(t, method, path, body)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: authCookieName, Value: token})
	}
	return req
}

func TestPasskeyLoginBegin_ReturnsOptionsAndCeremonyCookie(t *testing.T) {
	srv, mock, _ := newTestServer(t)
	mock.ExpectExec(`DELETE FROM webauthn_sessions\s+WHERE expires_at`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO webauthn_sessions`).
		WithArgs(sqlmock.AnyArg(), "login", nil, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp := doRequest(t, srv, mustJSONRequest(t, http.MethodPost, "/api/auth/passkeys/login/begin", ""))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var payload struct {
		PublicKey struct {
			Challenge string `json:"challenge"`
			RPID      string `json:"rpId"`
		} `json:"publicKey"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.PublicKey.Challenge == "" || payload.PublicKey.RPID != "localhost" {
		t.Fatalf("unexpected options %+v", payload)
	}
	cookie := findCookie(resp.Cookies(), passkeyCeremonyCookie)
	if cookie == nil || !cookie.HttpOnly || cookie.Path != "/api/auth/passkeys" {
		t.Fatalf("expected http-only ceremony cookie, got %+v", cookie)
	}
	if _, err := uuid.Parse(cookie.Value); err != nil {
		t.Fatalf("ceremony cookie must carry the ceremony ID, got %q", cookie.Value)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPasskeyLoginBegin_RequiresCSRF(t *testing.T) {
	srv, mock, _ := newTestServer(t)
	req, _ := http.NewRequest(http.MethodPost, "/api/auth/passkeys/login/begin", nil)
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected queries: %v", err)
	}
}

func TestPasskeyLoginFinish_WithoutCeremony(t *testing.T) {
	srv, mock, _ := newTestServer(t)
	resp := doRequest(t, srv, mustJSONRequest(t, http.MethodPost, "/api/auth/passkeys/login/finish", `{"credential":{"id":"x"}}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if findCookie(resp.Cookies(), authCookieName) != nil {
		t.Fatalf("must not sign in")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected queries: %v", err)
	}
}

func TestPasskeyLoginFinish_ConsumedCeremony(t *testing.T) {
	srv, mock, _ := newTestServer(t)
	ceremonyID := uuid.NewString()
	mock.ExpectQuery(`DELETE FROM webauthn_sessions`).WithArgs(ceremonyID, "login").
		WillReturnRows(sqlmock.NewRows([]string{"id", "ceremony", "user_id", "session_data", "expires_at", "created_at", "updated_at"}))

	req := mustJSONRequest(t, http.MethodPost, "/api/auth/passkeys/login/finish", `{"credential":{"id":"x"}}`)
	req.AddCookie(&http.Cookie{Name: passkeyCeremonyCookie, Value: ceremonyID})
	resp := doRequest(t, srv, req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPasskeyRegisterBegin_RequiresSession(t *testing.T) {
	srv, mock, _ := newTestServer(t)
	resp := doRequest(t, srv, mustJSONRequest(t, http.MethodPost, "/api/auth/passkeys/register/begin", `{"current_password":"secret-123"}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected queries: %v", err)
	}
}

func TestPasskeyRegisterBegin_RejectsAPIToken(t *testing.T) {
	srv, mock, _ := newTestServer(t)
	resp := doRequest(t, srv, bearerRequest(http.MethodPost, "/api/auth/passkeys/register/begin", `{"current_password":"secret-123"}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected queries: %v", err)
	}
}

func TestPasskeyRegisterBegin_WrongPassword(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	hash, token := hashAndToken(t, authSvc, userID, "correct-password", false)
	expectUserWithHash(mock, userID, hash, time.Now().UTC().Add(-time.Hour))

	resp := doRequest(t, srv, sessionRequest(t, http.MethodPost, "/api/auth/passkeys/register/begin", token, `{"current_password":"wrong-password"}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if findCookie(resp.Cookies(), passkeyCeremonyCookie) != nil {
		t.Fatalf("no ceremony may start without the correct password")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPasskeyRegisterBegin_Success(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	hash, token := hashAndToken(t, authSvc, userID, "correct-password", false)
	expectUserWithHash(mock, userID, hash, time.Now().UTC().Add(-time.Hour))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM webauthn_credentials`).WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	now := time.Now().UTC()
	mock.ExpectQuery(`SELECT .* FROM webauthn_credentials\s+WHERE user_id`).WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(passkeyColumns).
			AddRow(uuid.NewString(), userID, "YubiKey", []byte{1, 2, 3}, []byte{4}, "none", nil, 7, "usb", 0x05, nil, now, now))
	mock.ExpectExec(`DELETE FROM webauthn_sessions\s+WHERE expires_at`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`INSERT INTO webauthn_sessions`).
		WithArgs(sqlmock.AnyArg(), "registration", userID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp := doRequest(t, srv, sessionRequest(t, http.MethodPost, "/api/auth/passkeys/register/begin", token, `{"current_password":"correct-password"}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var payload struct {
		PublicKey struct {
			User struct {
				Name string `json:"name"`
			} `json:"user"`
			ExcludeCredentials []struct {
				ID string `json:"id"`
			} `json:"excludeCredentials"`
		} `json:"publicKey"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.PublicKey.User.Name != "me@example.com" {
		t.Fatalf("unexpected user name %q", payload.PublicKey.User.Name)
	}
	if len(payload.PublicKey.ExcludeCredentials) != 1 || payload.PublicKey.ExcludeCredentials[0].ID != "AQID" {
		t.Fatalf("existing passkeys must be excluded, got %+v", payload.PublicKey.ExcludeCredentials)
	}
	if findCookie(resp.Cookies(), passkeyCeremonyCookie) == nil {
		t.Fatalf("expected ceremony cookie")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPasskeyList(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	token, _ := authSvc.GenerateToken(userID, false)
	expectUserByID(mock, userID, nil)
	now := time.Now().UTC()
	mock.ExpectQuery(`SELECT .* FROM webauthn_credentials\s+WHERE user_id`).WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(passkeyColumns).
			AddRow(uuid.NewString(), userID, "iPhone", []byte{1}, []byte{2}, "none", nil, 0, "internal,hybrid", 0x1d, now, now, now))

	resp := doRequest(t, srv, sessionRequest(t, http.MethodGet, "/api/auth/passkeys", token, ""))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var payload []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(payload) != 1 || payload[0]["name"] != "iPhone" || payload[0]["synced"] != true {
		t.Fatalf("unexpected payload %+v", payload)
	}
	for _, secret := range []string{"public_key", "credential_id"} {
		if _, ok := payload[0][secret]; ok {
			t.Fatalf("response must not contain %s", secret)
		}
	}
}

func TestPasskeyDelete_OtherUsersPasskeyNotFound(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID, passkeyID := uuid.NewString(), uuid.NewString()
	token, _ := authSvc.GenerateToken(userID, false)
	expectUserByID(mock, userID, nil)
	mock.ExpectExec(`DELETE FROM webauthn_credentials\s+WHERE id = \$1 AND user_id = \$2`).WithArgs(passkeyID, userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp := doRequest(t, srv, sessionRequest(t, http.MethodDelete, "/api/auth/passkeys/"+passkeyID, token, ""))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPasskeyRename_RequiresName(t *testing.T) {
	srv, mock, authSvc := newTestServer(t)
	userID := uuid.NewString()
	token, _ := authSvc.GenerateToken(userID, false)
	expectUserByID(mock, userID, nil)

	resp := doRequest(t, srv, sessionRequest(t, http.MethodPatch, "/api/auth/passkeys/"+uuid.NewString(), token, `{"name":"  "}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestResetPassword_DeletesPasskeys(t *testing.T) {
	srv, mock, _ := newTestServer(t)
	userID := uuid.NewString()
	now := time.Now().UTC()
	mock.ExpectQuery(`DELETE FROM password_reset_tokens`).WithArgs(passwordreset.HashToken("reset-token")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token", "expires_at", "created_at"}).
			AddRow(uuid.NewString(), userID, "hashed", now.Add(time.Hour), now))
	mock.ExpectExec(`UPDATE users`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM api_tokens`).WithArgs(userID).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`DELETE FROM webauthn_credentials\s+WHERE user_id = \$1`).WithArgs(userID).WillReturnResult(sqlmock.NewResult(0, 2))

	resp := doRequest(t, srv, mustJSONRequest(t, http.MethodPost, "/api/auth/reset-password", `{"token":"reset-token","new_password":"new-password-123"}`))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
