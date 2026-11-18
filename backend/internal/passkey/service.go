// Package passkey implements sign-in with passkeys (WebAuthn credentials).
//
// A ceremony runs in two steps: "begin" creates a challenge and stores it in
// webauthn_sessions, "finish" consumes that row (single use) and verifies the
// browser's response against it. Only discoverable credentials are created,
// so login needs no e-mail address: the authenticator returns the user handle.
package passkey

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"

	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
)

const (
	ceremonyRegistration = "registration"
	ceremonyLogin        = "login"

	// CeremonyTimeout is how long a started ceremony stays valid.
	CeremonyTimeout = 5 * time.Minute
)

var (
	// ErrNotFound is returned for unknown passkeys or passkeys of another user.
	ErrNotFound = errors.New("passkey not found")
	// ErrSessionInvalid means the ceremony was never started, already used or expired.
	ErrSessionInvalid = errors.New("passkey ceremony invalid or expired")
	// ErrVerification means the authenticator response did not verify.
	ErrVerification = errors.New("passkey verification failed")
	// ErrAlreadyRegistered means the credential is already stored (for any user).
	ErrAlreadyRegistered = errors.New("passkey already registered")
)

// Config describes the relying party (this application) for WebAuthn.
type Config struct {
	// RPID is the domain passkeys are bound to, e.g. "kegelmaster.example".
	RPID string
	// DisplayName is shown by some authenticators.
	DisplayName string
	// Origins are the exact origins (scheme://host[:port]) allowed to run ceremonies.
	Origins []string
}

// Credential is a stored passkey without its key material.
type Credential struct {
	ID             string
	UserID         string
	Name           string
	CredentialID   []byte
	Transports     []string
	BackupEligible bool
	BackupState    bool
	LastUsedAt     *time.Time
	CreatedAt      time.Time
}

// Service runs WebAuthn ceremonies and stores passkeys.
type Service struct {
	webauthn *webauthn.WebAuthn
	queries  *db.Queries
	nowFn    func() time.Time
}

// NewService validates cfg and returns a ready Service.
func NewService(dbConn *sql.DB, cfg Config) (*Service, error) {
	timeout := webauthn.TimeoutConfig{Enforce: true, Timeout: CeremonyTimeout, TimeoutUVD: CeremonyTimeout}
	wa, err := webauthn.New(&webauthn.Config{
		RPID:          cfg.RPID,
		RPDisplayName: cfg.DisplayName,
		RPOrigins:     cfg.Origins,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			ResidentKey:        protocol.ResidentKeyRequirementRequired,
			RequireResidentKey: protocol.ResidentKeyRequired(),
			UserVerification:   protocol.VerificationPreferred,
		},
		AttestationPreference: protocol.PreferNoAttestation,
		Timeouts:              webauthn.TimeoutsConfig{Login: timeout, Registration: timeout},
	})
	if err != nil {
		return nil, fmt.Errorf("configure webauthn: %w", err)
	}
	return &Service{webauthn: wa, queries: db.New(dbConn), nowFn: time.Now}, nil
}

// webauthnUser adapts a Kegelmaster user to webauthn.User.
type webauthnUser struct {
	id          string
	email       string
	credentials []webauthn.Credential
}

func (u webauthnUser) WebAuthnID() []byte {
	id := uuid.MustParse(u.id)
	return id[:]
}
func (u webauthnUser) WebAuthnName() string                       { return u.email }
func (u webauthnUser) WebAuthnDisplayName() string                { return u.email }
func (u webauthnUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

// BeginRegistration starts adding a passkey for the user. It returns the
// options for navigator.credentials.create and the ceremony ID that has to be
// passed to FinishRegistration.
func (s *Service) BeginRegistration(ctx context.Context, userID, email string) (*protocol.CredentialCreation, string, error) {
	rows, err := s.queries.ListWebAuthnCredentialsByUserID(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	user := webauthnUser{id: userID, email: email}
	exclusions := make([]protocol.CredentialDescriptor, 0, len(rows))
	for _, row := range rows {
		c := toWebAuthnCredential(row)
		exclusions = append(exclusions, c.Descriptor())
	}

	creation, session, err := s.webauthn.BeginRegistration(user, webauthn.WithExclusions(exclusions))
	if err != nil {
		return nil, "", err
	}
	sessionID, err := s.storeSession(ctx, ceremonyRegistration, &userID, session)
	if err != nil {
		return nil, "", err
	}
	return creation, sessionID, nil
}

// FinishRegistration verifies the response of navigator.credentials.create
// (the JSON of PublicKeyCredential.toJSON()) and stores the new passkey.
func (s *Service) FinishRegistration(ctx context.Context, sessionID, userID, email, name string, response []byte) (Credential, error) {
	session, err := s.consumeSession(ctx, sessionID, ceremonyRegistration)
	if err != nil {
		return Credential{}, err
	}
	if session.userID == nil || *session.userID != userID {
		return Credential{}, ErrSessionInvalid
	}

	parsed, err := protocol.ParseCredentialCreationResponseBytes(response)
	if err != nil {
		return Credential{}, fmt.Errorf("%w: %v", ErrVerification, err)
	}
	cred, err := s.webauthn.CreateCredential(webauthnUser{id: userID, email: email}, session.data, parsed)
	if err != nil {
		return Credential{}, fmt.Errorf("%w: %v", ErrVerification, err)
	}

	transports := make([]string, 0, len(cred.Transport))
	for _, t := range cred.Transport {
		transports = append(transports, string(t))
	}
	row, err := s.queries.CreateWebAuthnCredential(ctx, db.CreateWebAuthnCredentialParams{
		ID:              uuid.NewString(),
		UserID:          userID,
		Name:            name,
		CredentialID:    cred.ID,
		PublicKey:       cred.PublicKey,
		AttestationType: cred.AttestationType,
		Aaguid:          cred.Authenticator.AAGUID,
		SignCount:       int64(cred.Authenticator.SignCount),
		Transports:      strings.Join(transports, ","),
		Flags:           int16(cred.Flags.ProtocolValue()),
		CreatedAt:       s.nowFn().UTC(),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return Credential{}, ErrAlreadyRegistered
		}
		return Credential{}, err
	}
	return fromRow(row), nil
}

// BeginLogin starts a sign-in with any passkey of this site (no user known
// yet). It returns the options for navigator.credentials.get and the
// ceremony ID that has to be passed to FinishLogin.
func (s *Service) BeginLogin(ctx context.Context) (*protocol.CredentialAssertion, string, error) {
	assertion, session, err := s.webauthn.BeginDiscoverableLogin()
	if err != nil {
		return nil, "", err
	}
	sessionID, err := s.storeSession(ctx, ceremonyLogin, nil, session)
	if err != nil {
		return nil, "", err
	}
	return assertion, sessionID, nil
}

// FinishLogin verifies the response of navigator.credentials.get and returns
// the ID of the user the passkey belongs to.
func (s *Service) FinishLogin(ctx context.Context, sessionID string, response []byte) (string, Credential, error) {
	session, err := s.consumeSession(ctx, sessionID, ceremonyLogin)
	if err != nil {
		return "", Credential{}, err
	}

	parsed, err := protocol.ParseCredentialRequestResponseBytes(response)
	if err != nil {
		return "", Credential{}, fmt.Errorf("%w: %v", ErrVerification, err)
	}

	var stored db.WebauthnCredential
	lookup := func(rawID, userHandle []byte) (webauthn.User, error) {
		row, err := s.queries.GetWebAuthnCredentialByCredentialID(ctx, rawID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrNotFound
			}
			return nil, err
		}
		owner := uuid.MustParse(row.UserID)
		if len(userHandle) != len(owner) || string(userHandle) != string(owner[:]) {
			return nil, ErrNotFound
		}
		stored = row
		return webauthnUser{id: row.UserID, credentials: []webauthn.Credential{toWebAuthnCredential(row)}}, nil
	}

	_, cred, err := s.webauthn.ValidatePasskeyLogin(lookup, session.data, parsed)
	if err != nil {
		return "", Credential{}, fmt.Errorf("%w: %v", ErrVerification, err)
	}
	// A counter that did not increase means a second copy of the private key
	// may exist. Synced passkeys always report 0 and never trigger this.
	if cred.Authenticator.CloneWarning {
		return "", Credential{}, fmt.Errorf("%w: signature counter did not increase (possible cloned authenticator)", ErrVerification)
	}

	now := s.nowFn().UTC()
	if err := s.queries.UpdateWebAuthnCredentialAfterLogin(ctx, db.UpdateWebAuthnCredentialAfterLoginParams{
		ID:         stored.ID,
		SignCount:  int64(cred.Authenticator.SignCount),
		Flags:      int16(cred.Flags.ProtocolValue()),
		LastUsedAt: sql.NullTime{Time: now, Valid: true},
	}); err != nil {
		return "", Credential{}, err
	}
	c := fromRow(stored)
	c.LastUsedAt = &now
	return stored.UserID, c, nil
}

// ListByUserID returns the user's passkeys, newest first.
func (s *Service) ListByUserID(ctx context.Context, userID string) ([]Credential, error) {
	rows, err := s.queries.ListWebAuthnCredentialsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]Credential, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromRow(row))
	}
	return out, nil
}

// CountByUserID returns how many passkeys the user has.
func (s *Service) CountByUserID(ctx context.Context, userID string) (int64, error) {
	return s.queries.CountWebAuthnCredentialsByUserID(ctx, userID)
}

// Rename changes the display name of one of the user's passkeys.
func (s *Service) Rename(ctx context.Context, id, userID, name string) (Credential, error) {
	row, err := s.queries.RenameWebAuthnCredential(ctx, db.RenameWebAuthnCredentialParams{
		ID:        id,
		UserID:    userID,
		Name:      name,
		UpdatedAt: s.nowFn().UTC(),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return Credential{}, ErrNotFound
	}
	if err != nil {
		return Credential{}, err
	}
	return fromRow(row), nil
}

// Delete removes one of the user's passkeys.
func (s *Service) Delete(ctx context.Context, id, userID string) error {
	n, err := s.queries.DeleteWebAuthnCredential(ctx, db.DeleteWebAuthnCredentialParams{ID: id, UserID: userID})
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteByUserID removes all passkeys of the user.
func (s *Service) DeleteByUserID(ctx context.Context, userID string) error {
	return s.queries.DeleteWebAuthnCredentialsByUserID(ctx, userID)
}

type storedSession struct {
	userID *string
	data   webauthn.SessionData
}

func (s *Service) storeSession(ctx context.Context, ceremony string, userID *string, session *webauthn.SessionData) (string, error) {
	now := s.nowFn().UTC()
	// Housekeeping: abandoned ceremonies are removed whenever a new one starts.
	if err := s.queries.DeleteExpiredWebAuthnSessions(ctx, now); err != nil {
		return "", err
	}
	data, err := json.Marshal(session)
	if err != nil {
		return "", err
	}
	id := uuid.NewString()
	if err := s.queries.CreateWebAuthnSession(ctx, db.CreateWebAuthnSessionParams{
		ID:          id,
		Ceremony:    ceremony,
		UserID:      userID,
		SessionData: data,
		ExpiresAt:   now.Add(CeremonyTimeout),
		CreatedAt:   now,
	}); err != nil {
		return "", err
	}
	return id, nil
}

func (s *Service) consumeSession(ctx context.Context, id, ceremony string) (storedSession, error) {
	if _, err := uuid.Parse(id); err != nil {
		return storedSession{}, ErrSessionInvalid
	}
	row, err := s.queries.ConsumeWebAuthnSession(ctx, db.ConsumeWebAuthnSessionParams{ID: id, Ceremony: ceremony})
	if errors.Is(err, sql.ErrNoRows) {
		return storedSession{}, ErrSessionInvalid
	}
	if err != nil {
		return storedSession{}, err
	}
	if !s.nowFn().Before(row.ExpiresAt) {
		return storedSession{}, ErrSessionInvalid
	}
	var data webauthn.SessionData
	if err := json.Unmarshal(row.SessionData, &data); err != nil {
		return storedSession{}, err
	}
	return storedSession{userID: row.UserID, data: data}, nil
}

func toWebAuthnCredential(row db.WebauthnCredential) webauthn.Credential {
	return webauthn.Credential{
		ID:              row.CredentialID,
		PublicKey:       row.PublicKey,
		AttestationType: row.AttestationType,
		Transport:       splitTransports(row.Transports),
		Flags:           webauthn.NewCredentialFlags(protocol.AuthenticatorFlags(row.Flags)),
		Authenticator: webauthn.Authenticator{
			AAGUID:    row.Aaguid,
			SignCount: uint32(row.SignCount),
		},
	}
}

func splitTransports(s string) []protocol.AuthenticatorTransport {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]protocol.AuthenticatorTransport, 0, len(parts))
	for _, p := range parts {
		out = append(out, protocol.AuthenticatorTransport(p))
	}
	return out
}

func fromRow(row db.WebauthnCredential) Credential {
	flags := webauthn.NewCredentialFlags(protocol.AuthenticatorFlags(row.Flags))
	c := Credential{
		ID:             row.ID,
		UserID:         row.UserID,
		Name:           row.Name,
		CredentialID:   row.CredentialID,
		Transports:     []string{},
		BackupEligible: flags.BackupEligible,
		BackupState:    flags.BackupState,
		CreatedAt:      row.CreatedAt,
	}
	if row.Transports != "" {
		c.Transports = strings.Split(row.Transports, ",")
	}
	if row.LastUsedAt.Valid {
		t := row.LastUsedAt.Time
		c.LastUsedAt = &t
	}
	return c
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}
