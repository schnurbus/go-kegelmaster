package player

import (
	"time"
)

// Player represents a player entity in the system.
type Player struct {
	ID           string
	ClubID       string
	UserID       *string // Optional user association
	RoleID       *string // Optional role assignment
	Name         string
	Balance      int // Cent-Betrag
	StartBalance int // Cent-Betrag
	Gender       *string // male, female, or nil (for evaluation of gender-specific competitions)
	Inactive     bool   // Inactive players do not pay base fee (Grundgebühr)
	PartnerID    *string // Optional partner (Paar-Modus); same club enforced in handler
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

