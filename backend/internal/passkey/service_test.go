package passkey

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/protocol/webauthncbor"
	"github.com/google/uuid"
)

const (
	testRPID   = "kegel.test"
	testOrigin = "https://kegel.test"
)

var (
	credentialColumns = []string{"id", "user_id", "name", "credential_id", "public_key", "attestation_type", "aaguid", "sign_count", "transports", "flags", "last_used_at", "created_at", "updated_at"}
	sessionColumns    = []string{"id", "ceremony", "user_id", "session_data", "expires_at", "created_at", "updated_at"}
)

// capture is a sqlmock argument matcher that records the value it sees.
type capture struct{ value driver.Value }

func (c *capture) Match(v driver.Value) bool {
	c.value = v
	return true
}

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

// softAuthenticator is a minimal software passkey (ES256, "none" attestation).
type softAuthenticator struct {
	t        *testing.T
	key      *ecdsa.PrivateKey
	credID   []byte
	userID   []byte
	counter  uint32
	origin   string
	flags    byte
	rpIDHash [32]byte
}

func newSoftAuthenticator(t *testing.T) *softAuthenticator {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	credID := make([]byte, 32)
	_, _ = rand.Read(credID)
	return &softAuthenticator{
		t:        t,
		key:      key,
		credID:   credID,
		origin:   testOrigin,
		flags:    0x01 | 0x04 | 0x08 | 0x10, // UP, UV, BE, BS (a synced passkey)
		rpIDHash: sha256.Sum256([]byte(testRPID)),
	}
}

func (a *softAuthenticator) clientData(typ, challenge string) []byte {
	data, _ := json.Marshal(map[string]any{"type": typ, "challenge": challenge, "origin": a.origin, "crossOrigin": false})
	return data
}

func (a *softAuthenticator) authData(attested bool) []byte {
	flags := a.flags
	if attested {
		flags |= 0x40
	}
	out := append([]byte{}, a.rpIDHash[:]...)
	out = append(out, flags)
	out = binary.BigEndian.AppendUint32(out, a.counter)
	if attested {
		out = append(out, make([]byte, 16)...) // AAGUID
		out = binary.BigEndian.AppendUint16(out, uint16(len(a.credID)))
		out = append(out, a.credID...)
		cose, err := webauthncbor.Marshal(map[int]any{
			1:  2,  // kty: EC2
			3:  -7, // alg: ES256
			-1: 1,  // crv: P-256
			-2: a.key.X.FillBytes(make([]byte, 32)),
			-3: a.key.Y.FillBytes(make([]byte, 32)),
		})
		if err != nil {
			a.t.Fatal(err)
		}
		out = append(out, cose...)
	}
	return out
}

// create answers navigator.credentials.create.
func (a *softAuthenticator) create(options *protocol.CredentialCreation) []byte {
	a.userID = options.Response.User.ID.(protocol.URLEncodedBase64)
	attObj, err := webauthncbor.Marshal(map[string]any{"fmt": "none", "attStmt": map[string]any{}, "authData": a.authData(true)})
	if err != nil {
		a.t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{
		"id":    b64(a.credID),
		"rawId": b64(a.credID),
		"type":  "public-key",
		"response": map[string]any{
			"clientDataJSON":    b64(a.clientData("webauthn.create", options.Response.Challenge.String())),
			"attestationObject": b64(attObj),
			"transports":        []string{"internal", "hybrid"},
		},
		"clientExtensionResults": map[string]any{},
	})
	return body
}

// get answers navigator.credentials.get.
func (a *softAuthenticator) get(options *protocol.CredentialAssertion) []byte {
	authData := a.authData(false)
	clientData := a.clientData("webauthn.get", options.Response.Challenge.String())
	hash := sha256.Sum256(clientData)
	digest := sha256.Sum256(append(append([]byte{}, authData...), hash[:]...))
	sig, err := ecdsa.SignASN1(rand.Reader, a.key, digest[:])
	if err != nil {
		a.t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{
		"id":    b64(a.credID),
		"rawId": b64(a.credID),
		"type":  "public-key",
		"response": map[string]any{
			"clientDataJSON":    b64(clientData),
			"authenticatorData": b64(authData),
			"signature":         b64(sig),
			"userHandle":        b64(a.userID),
		},
		"clientExtensionResults": map[string]any{},
	})
	return body
}

func newTestService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	t.Helper()
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	svc, err := NewService(conn, Config{RPID: testRPID, DisplayName: "Kegelmaster", Origins: []string{testOrigin}})
	if err != nil {
		t.Fatal(err)
	}
	return svc, mock
}

// storedCredential is what CreateWebAuthnCredential wrote, replayed as a row.
type storedCredential struct {
	id, userID string
	credID     []byte
	publicKey  []byte
	signCount  int64
	transports string
	flags      int64
}

func (c storedCredential) rows() *sqlmock.Rows {
	now := time.Now().UTC()
	return sqlmock.NewRows(credentialColumns).AddRow(c.id, c.userID, "iPhone", c.credID, c.publicKey, "none", make([]byte, 16), c.signCount, c.transports, c.flags, nil, now, now)
}

func sessionRow(id, ceremony string, userID *string, data driver.Value) *sqlmock.Rows {
	now := time.Now().UTC()
	var uid driver.Value
	if userID != nil {
		uid = *userID
	}
	return sqlmock.NewRows(sessionColumns).AddRow(id, ceremony, uid, data, now.Add(CeremonyTimeout), now, now)
}

// register runs a full registration ceremony and returns the stored credential.
func register(t *testing.T, svc *Service, mock sqlmock.Sqlmock, auth *softAuthenticator, userID string) storedCredential {
	t.Helper()
	ctx := t.Context()

	mock.ExpectQuery(`SELECT .* FROM webauthn_credentials\s+WHERE user_id = \$1`).WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(credentialColumns))
	mock.ExpectExec(`DELETE FROM webauthn_sessions\s+WHERE expires_at`).WillReturnResult(sqlmock.NewResult(0, 0))
	sessionData := &capture{}
	mock.ExpectExec(`INSERT INTO webauthn_sessions`).
		WithArgs(sqlmock.AnyArg(), "registration", userID, sessionData, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	options, ceremonyID, err := svc.BeginRegistration(ctx, userID, "kegler@example.com")
	if err != nil {
		t.Fatalf("begin registration: %v", err)
	}
	if got := options.Response.AuthenticatorSelection.ResidentKey; got != protocol.ResidentKeyRequirementRequired {
		t.Fatalf("expected discoverable credential to be required, got %q", got)
	}

	mock.ExpectQuery(`DELETE FROM webauthn_sessions\s+WHERE id = \$1 AND ceremony = \$2`).WithArgs(ceremonyID, "registration").
		WillReturnRows(sessionRow(ceremonyID, "registration", &userID, sessionData.value))
	publicKey, signCount, transports, flags := &capture{}, &capture{}, &capture{}, &capture{}
	mock.ExpectQuery(`INSERT INTO webauthn_credentials`).
		WithArgs(sqlmock.AnyArg(), userID, "iPhone", auth.credID, publicKey, "none", sqlmock.AnyArg(), signCount, transports, flags, sqlmock.AnyArg()).
		WillReturnRows(storedCredential{id: uuid.NewString(), userID: userID, credID: auth.credID}.rows())

	if _, err := svc.FinishRegistration(ctx, ceremonyID, userID, "kegler@example.com", "iPhone", auth.create(options)); err != nil {
		t.Fatalf("finish registration: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
	return storedCredential{
		id:         uuid.NewString(),
		userID:     userID,
		credID:     auth.credID,
		publicKey:  publicKey.value.([]byte),
		signCount:  signCount.value.(int64),
		transports: transports.value.(string),
		flags:      flags.value.(int64),
	}
}

// beginLogin starts a login ceremony and returns its options, ID and stored session data.
func beginLogin(t *testing.T, svc *Service, mock sqlmock.Sqlmock) (*protocol.CredentialAssertion, string, driver.Value) {
	t.Helper()
	mock.ExpectExec(`DELETE FROM webauthn_sessions\s+WHERE expires_at`).WillReturnResult(sqlmock.NewResult(0, 0))
	sessionData := &capture{}
	mock.ExpectExec(`INSERT INTO webauthn_sessions`).
		WithArgs(sqlmock.AnyArg(), "login", nil, sessionData, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	options, ceremonyID, err := svc.BeginLogin(t.Context())
	if err != nil {
		t.Fatalf("begin login: %v", err)
	}
	return options, ceremonyID, sessionData.value
}

func TestRegisterAndLogin(t *testing.T) {
	svc, mock := newTestService(t)
	auth := newSoftAuthenticator(t)
	userID := uuid.NewString()

	stored := register(t, svc, mock, auth, userID)
	if stored.transports != "internal,hybrid" {
		t.Fatalf("unexpected transports %q", stored.transports)
	}

	options, ceremonyID, data := beginLogin(t, svc, mock)
	if len(options.Response.AllowedCredentials) != 0 {
		t.Fatalf("discoverable login must not list credentials")
	}
	auth.counter = 1
	mock.ExpectQuery(`DELETE FROM webauthn_sessions`).WithArgs(ceremonyID, "login").WillReturnRows(sessionRow(ceremonyID, "login", nil, data))
	mock.ExpectQuery(`SELECT .* FROM webauthn_credentials\s+WHERE credential_id = \$1`).WithArgs(auth.credID).WillReturnRows(stored.rows())
	mock.ExpectExec(`UPDATE webauthn_credentials`).WithArgs(stored.id, int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	gotUser, cred, err := svc.FinishLogin(t.Context(), ceremonyID, auth.get(options))
	if err != nil {
		t.Fatalf("finish login: %v", err)
	}
	if gotUser != userID {
		t.Fatalf("expected user %s, got %s", userID, gotUser)
	}
	if !cred.BackupState || cred.LastUsedAt == nil {
		t.Fatalf("unexpected credential %+v", cred)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestFinishLoginRejections(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(auth *softAuthenticator, stored *storedCredential)
	}{
		{
			name:   "wrong origin",
			mutate: func(a *softAuthenticator, _ *storedCredential) { a.origin = "https://evil.test" },
		},
		{
			name:   "user handle of another user",
			mutate: func(a *softAuthenticator, _ *storedCredential) { id := uuid.New(); a.userID = id[:] },
		},
		{
			name: "signature counter went backwards",
			mutate: func(a *softAuthenticator, s *storedCredential) {
				s.signCount = 10
				a.counter = 5
			},
		},
		{
			name: "signed by another key",
			mutate: func(a *softAuthenticator, _ *storedCredential) {
				other, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				a.key = other
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, mock := newTestService(t)
			auth := newSoftAuthenticator(t)
			stored := register(t, svc, mock, auth, uuid.NewString())

			options, ceremonyID, data := beginLogin(t, svc, mock)
			tc.mutate(auth, &stored)
			mock.ExpectQuery(`DELETE FROM webauthn_sessions`).WithArgs(ceremonyID, "login").WillReturnRows(sessionRow(ceremonyID, "login", nil, data))
			mock.ExpectQuery(`SELECT .* FROM webauthn_credentials`).WillReturnRows(stored.rows())

			_, _, err := svc.FinishLogin(t.Context(), ceremonyID, auth.get(options))
			t.Logf("rejected: %v", err)
			if !errors.Is(err, ErrVerification) {
				t.Fatalf("expected ErrVerification, got %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations (credential must not be updated): %v", err)
			}
		})
	}
}

func TestFinishLoginUnknownOrUsedCeremony(t *testing.T) {
	svc, mock := newTestService(t)
	id := uuid.NewString()
	mock.ExpectQuery(`DELETE FROM webauthn_sessions`).WithArgs(id, "login").WillReturnRows(sqlmock.NewRows(sessionColumns))

	if _, _, err := svc.FinishLogin(t.Context(), id, []byte(`{}`)); !errors.Is(err, ErrSessionInvalid) {
		t.Fatalf("expected ErrSessionInvalid, got %v", err)
	}
	// Malformed IDs never reach the database.
	if _, _, err := svc.FinishLogin(t.Context(), "", []byte(`{}`)); !errors.Is(err, ErrSessionInvalid) {
		t.Fatalf("expected ErrSessionInvalid, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFinishRegistrationRejectsCeremonyOfAnotherUser(t *testing.T) {
	svc, mock := newTestService(t)
	id, owner := uuid.NewString(), uuid.NewString()
	mock.ExpectQuery(`DELETE FROM webauthn_sessions`).WithArgs(id, "registration").
		WillReturnRows(sessionRow(id, "registration", &owner, []byte(`{"challenge":"abc"}`)))

	_, err := svc.FinishRegistration(t.Context(), id, uuid.NewString(), "x@example.com", "Key", []byte(`{}`))
	if !errors.Is(err, ErrSessionInvalid) {
		t.Fatalf("expected ErrSessionInvalid, got %v", err)
	}
}
