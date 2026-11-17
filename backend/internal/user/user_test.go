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
