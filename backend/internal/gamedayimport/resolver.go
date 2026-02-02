package gamedayimport

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/competition"
	"github.com/schnurbus/go-kegelmaster/backend/internal/gameday"
	"github.com/schnurbus/go-kegelmaster/backend/internal/penaltytype"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
	"github.com/schnurbus/go-kegelmaster/backend/internal/transaction"
)

var (
	ErrClubNotFound     = errors.New("Club nicht gefunden")
	ErrClubAmbiguous    = errors.New("Club-Name ist mehrdeutig")
	ErrUnknownColumn   = errors.New("unbekannte Spalte (weder Strafentyp noch Wettbewerb)")
	ErrAmbiguousColumn  = errors.New("Spaltenname existiert sowohl als Strafentyp als auch als Wettbewerb")
	ErrUnknownPlayer    = errors.New("Spieler nicht im Club gefunden")
)

// Dependencies holds repositories required for resolve and apply.
// Resolve uses ClubRepo, PlayerRepo, PenaltyTypeRepo, CompetitionRepo.
// Apply uses all of them plus GameDayRepo, TransactionRepo, RoleRepo.
type Dependencies struct {
	ClubRepo         *club.Repository
	PlayerRepo       *player.Repository
	PenaltyTypeRepo  *penaltytype.Repository
	CompetitionRepo  *competition.Repository
	GameDayRepo      *gameday.Repository
	TransactionRepo  *transaction.Repository
	RoleRepo         *role.Repository
}

// Resolve resolves the parsed CSV against club data.
// clubName must match exactly one club. Every value column (header index 2+) must match
// exactly one penalty type or one competition (not both). Every player name in rows must
// exist in the club. Returns resolved input and a map player name -> player ID, or an error.
func Resolve(ctx context.Context, parsed *ParsedCSV, clubName string, deps *Dependencies) (*ResolvedInput, map[string]string, error) {
	clubs, err := deps.ClubRepo.GetAll(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("Clubs laden: %w", err)
	}

	var found *club.Club
	for i := range clubs {
		if clubs[i].Name == clubName {
			if found != nil {
				return nil, nil, ErrClubAmbiguous
			}
			c := clubs[i]
			found = &c
		}
	}
	if found == nil {
		return nil, nil, ErrClubNotFound
	}

	clubID := found.ID

	// Resolve value columns (header index 2+)
	columnNames := parsed.Header[2:]
	penaltyTypes, err := deps.PenaltyTypeRepo.GetByClubID(ctx, clubID)
	if err != nil {
		return nil, nil, fmt.Errorf("Strafentypen laden: %w", err)
	}
	competitions, err := deps.CompetitionRepo.GetByClubID(ctx, clubID)
	if err != nil {
		return nil, nil, fmt.Errorf("Wettbewerbe laden: %w", err)
	}

	penaltyByName := make(map[string]penaltytype.PenaltyType)
	for _, pt := range penaltyTypes {
		if pt.DeletedAt != nil {
			continue
		}
		penaltyByName[pt.Name] = pt
	}
	competitionByName := make(map[string]competition.Competition)
	for _, c := range competitions {
		competitionByName[c.Name] = c
	}

	mapping := make([]ColumnMapping, 0, len(columnNames))
	for _, colName := range columnNames {
		colName = strings.TrimSpace(colName)
		inPenalty := penaltyByName[colName]
		inCompetition := competitionByName[colName]
		hasPenalty := inPenalty.ID != ""
		hasCompetition := inCompetition.ID != ""

		switch {
		case hasPenalty && hasCompetition:
			return nil, nil, fmt.Errorf("%w: %q", ErrAmbiguousColumn, colName)
		case hasPenalty:
			mapping = append(mapping, ColumnMapping{
				ColumnName:             colName,
				Kind:                   ColumnKindPenalty,
				PenaltyTypeID:          inPenalty.ID,
				PenaltyTypeName:        inPenalty.Name,
				PenaltyTypeDescription: inPenalty.Description,
				PenaltyTypePrice:       inPenalty.Price,
				AllowsDecimalQuantity:  inPenalty.AllowsDecimalQuantity,
			})
		case hasCompetition:
			mapping = append(mapping, ColumnMapping{
				ColumnName:   colName,
				Kind:         ColumnKindCompetition,
				CompetitionID: inCompetition.ID,
			})
		default:
			return nil, nil, fmt.Errorf("%w: %q", ErrUnknownColumn, colName)
		}
	}

	// Resolve player names
	players, err := deps.PlayerRepo.GetByClubID(ctx, clubID)
	if err != nil {
		return nil, nil, fmt.Errorf("Spieler laden: %w", err)
	}
	playerByName := make(map[string]string)
	for _, p := range players {
		playerByName[p.Name] = p.ID
	}

	seenPlayers := make(map[string]struct{})
	for _, row := range parsed.Rows {
		seenPlayers[row.PlayerName] = struct{}{}
	}
	for name := range seenPlayers {
		if _, ok := playerByName[name]; !ok {
			return nil, nil, fmt.Errorf("%w: %q", ErrUnknownPlayer, name)
		}
	}

	return &ResolvedInput{
		ClubID:        clubID,
		ClubName:     found.Name,
		ColumnMapping: mapping,
		ColumnNames:   columnNames,
		Rows:         parsed.Rows,
	}, playerByName, nil
}
