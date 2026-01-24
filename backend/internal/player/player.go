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
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

