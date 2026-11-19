package user

import (
	"testing"
	"time"
)

func TestSessionValid(t *testing.T) {
	changed := time.Date(2026, 1, 10, 12, 0, 30, 700_000_000, time.UTC)
	u := User{PasswordChangedAt: &changed}

	tests := []struct {
		name     string
		issuedAt time.Time
		want     bool
	}{
		{"issued before change", changed.Add(-time.Minute), false},
		{"issued in the same second (JWT precision)", changed.Truncate(time.Second), true},
		{"issued after change", changed.Add(time.Minute), true},
	}
	for _, tt := range tests {
		if got := u.SessionValid(tt.issuedAt); got != tt.want {
			t.Errorf("%s: SessionValid = %v, want %v", tt.name, got, tt.want)
		}
	}

	if !(User{}).SessionValid(time.Unix(0, 0)) {
		t.Error("user without password change must accept every session")
	}
}

func TestSessionValidRevokedByAdmin(t *testing.T) {
	changed := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	revoked := time.Date(2026, 2, 1, 8, 0, 15, 300_000_000, time.UTC)
	u := User{PasswordChangedAt: &changed, SessionsValidAfter: &revoked}

	tests := []struct {
		name     string
		issuedAt time.Time
		want     bool
	}{
		{"issued before password change", changed.Add(-time.Hour), false},
		{"issued between password change and revocation", revoked.Add(-time.Hour), false},
		{"issued in the revocation second (JWT precision)", revoked.Truncate(time.Second), true},
		{"issued after revocation", revoked.Add(time.Minute), true},
	}
	for _, tt := range tests {
		if got := u.SessionValid(tt.issuedAt); got != tt.want {
			t.Errorf("%s: SessionValid = %v, want %v", tt.name, got, tt.want)
		}
	}

	onlyRevoked := User{SessionsValidAfter: &revoked}
	if onlyRevoked.SessionValid(revoked.Add(-time.Second)) {
		t.Error("session issued before revocation must be rejected without password change")
	}
}

func TestDisabled(t *testing.T) {
	if (User{}).Disabled() {
		t.Error("user without disabled_at must not be disabled")
	}
	now := time.Now()
	if !(User{DisabledAt: &now}).Disabled() {
		t.Error("user with disabled_at must be disabled")
	}
}
