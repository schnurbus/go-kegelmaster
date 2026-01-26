package invitation

import (
	"errors"
	"time"
)

// PlayerInvitation represents a player invitation entity.
type PlayerInvitation struct {
	ID         string
	PlayerID   string
	Email      string
	Token      string
	ExpiresAt  time.Time
	AcceptedAt *time.Time
	CreatedAt  time.Time
}

var (
	ErrNotFound       = errors.New("invitation not found")
	ErrExpired        = errors.New("invitation expired")
	ErrAlreadyAccepted = errors.New("invitation already accepted")
)
