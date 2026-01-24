package club

import (
	"time"
)

// Club represents a club entity in the system.
type Club struct {
	ID        string
	Name      string
	Balance   int // Cent-Betrag
	BaseFee   int // Cent-Betrag
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

