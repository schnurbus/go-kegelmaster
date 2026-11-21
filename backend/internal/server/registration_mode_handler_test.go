package server

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/config"
	"github.com/schnurbus/go-kegelmaster/backend/internal/invitation"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
)

// newRegistrationModeServer returns a test server with the given
// REGISTRATION_MODE and an invitation repository backed by its own sqlmock.
func newRegistrationModeServer(t *testing.T, mode string) (*Server, sqlmock.Sqlmock, sqlmock.Sqlmock) {
	t.Helper()
	srv, mock, _ := newTestServer(t)
	srv.cfg.RegistrationMode = mode
	srv.handlers.Config.RegistrationMode = mode

	invDB, invMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create invitation sqlmock: %v", err)
	}
	t.Cleanup(func() { invDB.Close() })
	srv.handlers.InvitationRepo = invitation.NewRepository(invDB, player.NewRepository(invDB))
	return srv, mock, invMock
}

func expectUserInsert(mock sqlmock.Sqlmock, email string) {
	now := time.Now().UTC()
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), email, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at", "password_changed_at", "notify_gameday_settled", "is_instance_admin", "last_login_at", "disabled_at", "sessions_valid_after", "tour_seen_at", "setup_hint_seen_at"}).
			AddRow(uuid.NewString(), email, "hash", now, now, nil, false, false, nil, nil, nil, nil, nil))
}

func expectInvitationByToken(mock sqlmock.Sqlmock, token, email string, expiresAt time.Time, acceptedAt *time.Time) {
	rows := sqlmock.NewRows([]string{"id", "player_id", "email", "token", "expires_at", "accepted_at", "created_at"})
	var accepted sql.NullTime
	if acceptedAt != nil {
		accepted = sql.NullTime{Time: *acceptedAt, Valid: true}
	}
	rows.AddRow(uuid.NewString(), uuid.NewString(), email, token, expiresAt, accepted, time.Now().UTC())
	mock.ExpectQuery(`FROM player_invitations`).WithArgs(token).WillReturnRows(rows)
}

func registerStatus(t *testing.T, srv *Server, body string) (int, string) {
	t.Helper()
	resp := doRequest(t, srv, mustJSONRequest(t, http.MethodPost, "/api/auth/register", body))
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

func TestRegister_ModeOpen_DefaultWithoutToken(t *testing.T) {
	// Empty mode (variable not set) behaves like "open".
	for _, mode := range []string{"", config.RegistrationModeOpen} {
		srv, mock, invMock := newRegistrationModeServer(t, mode)
		expectUserInsert(mock, "open@example.com")

		status, body := registerStatus(t, srv, `{"email":"open@example.com","password":"super-safe-123"}`)
		if status != http.StatusCreated {
			t.Fatalf("mode %q: expected 201, got %d: %s", mode, status, body)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("mode %q: unmet expectations: %v", mode, err)
		}
		if err := invMock.ExpectationsWereMet(); err != nil {
			t.Fatalf("mode %q: unexpected invitation lookup: %v", mode, err)
		}
	}
}

func TestRegister_ModeOpen_IgnoresInviteToken(t *testing.T) {
	srv, mock, _ := newRegistrationModeServer(t, config.RegistrationModeOpen)
	expectUserInsert(mock, "open@example.com")

	status, body := registerStatus(t, srv, `{"email":"open@example.com","password":"super-safe-123","invite_token":"whatever"}`)
	if status != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", status, body)
	}
}

func TestRegister_ModeClosed(t *testing.T) {
	srv, mock, invMock := newRegistrationModeServer(t, config.RegistrationModeClosed)

	// Even a valid invitation token does not help; no DB access at all.
	status, body := registerStatus(t, srv, `{"email":"closed@example.com","password":"super-safe-123","invite_token":"tok"}`)
	if status != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", status, body)
	}
	if !strings.Contains(body, "deaktiviert") {
		t.Errorf("expected hint about disabled registration, got %s", body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB access: %v", err)
	}
	if err := invMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected invitation lookup: %v", err)
	}
}

func TestRegister_ModeInviteOnly(t *testing.T) {
	future := time.Now().UTC().Add(24 * time.Hour)
	past := time.Now().UTC().Add(-time.Hour)

	tests := []struct {
		name       string
		body       string
		setup      func(inv sqlmock.Sqlmock)
		wantStatus int
		wantUser   string // e-mail of the created user, "" if none
	}{
		{
			name:       "without token",
			body:       `{"email":"invited@example.com","password":"super-safe-123"}`,
			wantStatus: http.StatusForbidden,
		},
		{
			name: "unknown token",
			body: `{"email":"invited@example.com","password":"super-safe-123","invite_token":"unknown"}`,
			setup: func(inv sqlmock.Sqlmock) {
				inv.ExpectQuery(`FROM player_invitations`).WithArgs("unknown").WillReturnError(sql.ErrNoRows)
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "expired token",
			body: `{"email":"invited@example.com","password":"super-safe-123","invite_token":"expired"}`,
			setup: func(inv sqlmock.Sqlmock) {
				expectInvitationByToken(inv, "expired", "invited@example.com", past, nil)
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "already accepted token",
			body: `{"email":"invited@example.com","password":"super-safe-123","invite_token":"used"}`,
			setup: func(inv sqlmock.Sqlmock) {
				expectInvitationByToken(inv, "used", "invited@example.com", future, &past)
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "email does not match invitation",
			body: `{"email":"other@example.com","password":"super-safe-123","invite_token":"valid"}`,
			setup: func(inv sqlmock.Sqlmock) {
				expectInvitationByToken(inv, "valid", "invited@example.com", future, nil)
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "valid token and matching email (case-insensitive)",
			body: `{"email":"Invited@Example.com","password":"super-safe-123","invite_token":"valid"}`,
			setup: func(inv sqlmock.Sqlmock) {
				expectInvitationByToken(inv, "valid", "invited@example.com", future, nil)
			},
			wantStatus: http.StatusCreated,
			wantUser:   "invited@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, mock, invMock := newRegistrationModeServer(t, config.RegistrationModeInviteOnly)
			if tt.setup != nil {
				tt.setup(invMock)
			}
			if tt.wantUser != "" {
				expectUserInsert(mock, tt.wantUser)
			}

			status, body := registerStatus(t, srv, tt.body)
			if status != tt.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tt.wantStatus, status, body)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet user expectations: %v", err)
			}
			if err := invMock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet invitation expectations: %v", err)
			}
		})
	}
}

func TestPublicConfig_RegistrationMode(t *testing.T) {
	for _, tt := range []struct{ mode, want string }{
		{"", config.RegistrationModeOpen},
		{config.RegistrationModeInviteOnly, config.RegistrationModeInviteOnly},
		{config.RegistrationModeClosed, config.RegistrationModeClosed},
	} {
		srv, _, _ := newRegistrationModeServer(t, tt.mode)
		req, err := http.NewRequest(http.MethodGet, "/api/config/public", nil)
		if err != nil {
			t.Fatalf("new request: %v", err)
		}
		resp := doRequest(t, srv, req)
		var payload map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			t.Fatalf("decode: %v", err)
		}
		resp.Body.Close()
		if got := payload["registration_mode"]; got != tt.want {
			t.Errorf("mode %q: registration_mode = %v, want %s", tt.mode, got, tt.want)
		}
	}
}
