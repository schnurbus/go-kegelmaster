package gameday

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
)

type Repository struct {
	queries *db.Queries
}

func NewRepository(dbConn *sql.DB) *Repository {
	return &Repository{
		queries: db.New(dbConn),
	}
}

// ==================== GAME DAY OPERATIONS ====================

type CreateGameDayParams struct {
	ClubID string
	Date   time.Time
	Notes  string
}

func (r *Repository) Create(ctx context.Context, params CreateGameDayParams) (GameDay, error) {
	now := time.Now().UTC()
	id := uuid.NewString()

	notes := sql.NullString{String: params.Notes, Valid: params.Notes != ""}

	dbGameDay, err := r.queries.CreateGameDay(ctx, db.CreateGameDayParams{
		ID:        id,
		ClubID:    params.ClubID,
		Date:      params.Date,
		Notes:     notes,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return GameDay{}, err
	}

	return dbGameDayToGameDay(dbGameDay), nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (GameDay, error) {
	dbGameDay, err := r.queries.GetGameDayByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return GameDay{}, ErrNotFound
	}
	if err != nil {
		return GameDay{}, err
	}

	return dbGameDayToGameDay(dbGameDay), nil
}

func (r *Repository) GetByClubID(ctx context.Context, clubID string) ([]GameDay, error) {
	dbGameDays, err := r.queries.GetGameDaysByClubID(ctx, clubID)
	if err != nil {
		return nil, err
	}

	gameDays := make([]GameDay, len(dbGameDays))
	for i, dbGameDay := range dbGameDays {
		gameDays[i] = dbGameDayToGameDay(dbGameDay)
	}

	return gameDays, nil
}

type UpdateGameDayParams struct {
	ID    string
	Date  time.Time
	Notes string
}

func (r *Repository) Update(ctx context.Context, params UpdateGameDayParams) (GameDay, error) {
	now := time.Now().UTC()

	notes := sql.NullString{String: params.Notes, Valid: params.Notes != ""}

	dbGameDay, err := r.queries.UpdateGameDay(ctx, db.UpdateGameDayParams{
		ID:        params.ID,
		Date:      params.Date,
		Notes:     notes,
		UpdatedAt: now,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return GameDay{}, ErrNotFound
	}
	if err != nil {
		return GameDay{}, err
	}

	return dbGameDayToGameDay(dbGameDay), nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	// Check if exists
	_, err := r.queries.GetGameDayByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	// Delete (CASCADE will handle participants and fees)
	err = r.queries.DeleteGameDay(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) CheckExistsByClubAndDate(ctx context.Context, clubID string, date time.Time) (bool, error) {
	exists, err := r.queries.CheckGameDayExistsByClubAndDate(ctx, db.CheckGameDayExistsByClubAndDateParams{
		ClubID: clubID,
		Date:   date,
	})
	if err != nil {
		return false, err
	}

	return exists, nil
}

// ==================== PARTICIPANT OPERATIONS ====================

func (r *Repository) AddParticipant(ctx context.Context, gameDayID, playerID string) (GameDayParticipant, error) {
	now := time.Now().UTC()
	id := uuid.NewString()

	dbParticipant, err := r.queries.CreateGameDayParticipant(ctx, db.CreateGameDayParticipantParams{
		ID:        id,
		GameDayID: gameDayID,
		PlayerID:  playerID,
		CreatedAt: now,
	})
	if err != nil {
		return GameDayParticipant{}, err
	}

	return dbParticipantToParticipant(dbParticipant), nil
}

func (r *Repository) RemoveParticipant(ctx context.Context, gameDayID, playerID string) error {
	// Check if exists
	_, err := r.queries.GetParticipantByGameDayAndPlayer(ctx, db.GetParticipantByGameDayAndPlayerParams{
		GameDayID: gameDayID,
		PlayerID:  playerID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return ErrParticipantNotFound
	}
	if err != nil {
		return err
	}

	// Delete (CASCADE will handle fees)
	err = r.queries.DeleteGameDayParticipant(ctx, db.DeleteGameDayParticipantParams{
		GameDayID: gameDayID,
		PlayerID:  playerID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetParticipants(ctx context.Context, gameDayID string) ([]GameDayParticipant, error) {
	dbParticipants, err := r.queries.GetGameDayParticipants(ctx, gameDayID)
	if err != nil {
		return nil, err
	}

	participants := make([]GameDayParticipant, len(dbParticipants))
	for i, dbPart := range dbParticipants {
		participants[i] = dbParticipantWithNameToParticipant(dbPart)
	}

	return participants, nil
}

func (r *Repository) GetParticipantByGameDayAndPlayer(ctx context.Context, gameDayID, playerID string) (GameDayParticipant, error) {
	dbParticipant, err := r.queries.GetParticipantByGameDayAndPlayer(ctx, db.GetParticipantByGameDayAndPlayerParams{
		GameDayID: gameDayID,
		PlayerID:  playerID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return GameDayParticipant{}, ErrParticipantNotFound
	}
	if err != nil {
		return GameDayParticipant{}, err
	}

	return dbParticipantToParticipant(dbParticipant), nil
}

// ==================== FEE OPERATIONS ====================

type UpsertFeeParams struct {
	ParticipantID          string
	PenaltyTypeID          string
	PenaltyTypeName        string
	PenaltyTypeDescription string
	PenaltyTypePrice       int
	Count                  int   // scaled when QuantityScale > 1
	QuantityScale          int   // 1 or 100
}

func (r *Repository) UpsertFee(ctx context.Context, params UpsertFeeParams) (GameDayFee, error) {
	now := time.Now().UTC()
	id := uuid.NewString()

	description := sql.NullString{String: params.PenaltyTypeDescription, Valid: params.PenaltyTypeDescription != ""}

	dbFee, err := r.queries.UpsertGameDayFee(ctx, db.UpsertGameDayFeeParams{
		ID:                     id,
		GameDayParticipantID:   params.ParticipantID,
		PenaltyTypeID:          params.PenaltyTypeID,
		PenaltyTypeName:        params.PenaltyTypeName,
		PenaltyTypeDescription: description,
		PenaltyTypePrice:       int32(params.PenaltyTypePrice),
		Count:                  int32(params.Count),
		QuantityScale:          int32(params.QuantityScale),
		CreatedAt:              now,
		UpdatedAt:              now,
	})
	if err != nil {
		return GameDayFee{}, err
	}

	return dbFeeToFee(dbFee), nil
}

func (r *Repository) GetFeesByParticipant(ctx context.Context, participantID string) ([]GameDayFee, error) {
	dbFees, err := r.queries.GetGameDayFeesByParticipant(ctx, participantID)
	if err != nil {
		return nil, err
	}

	fees := make([]GameDayFee, len(dbFees))
	for i, dbFee := range dbFees {
		fees[i] = dbFeeToFee(dbFee)
	}

	return fees, nil
}

func (r *Repository) DeleteFee(ctx context.Context, participantID, penaltyTypeID string) error {
	err := r.queries.DeleteGameDayFeeByParticipantAndType(ctx, db.DeleteGameDayFeeByParticipantAndTypeParams{
		GameDayParticipantID: participantID,
		PenaltyTypeID:        penaltyTypeID,
	})
	if err != nil {
		return err
	}

	return nil
}

// ==================== COMPETITION VALUE OPERATIONS ====================

type UpsertCompetitionValueParams struct {
	ParticipantID string
	CompetitionID string
	Value         int
}

func (r *Repository) UpsertCompetitionValue(ctx context.Context, params UpsertCompetitionValueParams) (GameDayCompetitionValue, error) {
	now := time.Now().UTC()
	id := uuid.NewString()

	dbVal, err := r.queries.UpsertGameDayCompetitionValue(ctx, db.UpsertGameDayCompetitionValueParams{
		ID:                   id,
		GameDayParticipantID: params.ParticipantID,
		CompetitionID:        params.CompetitionID,
		Value:                int32(params.Value),
		CreatedAt:            now,
		UpdatedAt:            now,
	})
	if err != nil {
		return GameDayCompetitionValue{}, err
	}

	return dbCompetitionValueToValue(dbVal), nil
}

func (r *Repository) GetCompetitionValuesByParticipant(ctx context.Context, participantID string) ([]GameDayCompetitionValue, error) {
	dbVals, err := r.queries.GetGameDayCompetitionValuesByParticipant(ctx, participantID)
	if err != nil {
		return nil, err
	}

	vals := make([]GameDayCompetitionValue, len(dbVals))
	for i, dbVal := range dbVals {
		vals[i] = dbCompetitionValueToValue(dbVal)
	}

	return vals, nil
}

func dbCompetitionValueToValue(dbV db.GameDayCompetitionValue) GameDayCompetitionValue {
	return GameDayCompetitionValue{
		ID:                   dbV.ID,
		GameDayParticipantID: dbV.GameDayParticipantID,
		CompetitionID:        dbV.CompetitionID,
		Value:                int(dbV.Value),
		CreatedAt:            dbV.CreatedAt,
		UpdatedAt:            dbV.UpdatedAt,
	}
}

// ==================== SUMMARIES ====================

func (r *Repository) GetSummariesByClubID(ctx context.Context, clubID string) ([]GameDaySummary, error) {
	dbSummaries, err := r.queries.GetGameDaySummariesByClubID(ctx, clubID)
	if err != nil {
		return nil, err
	}

	summaries := make([]GameDaySummary, len(dbSummaries))
	for i, dbSum := range dbSummaries {
		summaries[i] = dbSummaryToSummary(dbSum)
	}

	return summaries, nil
}

// ==================== COMPLEX QUERIES ====================

func (r *Repository) GetGameDayWithDetails(ctx context.Context, gameDayID string) (GameDayDetail, error) {
	// Get game day
	gameDay, err := r.GetByID(ctx, gameDayID)
	if err != nil {
		return GameDayDetail{}, err
	}

	// Get participants
	participants, err := r.GetParticipants(ctx, gameDayID)
	if err != nil {
		return GameDayDetail{}, err
	}

	// Get fees and competition values for each participant
	participantsWithFees := make([]ParticipantWithFees, len(participants))
	for i, participant := range participants {
		fees, err := r.GetFeesByParticipant(ctx, participant.ID)
		if err != nil {
			return GameDayDetail{}, err
		}

		competitionValues, err := r.GetCompetitionValuesByParticipant(ctx, participant.ID)
		if err != nil {
			return GameDayDetail{}, err
		}

		participantsWithFees[i] = ParticipantWithFees{
			Participant:      participant,
			Fees:             fees,
			CompetitionValues: competitionValues,
		}
	}

	return GameDayDetail{
		GameDay:      gameDay,
		Participants: participantsWithFees,
	}, nil
}

// GetPlayerPenaltyHistory returns penalty counts per game day for a player (for dashboard chart).
// Only game days where the player participated are included. since limits to game days on or after that date.
func (r *Repository) GetPlayerPenaltyHistory(ctx context.Context, clubID, playerID string, since time.Time) ([]PenaltyHistoryDay, error) {
	rows, err := r.queries.GetPlayerPenaltyHistoryByClubAndPlayer(ctx, db.GetPlayerPenaltyHistoryByClubAndPlayerParams{
		ClubID:   clubID,
		PlayerID: playerID,
		Date:     since,
	})
	if err != nil {
		return nil, err
	}
	// Group by game_day_id
	byDay := make(map[string]*PenaltyHistoryDay)
	for _, row := range rows {
		day, ok := byDay[row.GameDayID]
		if !ok {
			day = &PenaltyHistoryDay{GameDayID: row.GameDayID, Date: row.Date, Penalties: nil}
			byDay[row.GameDayID] = day
		}
		if row.PenaltyTypeID != nil && row.PenaltyTypeName.Valid {
			scale := 1
			if row.QuantityScale.Valid && row.QuantityScale.Int32 > 0 {
				scale = int(row.QuantityScale.Int32)
			}
			qty := float64(0)
			if row.Count.Valid {
				qty = float64(row.Count.Int32) / float64(scale)
			}
			day.Penalties = append(day.Penalties, PenaltyHistoryPenalty{
				PenaltyTypeID:   *row.PenaltyTypeID,
				PenaltyTypeName: row.PenaltyTypeName.String,
				Quantity:        qty,
			})
		}
	}
	// Preserve order by date (rows are ordered by gd.date ASC)
	ordered := make([]PenaltyHistoryDay, 0, len(byDay))
	seen := make(map[string]struct{})
	for _, row := range rows {
		if _, ok := seen[row.GameDayID]; ok {
			continue
		}
		seen[row.GameDayID] = struct{}{}
		ordered = append(ordered, *byDay[row.GameDayID])
	}
	return ordered, nil
}

// GetPlayerCompetitionHistory returns competition values per game day for a player (for dashboard chart).
// Only game days where the player participated and had competition values are included.
func (r *Repository) GetPlayerCompetitionHistory(ctx context.Context, clubID, playerID string, since time.Time) ([]CompetitionHistoryDay, error) {
	rows, err := r.queries.GetPlayerCompetitionHistoryByClubAndPlayer(ctx, db.GetPlayerCompetitionHistoryByClubAndPlayerParams{
		ClubID:   clubID,
		PlayerID: playerID,
		Date:     since,
	})
	if err != nil {
		return nil, err
	}
	byDay := make(map[string]*CompetitionHistoryDay)
	for _, row := range rows {
		day, ok := byDay[row.GameDayID]
		if !ok {
			day = &CompetitionHistoryDay{GameDayID: row.GameDayID, Date: row.Date, Values: nil}
			byDay[row.GameDayID] = day
		}
		day.Values = append(day.Values, CompetitionHistoryValue{
			CompetitionID:   row.CompetitionID,
			CompetitionName: row.CompetitionName,
			Value:           int(row.Value),
		})
	}
	ordered := make([]CompetitionHistoryDay, 0, len(byDay))
	seen := make(map[string]struct{})
	for _, row := range rows {
		if _, ok := seen[row.GameDayID]; ok {
			continue
		}
		seen[row.GameDayID] = struct{}{}
		ordered = append(ordered, *byDay[row.GameDayID])
	}
	return ordered, nil
}

// ==================== CONVERTERS ====================

func dbGameDayToGameDay(dbGD db.GameDay) GameDay {
	notes := ""
	if dbGD.Notes.Valid {
		notes = dbGD.Notes.String
	}

	return GameDay{
		ID:        dbGD.ID,
		ClubID:    dbGD.ClubID,
		Date:      dbGD.Date,
		Notes:     notes,
		CreatedAt: dbGD.CreatedAt,
		UpdatedAt: dbGD.UpdatedAt,
	}
}

func dbParticipantToParticipant(dbP db.GameDayParticipant) GameDayParticipant {
	return GameDayParticipant{
		ID:        dbP.ID,
		GameDayID: dbP.GameDayID,
		PlayerID:  dbP.PlayerID,
		CreatedAt: dbP.CreatedAt,
	}
}

func dbParticipantWithNameToParticipant(dbP db.GetGameDayParticipantsRow) GameDayParticipant {
	var playerGender *string
	if dbP.PlayerGender.Valid {
		playerGender = &dbP.PlayerGender.String
	}
	return GameDayParticipant{
		ID:           dbP.ID,
		GameDayID:    dbP.GameDayID,
		PlayerID:     dbP.PlayerID,
		PlayerName:   dbP.PlayerName,
		PlayerGender: playerGender,
		CreatedAt:    dbP.CreatedAt,
	}
}

func dbFeeToFee(dbF db.GameDayFee) GameDayFee {
	description := ""
	if dbF.PenaltyTypeDescription.Valid {
		description = dbF.PenaltyTypeDescription.String
	}

	return GameDayFee{
		ID:                     dbF.ID,
		GameDayParticipantID:   dbF.GameDayParticipantID,
		PenaltyTypeID:          dbF.PenaltyTypeID,
		PenaltyTypeName:        dbF.PenaltyTypeName,
		PenaltyTypeDescription: description,
		PenaltyTypePrice:       int(dbF.PenaltyTypePrice),
		Count:                  int(dbF.Count),
		QuantityScale:          int(dbF.QuantityScale),
		CreatedAt:              dbF.CreatedAt,
		UpdatedAt:              dbF.UpdatedAt,
	}
}

func dbSummaryToSummary(dbS db.GetGameDaySummariesByClubIDRow) GameDaySummary {
	notes := ""
	if dbS.Notes.Valid {
		notes = dbS.Notes.String
	}

	return GameDaySummary{
		ID:               dbS.ID,
		ClubID:           dbS.ClubID,
		Date:             dbS.Date,
		Notes:            notes,
		ParticipantCount: int(dbS.ParticipantCount),
		PenaltyFeeTotal:  int(dbS.PenaltyFeeTotal),
		CreatedAt:        dbS.CreatedAt,
		UpdatedAt:        dbS.UpdatedAt,
	}
}
