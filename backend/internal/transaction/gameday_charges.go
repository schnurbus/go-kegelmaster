package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
	"github.com/schnurbus/go-kegelmaster/backend/internal/gameday"
)

// ChargeError is a problem the user can act on. The database transaction is rolled back.
type ChargeError struct {
	Message string
}

func (e *ChargeError) Error() string {
	return e.Message
}

// SaveGameDayWithChargesParams creates or updates a game day and, when the day is
// published, books every base fee and missing penalty transaction in the same
// database transaction.
type SaveGameDayWithChargesParams struct {
	ID         string // empty creates a new game day
	ClubID     string
	Date       time.Time
	Notes      string
	IsDraft    bool
	ChargeFees bool
}

type postedCharge struct {
	playerID    string
	txType      TransactionType
	amount      int
	description string
	feeID       string
	hasFee      bool
}

// SaveGameDayWithCharges persists the game day together with all charges.
// A failure rolls the game day change back as well, so a published day cannot
// exist with only some of its fees.
func (r *Repository) SaveGameDayWithCharges(ctx context.Context, params SaveGameDayWithChargesParams) (gameday.GameDay, error) {
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return gameday.GameDay{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	// Lock players before the game day row. The other order deadlocks with a
	// fee update that already holds a player lock and then references the game day.
	players, err := qtx.GetPlayersByClubIDForUpdate(ctx, params.ClubID)
	if err != nil {
		return gameday.GameDay{}, fmt.Errorf("load players: %w", err)
	}
	roles, err := qtx.GetRolesByClubID(ctx, params.ClubID)
	if err != nil {
		return gameday.GameDay{}, fmt.Errorf("load roles: %w", err)
	}
	roleByID := make(map[string]db.Role, len(roles))
	for _, role := range roles {
		roleByID[role.ID] = role
	}
	clubRow, err := qtx.GetClubByID(ctx, params.ClubID)
	if err != nil {
		return gameday.GameDay{}, fmt.Errorf("load club: %w", err)
	}

	now := time.Now().UTC()
	notes := sql.NullString{String: params.Notes, Valid: params.Notes != ""}
	var saved db.GameDay
	if params.ID == "" {
		saved, err = qtx.CreateGameDay(ctx, db.CreateGameDayParams{
			ID:        uuid.NewString(),
			ClubID:    params.ClubID,
			Date:      params.Date,
			Notes:     notes,
			IsDraft:   params.IsDraft,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return gameday.GameDay{}, fmt.Errorf("create game day: %w", err)
		}
	} else {
		saved, err = qtx.UpdateGameDay(ctx, db.UpdateGameDayParams{
			ID:        params.ID,
			Date:      params.Date,
			Notes:     notes,
			IsDraft:   params.IsDraft,
			UpdatedAt: now,
		})
		if errors.Is(err, sql.ErrNoRows) {
			return gameday.GameDay{}, gameday.ErrNotFound
		}
		if err != nil {
			return gameday.GameDay{}, fmt.Errorf("update game day: %w", err)
		}
	}

	if params.ChargeFees && !params.IsDraft {
		if err := postGameDayCharges(ctx, qtx, params, saved.ID, clubRow, players, roleByID, now); err != nil {
			return gameday.GameDay{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return gameday.GameDay{}, fmt.Errorf("commit transaction: %w", err)
	}
	return gameDayFromDB(saved), nil
}

func postGameDayCharges(ctx context.Context, qtx *db.Queries, params SaveGameDayWithChargesParams, gameDayID string, clubRow db.Club, players []db.Player, roleByID map[string]db.Role, now time.Time) error {
	gameDayKey := gameDayID
	existing, err := qtx.ListTransactionsByGameDay(ctx, &gameDayKey)
	if err != nil {
		return fmt.Errorf("load transactions: %w", err)
	}
	fees, err := qtx.GetGameDayFeesByGameDay(ctx, gameDayID)
	if err != nil {
		return fmt.Errorf("load fees: %w", err)
	}

	charges, err := gameDayCharges(params, clubRow, players, roleByID, existing, fees)
	if err != nil {
		return err
	}
	if len(charges) == 0 {
		return nil
	}

	start := make(map[string]int, len(players))
	running := make(map[string]int, len(players))
	for _, player := range players {
		start[player.ID] = int(player.Balance)
		running[player.ID] = int(player.Balance)
	}

	clubBalance := int(clubRow.Balance)
	dateOnly := ledgerDate(params.Date)
	for _, charge := range charges {
		balance, ok := running[charge.playerID]
		if !ok {
			return &ChargeError{Message: "Ein Spieler der Gebühren gehört nicht zum Klub. Es wurde nichts geändert."}
		}
		before := balance
		after := balance + charge.amount
		running[charge.playerID] = after

		playerID := charge.playerID
		dayID := gameDayID
		var feeID *string
		if charge.hasFee {
			id := charge.feeID
			feeID = &id
		}
		if _, err := qtx.CreateTransaction(ctx, db.CreateTransactionParams{
			ID:                  uuid.NewString(),
			ClubID:              params.ClubID,
			PlayerID:            &playerID,
			TransactionType:     string(charge.txType),
			Amount:              int32(charge.amount),
			Description:         sql.NullString{String: charge.description, Valid: charge.description != ""},
			GameDayFeeID:        feeID,
			GameDayID:           &dayID,
			PlayerBalanceBefore: toNullInt32(&before),
			PlayerBalanceAfter:  toNullInt32(&after),
			ClubBalanceBefore:   int32(clubBalance),
			ClubBalanceAfter:    int32(clubBalance),
			TransactionDate:     dateOnly,
			CreatedAt:           now,
			UpdatedAt:           now,
		}); err != nil {
			return fmt.Errorf("create charge: %w", err)
		}
	}

	playerIDs := make([]string, 0, len(running))
	for id, balance := range running {
		if balance != start[id] {
			playerIDs = append(playerIDs, id)
		}
	}
	sort.Strings(playerIDs)
	for _, id := range playerIDs {
		if err := qtx.UpdatePlayerBalance(ctx, db.UpdatePlayerBalanceParams{
			ID:        id,
			Balance:   int32(running[id]),
			UpdatedAt: now,
		}); err != nil {
			return fmt.Errorf("update player balance: %w", err)
		}
	}
	return nil
}

func gameDayCharges(params SaveGameDayWithChargesParams, clubRow db.Club, players []db.Player, roleByID map[string]db.Role, existing []db.ListTransactionsByGameDayRow, fees []db.GetGameDayFeesByGameDayRow) ([]postedCharge, error) {
	baseFeePlayers := make(map[string]struct{})
	feeWithTx := make(map[string]struct{})
	for _, tx := range existing {
		if tx.TransactionType == string(TransactionTypeBaseFee) && tx.PlayerID != nil {
			baseFeePlayers[*tx.PlayerID] = struct{}{}
		}
		if tx.GameDayFeeID != nil {
			feeWithTx[*tx.GameDayFeeID] = struct{}{}
		}
	}

	charges := make([]postedCharge, 0)
	if clubRow.BaseFee > 0 {
		ordered := append([]db.Player(nil), players...)
		sort.Slice(ordered, func(i, j int) bool {
			if ordered[i].CreatedAt.Equal(ordered[j].CreatedAt) {
				return ordered[i].ID < ordered[j].ID
			}
			return ordered[i].CreatedAt.After(ordered[j].CreatedAt)
		})
		description := fmt.Sprintf("Grundgebühr für %s", params.Date.UTC().Format("02.01.2006"))
		for _, player := range ordered {
			if player.Inactive || player.RoleID == nil {
				continue
			}
			if _, ok := baseFeePlayers[player.ID]; ok {
				continue
			}
			role, ok := roleByID[*player.RoleID]
			if !ok {
				return nil, &ChargeError{Message: fmt.Sprintf("Für %s fehlt die Rolle. Die Gebühren wurden nicht gebucht und es wurde nichts geändert.", player.Name)}
			}
			if !role.PaysBaseFee {
				continue
			}
			charges = append(charges, postedCharge{
				playerID:    player.ID,
				txType:      TransactionTypeBaseFee,
				amount:      -int(clubRow.BaseFee),
				description: description,
			})
		}
	}

	for _, fee := range fees {
		if _, ok := feeWithTx[fee.ID]; ok {
			continue
		}
		amount := FeeAmount(int(fee.PenaltyTypePrice), int(fee.Count), int(fee.QuantityScale))
		if amount == 0 {
			continue
		}
		charges = append(charges, postedCharge{
			playerID:    fee.PlayerID,
			txType:      TransactionTypeFee,
			amount:      amount,
			description: FeeTransactionDescription(fee.PenaltyTypeName, int(fee.Count), int(fee.QuantityScale)),
			feeID:       fee.ID,
			hasFee:      true,
		})
	}
	return charges, nil
}

func ledgerDate(txDate time.Time) time.Time {
	if txDate.IsZero() {
		txDate = time.Now()
	}
	return txDate.UTC().Truncate(24 * time.Hour).Add(12 * time.Hour)
}

func gameDayFromDB(row db.GameDay) gameday.GameDay {
	notes := ""
	if row.Notes.Valid {
		notes = row.Notes.String
	}
	return gameday.GameDay{
		ID:        row.ID,
		ClubID:    row.ClubID,
		Date:      row.Date,
		Notes:     notes,
		IsDraft:   row.IsDraft,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
