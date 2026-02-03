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
}
