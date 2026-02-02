package competition

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("competition not found")
)

// Competition represents a competition entity in the system.
type Competition struct {
	ID               string
	ClubID           string
	Name             string
	ScoringType      string // winner, loser, both
	IsGenderSpecific bool
	DisplayOrder     int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
