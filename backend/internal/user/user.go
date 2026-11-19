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
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// SessionValid reports whether a session issued at issuedAt is still valid,
// i.e. it was not issued before the last password change. JWT timestamps
// have second precision, so the comparison truncates to seconds.
func (u User) SessionValid(issuedAt time.Time) bool {
	if u.PasswordChangedAt == nil {
		return true
	}
	return !issuedAt.Truncate(time.Second).Before(u.PasswordChangedAt.Truncate(time.Second))
}
