package statistics

import (
	"context"
	"database/sql"
	"time"

	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
)

// earliestDate is used as lower bound when no start date is given ("Gesamt").
var earliestDate = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

// Repository loads the raw data for player statistics.
type Repository struct {
	queries *db.Queries
}

func NewRepository(dbConn *sql.DB) *Repository {
	return &Repository{queries: db.New(dbConn)}
}

// PlayerParams identifies the player and period for which statistics are computed.
type PlayerParams struct {
	ClubID       string
	PlayerID     string
	PlayerName   string
	PlayerGender string
	StartBalance int
	From         *time.Time // nil = since the beginning
	To           time.Time
}

// ForPlayer loads all required data and computes the player's statistics.
func (r *Repository) ForPlayer(ctx context.Context, p PlayerParams) (PlayerStatistics, error) {
	in, err := r.Load(ctx, p)
	if err != nil {
		return PlayerStatistics{}, err
	}
	return Compute(in), nil
}

// Load reads the raw data for the given player and period.
func (r *Repository) Load(ctx context.Context, p PlayerParams) (Input, error) {
	from := earliestDate
	if p.From != nil {
		from = *p.From
	}
	in := Input{
		PlayerID:           p.PlayerID,
		PlayerName:         p.PlayerName,
		PlayerGender:       p.PlayerGender,
		PlayerStartBalance: p.StartBalance,
		From:               p.From,
		To:                 p.To,
	}

	gameDays, err := r.queries.StatsListGameDays(ctx, db.StatsListGameDaysParams{ClubID: p.ClubID, Date: from, Date_2: p.To})
	if err != nil {
		return Input{}, err
	}
	for _, gd := range gameDays {
		in.GameDays = append(in.GameDays, GameDayRef{ID: gd.ID, Date: gd.Date})
	}

	participants, err := r.queries.StatsListParticipants(ctx, db.StatsListParticipantsParams{ClubID: p.ClubID, Date: from, Date_2: p.To})
	if err != nil {
		return Input{}, err
	}
	for _, row := range participants {
		in.Participants = append(in.Participants, Participation{
			ParticipantID: row.ParticipantID,
			GameDayID:     row.GameDayID,
			PlayerID:      row.PlayerID,
			Attended:      row.Attended,
			Gender:        row.Gender.String,
		})
	}

	fees, err := r.queries.StatsListFees(ctx, db.StatsListFeesParams{ClubID: p.ClubID, Date: from, Date_2: p.To})
	if err != nil {
		return Input{}, err
	}
	for _, row := range fees {
		in.Fees = append(in.Fees, Fee{
			ParticipantID:   row.GameDayParticipantID,
			PenaltyTypeID:   row.PenaltyTypeID,
			PenaltyTypeName: row.PenaltyTypeName,
			Price:           int(row.PenaltyTypePrice),
			Count:           int(row.Count),
			QuantityScale:   int(row.QuantityScale),
		})
	}

	values, err := r.queries.StatsListCompetitionValues(ctx, db.StatsListCompetitionValuesParams{ClubID: p.ClubID, Date: from, Date_2: p.To})
	if err != nil {
		return Input{}, err
	}
	for _, row := range values {
		in.CompetitionValues = append(in.CompetitionValues, CompetitionValue{
			ParticipantID: row.GameDayParticipantID,
			CompetitionID: row.CompetitionID,
			Value:         int(row.Value),
		})
	}

	competitions, err := r.queries.GetCompetitionsByClubID(ctx, p.ClubID)
	if err != nil {
		return Input{}, err
	}
	for _, c := range competitions {
		in.Competitions = append(in.Competitions, CompetitionInfo{
			ID:               c.ID,
			Name:             c.Name,
			ScoringType:      c.ScoringType,
			IsGenderSpecific: c.IsGenderSpecific,
		})
	}

	playerID := p.PlayerID
	before, err := r.queries.StatsPlayerBalanceDeltaBefore(ctx, db.StatsPlayerBalanceDeltaBeforeParams{PlayerID: &playerID, TransactionDate: from})
	if err != nil {
		return Input{}, err
	}
	in.BalanceDeltaBefore = int(before)

	txs, err := r.queries.StatsListPlayerTransactions(ctx, db.StatsListPlayerTransactionsParams{
		PlayerID:          &playerID,
		TransactionDate:   from,
		TransactionDate_2: p.To,
	})
	if err != nil {
		return Input{}, err
	}
	for _, t := range txs {
		in.Transactions = append(in.Transactions, BalanceTransaction{Date: t.TransactionDate, Type: t.TransactionType, Amount: int(t.Amount)})
	}

	return in, nil
}
