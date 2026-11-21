package club

import (
	"time"
)

// Club represents a club entity in the system.
type Club struct {
	ID                 string
	Name               string
	Balance            int  // Cent-Betrag
	StartBalance       int  // Cent-Betrag; Basis für Neuberechnung aus Transaktionen
	BaseFee            int  // Cent-Betrag
	AutoTipEnabled     bool // Auto-tip feature enabled
	CouplesModeEnabled bool // Paar-Modus: Einzahlung auf mehrere Spieler verteilen
	UserID             string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	// SetupDismissedAt is set once the owner hid the Ersteinrichtung card. It
	// belongs to the club and therefore survives an ownership transfer.
	SetupDismissedAt *time.Time
}

// SetupCounts is the data the Ersteinrichtung of a club is derived from.
type SetupCounts struct {
	PenaltyTypes      int64
	Roles             int64
	Players           int64
	PublishedGameDays int64
	Invitations       int64
	Competitions      int64
	// OwnPlayers counts the players linked to the requested user.
	OwnPlayers int64
}
