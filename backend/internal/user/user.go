package user

import (
	"time"
)

// User represents the persisted state of an application user.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	// PasswordChangedAt is set when the password was reset or changed.
	// Sessions issued before this point are rejected.
	PasswordChangedAt *time.Time
	// NotifyGameDaySettled enables an e-mail once a game day of one of the
	// user's clubs was settled (published). Off by default.
	NotifyGameDaySettled bool
	// IsInstanceAdmin grants access to the read-only operator area (/api/admin).
	// It is independent of club permissions and only set via CLI or
	// INSTANCE_ADMIN_EMAILS, never through the web UI.
	IsInstanceAdmin bool
	// LastLoginAt is the last successful password or passkey login (nil = unknown).
	LastLoginAt *time.Time
	// DisabledAt is set when an instance admin disabled the account. A disabled
	// account cannot log in; its sessions and API tokens are rejected.
	DisabledAt *time.Time
	// SessionsValidAfter is set when an instance admin revoked the sessions.
	// Sessions issued before this point are rejected (like PasswordChangedAt).
	SessionsValidAfter *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Disabled reports whether an instance admin disabled the account.
func (u User) Disabled() bool {
	return u.DisabledAt != nil
}

// SessionValid reports whether a session issued at issuedAt is still valid,
// i.e. it was issued neither before the last password change nor before the
// last session revocation by an instance admin. JWT timestamps have second
// precision, so the comparison truncates to seconds.
func (u User) SessionValid(issuedAt time.Time) bool {
	issued := issuedAt.Truncate(time.Second)
	for _, after := range []*time.Time{u.PasswordChangedAt, u.SessionsValidAfter} {
		if after != nil && issued.Before(after.Truncate(time.Second)) {
			return false
		}
	}
	return true
}
