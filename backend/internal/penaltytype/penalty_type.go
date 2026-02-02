package penaltytype

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("penalty type not found")
)

// PenaltyType represents a penalty type entity in the system.
type PenaltyType struct {
	ID                    string
	ClubID                string
	Name                  string
	Description           string
	Price                 int  // Cent-Betrag
	DisplayOrder          int
	AllowsDecimalQuantity bool // when true, fee count can be decimal (stored as count*100)
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             *time.Time
	ReplacedByID          *string
}
