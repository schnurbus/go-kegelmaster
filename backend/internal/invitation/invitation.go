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
	ErrNotFound        = errors.New("invitation not found")
	ErrExpired         = errors.New("invitation expired")
	ErrAlreadyAccepted = errors.New("invitation already accepted")
	// ErrPlayerAlreadyLinked is returned when the invited player already has a user.
	ErrPlayerAlreadyLinked = errors.New("player already linked to a user")
	// ErrUserAlreadyInClub is returned when the accepting user already has a player in the club.
	ErrUserAlreadyInClub = errors.New("user already has a player in this club")
)
