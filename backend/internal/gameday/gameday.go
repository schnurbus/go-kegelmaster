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
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GameDayParticipant links a player to a game day
type GameDayParticipant struct {
	ID         string
	GameDayID  string
	PlayerID   string
	PlayerName string // Joined from players table
	CreatedAt  time.Time
}

// GameDayFee records a penalty fee with snapshot of penalty type info
type GameDayFee struct {
	ID                     string
	GameDayParticipantID   string
	PenaltyTypeID          string
	PenaltyTypeName        string // Snapshot
	PenaltyTypeDescription string // Snapshot
	PenaltyTypePrice       int    // Snapshot (Cent)
	Count                  int
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// GameDayDetail combines game day with participants and their fees
type GameDayDetail struct {
	GameDay      GameDay
	Participants []ParticipantWithFees
}

// ParticipantWithFees combines participant info with their fees
type ParticipantWithFees struct {
	Participant GameDayParticipant
	Fees        []GameDayFee
}

// GameDaySummary represents a game day with aggregated statistics
type GameDaySummary struct {
	ID               string
	ClubID           string
	Date             time.Time
	Notes            string
	ParticipantCount int
	PenaltyFeeTotal  int // in cents
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
