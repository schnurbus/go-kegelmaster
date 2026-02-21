package gameday

import (
	"errors"
	"time"
)

var (
	ErrNotFound            = errors.New("game day not found")
	ErrParticipantNotFound = errors.New("participant not found")
	ErrFeeNotFound         = errors.New("fee not found")
)

// GameDay represents a match day for a club
type GameDay struct {
	ID        string
	ClubID    string
	Date      time.Time
	Notes     string
	IsDraft   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GameDayParticipant links a player to a game day
type GameDayParticipant struct {
	ID           string
	GameDayID    string
	PlayerID     string
	PlayerName   string  // Joined from players table
	PlayerGender *string // male, female, or nil (for gender-specific competition evaluation)
	CreatedAt    time.Time
}

// GameDayFee records a penalty fee with snapshot of penalty type info
type GameDayFee struct {
	ID                     string
	GameDayParticipantID   string
	PenaltyTypeID          string
	PenaltyTypeName        string // Snapshot
	PenaltyTypeDescription string // Snapshot
	PenaltyTypePrice       int    // Snapshot (Cent)
	Count                  int    // scaled when QuantityScale > 1 (e.g. 250 = 2.5)
	QuantityScale          int    // 1 = integer, 100 = 2 decimal places
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// GameDayCompetitionValue records a competition value per participant
type GameDayCompetitionValue struct {
	ID                   string
	GameDayParticipantID string
	CompetitionID        string
	Value                int
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// GameDayDetail combines game day with participants, their fees and competition values
type GameDayDetail struct {
	GameDay      GameDay
	Participants []ParticipantWithFees
}

// ParticipantWithFees combines participant info with their fees and competition values
type ParticipantWithFees struct {
	Participant      GameDayParticipant
	Fees             []GameDayFee
	CompetitionValues []GameDayCompetitionValue
}

// GameDaySummary represents a game day with aggregated statistics
type GameDaySummary struct {
	ID               string
	ClubID           string
	Date             time.Time
	Notes            string
	IsDraft          bool
	ParticipantCount int
	PenaltyFeeTotal  int // in cents
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// PenaltyHistoryDay is one game day in a player's penalty history (for dashboard chart).
type PenaltyHistoryDay struct {
	GameDayID string
	Date      time.Time
	Penalties []PenaltyHistoryPenalty
}

// PenaltyHistoryPenalty is one penalty type's count for a game day (display quantity).
type PenaltyHistoryPenalty struct {
	PenaltyTypeID   string
	PenaltyTypeName string
	Quantity        float64 // display count (count/quantity_scale when scale > 1)
}

// CompetitionHistoryDay is one game day in a player's competition history (for dashboard chart).
type CompetitionHistoryDay struct {
	GameDayID string
	Date      time.Time
	Values    []CompetitionHistoryValue
}

// CompetitionHistoryValue is one competition's value for a game day.
type CompetitionHistoryValue struct {
	CompetitionID   string
	CompetitionName string
	Value           int
}
