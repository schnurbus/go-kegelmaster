package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
)

// ParticipantFeeChange is one penalty line for a game-day participant.
// Count is the stored quantity (already scaled). Count <= 0 deletes the fee.
type ParticipantFeeChange struct {
	PenaltyTypeID          string
	PenaltyTypeName        string
	PenaltyTypeDescription string
	PenaltyTypePrice       int
	Count                  int
	QuantityScale          int
}

// SyncParticipantFeesParams replaces the fee lines of one participant.
// When RecordTransactions is set, matching money transactions and the player
// balance are updated in the same database transaction.
type SyncParticipantFeesParams struct {
	ClubID             string
	PlayerID           string
	GameDayID          string
	ParticipantID      string
	TransactionDate    time.Time
	RecordTransactions bool
	Fees               []ParticipantFeeChange
}

// FeeAmount is the signed transaction amount in cents for a fee line.
// Fees are stored as a negative amount (debt).
func FeeAmount(price, count, quantityScale int) int {
	if count <= 0 {
		return 0
	}
	scale := quantityScale
	if scale < 1 {
		scale = 1
	}
	return -(price * count / scale)
}

// FeeTransactionDescription is the ledger text for a fee transaction.
func FeeTransactionDescription(name string, count, quantityScale int) string {
	scale := quantityScale
	if scale < 1 {
		scale = 1
	}
	quantity := fmt.Sprintf("%d", count)
	if scale > 1 {
		quantity = fmt.Sprintf("%.2f", float64(count)/float64(scale))
	}
	return fmt.Sprintf("%s ×%s", name, quantity)
}

type feeSyncOp struct {
	delete      bool
	change      ParticipantFeeChange
	oldTxID     string
	oldTxAmount int
	hasOldTx    bool
}

// SyncParticipantFees writes every fee change for one participant in a single
// database transaction. Unchanged lines are left untouched, so saving a game
// day again does not rewrite transactions or player balances.
func (r *Repository) SyncParticipantFees(ctx context.Context, params SyncParticipantFeesParams) error {
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	// Lock the player before the fee rows. Fee transactions also reference the
	// player, and the opposite order deadlocks with a concurrent insert.
	var running int
	if params.RecordTransactions {
		playerRow, err := qtx.GetPlayerByIDForUpdate(ctx, params.PlayerID)
		if err != nil {
			return fmt.Errorf("load player: %w", err)
		}
		running = int(playerRow.Balance)
	}

	existingFees, err := qtx.GetGameDayFeesByParticipantForUpdate(ctx, params.ParticipantID)
	if err != nil {
		return fmt.Errorf("load fees: %w", err)
	}
	existingByType := make(map[string]db.GameDayFee, len(existingFees))
	for _, fee := range existingFees {
		existingByType[fee.PenaltyTypeID] = fee
	}

	ops := make([]feeSyncOp, 0, len(params.Fees))
	for _, change := range params.Fees {
		if change.PenaltyTypeID == "" {
			continue
		}
		existing, ok := existingByType[change.PenaltyTypeID]
		if change.Count <= 0 {
			if !ok {
				continue
			}
			op := feeSyncOp{delete: true, change: change}
			if params.RecordTransactions {
				oldTx, found, err := feeTransactionByFeeID(ctx, qtx, existing.ID)
				if err != nil {
					return err
				}
				if found {
					op.hasOldTx = true
					op.oldTxID = oldTx.ID
					op.oldTxAmount = int(oldTx.Amount)
				}
			}
			ops = append(ops, op)
			continue
		}

		scale := change.QuantityScale
		if scale < 1 {
			scale = 1
		}
		amount := FeeAmount(change.PenaltyTypePrice, change.Count, scale)
		description := FeeTransactionDescription(change.PenaltyTypeName, change.Count, scale)

		var oldTx db.Transaction
		var foundTx bool
		if ok && params.RecordTransactions {
			oldTx, foundTx, err = feeTransactionByFeeID(ctx, qtx, existing.ID)
			if err != nil {
				return err
			}
		}
		if ok && int(existing.Count) == change.Count && int(existing.QuantityScale) == scale {
			if !params.RecordTransactions {
				continue
			}
			if foundTx && int(oldTx.Amount) == amount && oldTx.Description.String == description {
				continue
			}
		}

		op := feeSyncOp{change: change}
		if foundTx {
			op.hasOldTx = true
			op.oldTxID = oldTx.ID
			op.oldTxAmount = int(oldTx.Amount)
		}
		ops = append(ops, op)
	}

	if len(ops) == 0 {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit transaction: %w", err)
		}
		return nil
	}

	startBalance := running
	var (
		clubBalance int
		clubLoaded  bool
	)
	now := time.Now().UTC()
	dateOnly := params.TransactionDate
	if dateOnly.IsZero() {
		dateOnly = now
	}
	dateOnly = dateOnly.UTC().Truncate(24 * time.Hour).Add(12 * time.Hour)
	playerID := params.PlayerID
	gameDayID := params.GameDayID

	for _, op := range ops {
		if op.delete {
			if op.hasOldTx {
				if err := qtx.DeleteTransaction(ctx, op.oldTxID); err != nil {
					return fmt.Errorf("delete fee transaction: %w", err)
				}
				running -= op.oldTxAmount
			}
			if err := qtx.DeleteGameDayFeeByParticipantAndType(ctx, db.DeleteGameDayFeeByParticipantAndTypeParams{
				GameDayParticipantID: params.ParticipantID,
				PenaltyTypeID:        op.change.PenaltyTypeID,
			}); err != nil {
				return fmt.Errorf("delete fee: %w", err)
			}
			continue
		}

		change := op.change
		scale := change.QuantityScale
		if scale < 1 {
			scale = 1
		}
		created, err := qtx.UpsertGameDayFee(ctx, db.UpsertGameDayFeeParams{
			ID:                   uuid.NewString(),
			GameDayParticipantID: params.ParticipantID,
			PenaltyTypeID:        change.PenaltyTypeID,
			PenaltyTypeName:      change.PenaltyTypeName,
			PenaltyTypeDescription: sql.NullString{
				String: change.PenaltyTypeDescription,
				Valid:  change.PenaltyTypeDescription != "",
			},
			PenaltyTypePrice: int32(change.PenaltyTypePrice),
			Count:            int32(change.Count),
			QuantityScale:    int32(scale),
			CreatedAt:        now,
			UpdatedAt:        now,
		})
		if err != nil {
			return fmt.Errorf("upsert fee: %w", err)
		}

		if !params.RecordTransactions {
			continue
		}
		if op.hasOldTx {
			if err := qtx.DeleteTransaction(ctx, op.oldTxID); err != nil {
				return fmt.Errorf("delete fee transaction: %w", err)
			}
			running -= op.oldTxAmount
		}

		amount := FeeAmount(change.PenaltyTypePrice, change.Count, scale)
		if amount == 0 {
			continue
		}
		if !clubLoaded {
			clubRow, err := qtx.GetClubByID(ctx, params.ClubID)
			if err != nil {
				return fmt.Errorf("load club: %w", err)
			}
			clubBalance = int(clubRow.Balance)
			clubLoaded = true
		}

		before := running
		running += amount
		feeID := created.ID
		description := FeeTransactionDescription(change.PenaltyTypeName, change.Count, scale)
		if _, err := qtx.CreateTransaction(ctx, db.CreateTransactionParams{
			ID:                  uuid.NewString(),
			ClubID:              params.ClubID,
			PlayerID:            &playerID,
			TransactionType:     string(TransactionTypeFee),
			Amount:              int32(amount),
			Description:         sql.NullString{String: description, Valid: description != ""},
			GameDayFeeID:        &feeID,
			GameDayID:           &gameDayID,
			PlayerBalanceBefore: toNullInt32(&before),
			PlayerBalanceAfter:  toNullInt32(&running),
			ClubBalanceBefore:   int32(clubBalance),
			ClubBalanceAfter:    int32(clubBalance),
			TransactionDate:     dateOnly,
			CreatedAt:           now,
			UpdatedAt:           now,
		}); err != nil {
			return fmt.Errorf("create fee transaction: %w", err)
		}
	}

	if params.RecordTransactions && running != startBalance {
		if err := qtx.UpdatePlayerBalance(ctx, db.UpdatePlayerBalanceParams{
			ID:        params.PlayerID,
			Balance:   int32(running),
			UpdatedAt: now,
		}); err != nil {
			return fmt.Errorf("update player balance: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func feeTransactionByFeeID(ctx context.Context, qtx *db.Queries, feeID string) (db.Transaction, bool, error) {
	id := feeID
	row, err := qtx.GetTransactionByGameDayFee(ctx, &id)
	if errors.Is(err, sql.ErrNoRows) {
		return db.Transaction{}, false, nil
	}
	if err != nil {
		return db.Transaction{}, false, fmt.Errorf("load fee transaction: %w", err)
	}
	return row, true, nil
}
