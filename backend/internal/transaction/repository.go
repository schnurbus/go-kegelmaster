package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
	"github.com/schnurbus/go-kegelmaster/backend/internal/gameday"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
)

// Repository provides access to transaction storage
type Repository struct {
	queries     *db.Queries
	database    *sql.DB
	playerRepo  *player.Repository
	clubRepo    *club.Repository
	gameDayRepo *gameday.Repository
}

// NewRepository creates a new transaction repository
func NewRepository(database *sql.DB, playerRepo *player.Repository, clubRepo *club.Repository, gameDayRepo *gameday.Repository) *Repository {
	return &Repository{
		queries:     db.New(database),
		database:    database,
		playerRepo:  playerRepo,
		clubRepo:    clubRepo,
		gameDayRepo: gameDayRepo,
	}
}

// Create creates a new transaction and updates balances in one database
// transaction. The player row is locked and the club balance is changed with
// an atomic increment, so concurrent bookings cannot overwrite each other.
func (r *Repository) Create(ctx context.Context, params CreateTransactionParams) (*Transaction, error) {
	if err := validateCreateParams(params); err != nil {
		return nil, err
	}

	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	created, err := createInTx(ctx, r.queries.WithTx(tx), params, time.Now())
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return created, nil
}

func validateCreateParams(params CreateTransactionParams) error {
	if !params.TransactionType.IsValid() {
		return ErrInvalidType
	}
	if params.Amount == 0 || params.Amount > MaxAmount || params.Amount < -MaxAmount {
		return ErrInvalidAmount
	}
	return nil
}

// lockPlayer loads the player with a row lock and checks the club.
func lockPlayer(ctx context.Context, qtx *db.Queries, playerID, clubID string) (db.Player, error) {
	p, err := qtx.GetPlayerByIDForUpdate(ctx, playerID)
	if errors.Is(err, sql.ErrNoRows) {
		return db.Player{}, player.ErrNotFound
	}
	if err != nil {
		return db.Player{}, fmt.Errorf("failed to get player: %w", err)
	}
	if p.ClubID != clubID {
		return db.Player{}, ErrPlayerNotInClub
	}
	return p, nil
}

// changeClubBalance adds delta to the club balance and returns the balance
// before and after. A zero delta only reads the balance.
func changeClubBalance(ctx context.Context, qtx *db.Queries, clubID string, delta int, now time.Time) (before, after int, err error) {
	if delta == 0 {
		c, err := qtx.GetClubByID(ctx, clubID)
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, club.ErrNotFound
		}
		if err != nil {
			return 0, 0, fmt.Errorf("failed to get club: %w", err)
		}
		return int(c.Balance), int(c.Balance), nil
	}
	newBalance, err := qtx.IncrementClubBalance(ctx, db.IncrementClubBalanceParams{
		Balance:   int32(delta),
		UpdatedAt: now,
		ID:        clubID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, club.ErrNotFound
	}
	if err != nil {
		return 0, 0, fmt.Errorf("failed to update club balance: %w", err)
	}
	return int(newBalance) - delta, int(newBalance), nil
}

// createInTx books one transaction inside an open database transaction.
func createInTx(ctx context.Context, qtx *db.Queries, params CreateTransactionParams, now time.Time) (*Transaction, error) {
	var (
		playerBalanceBefore, playerBalanceAfter *int
		playerRow                               db.Player
		hasPlayer                               bool
		playerDelta, clubDelta                  int
	)
	if params.PlayerID != nil {
		p, err := lockPlayer(ctx, qtx, *params.PlayerID, params.ClubID)
		if err != nil {
			return nil, err
		}
		playerRow = p
		hasPlayer = true
		before := int(p.Balance)
		playerBalanceBefore = &before
	}

	switch params.TransactionType {
	case TransactionTypeBaseFee, TransactionTypeFee:
		// Fees are negative (debt); club cash does not change.
		playerDelta = params.Amount
	case TransactionTypeDeposit:
		// Deposits reduce the player's debt and add cash to the club.
		playerDelta = params.Amount
		clubDelta = params.Amount
	case TransactionTypeTip:
		// Tips add cash to the club; the player's balance is unchanged.
		clubDelta = params.Amount
	case TransactionTypeExpense:
		// Expenses are negative and only reduce club cash.
		clubDelta = params.Amount
	}
	if hasPlayer && params.TransactionType != TransactionTypeExpense {
		after := int(playerRow.Balance) + playerDelta
		playerBalanceAfter = &after
	}

	clubBalanceBefore, clubBalanceAfter, err := changeClubBalance(ctx, qtx, params.ClubID, clubDelta, now)
	if err != nil {
		return nil, err
	}

	txDate := params.TransactionDate
	if txDate.IsZero() && params.GameDayID != nil {
		if gd, err := qtx.GetGameDayByID(ctx, *params.GameDayID); err == nil {
			txDate = gd.Date
		}
	}
	if txDate.IsZero() {
		txDate = now
	}
	// Normalize to noon UTC so the calendar day is preserved when PostgreSQL
	// converts to session timezone for DATE storage (avoids off-by-one in western TZ).
	dateOnly := txDate.UTC().Truncate(24 * time.Hour).Add(12 * time.Hour)
	dbTx, err := qtx.CreateTransaction(ctx, db.CreateTransactionParams{
		ID:                  uuid.New().String(),
		ClubID:              params.ClubID,
		PlayerID:            params.PlayerID,
		TransactionType:     string(params.TransactionType),
		Amount:              int32(params.Amount),
		Description:         sql.NullString{String: params.Description, Valid: params.Description != ""},
		GameDayFeeID:        params.GameDayFeeID,
		GameDayID:           params.GameDayID,
		PlayerBalanceBefore: toNullInt32(playerBalanceBefore),
		PlayerBalanceAfter:  toNullInt32(playerBalanceAfter),
		ClubBalanceBefore:   int32(clubBalanceBefore),
		ClubBalanceAfter:    int32(clubBalanceAfter),
		TransactionDate:     dateOnly,
		CreatedAt:           now,
		UpdatedAt:           now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	if hasPlayer && playerDelta != 0 {
		if err := qtx.UpdatePlayerBalance(ctx, db.UpdatePlayerBalanceParams{
			ID:        playerRow.ID,
			Balance:   int32(*playerBalanceAfter),
			UpdatedAt: now,
		}); err != nil {
			return nil, fmt.Errorf("failed to update player balance: %w", err)
		}
	}

	return dbTransactionToTransaction(dbTx), nil
}

// CreateWithAutoTip creates a deposit and, with auto-tip enabled, books the
// part exceeding the player's debt as a tip. Deposit and tip share one
// database transaction, so a failure never leaves only half of the split.
func (r *Repository) CreateWithAutoTip(ctx context.Context, params CreateTransactionParams, autoTipEnabled bool) ([]Transaction, error) {
	if params.TransactionType != TransactionTypeDeposit || !autoTipEnabled {
		created, err := r.Create(ctx, params)
		if err != nil {
			return nil, err
		}
		return []Transaction{*created}, nil
	}
	if params.PlayerID == nil {
		return nil, fmt.Errorf("player_id required for deposits")
	}
	if err := validateCreateParams(params); err != nil {
		return nil, err
	}

	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	qtx := r.queries.WithTx(tx)
	now := time.Now()

	// Decide the split on the locked balance; createInTx locks the same row again.
	playerRow, err := lockPlayer(ctx, qtx, *params.PlayerID, params.ClubID)
	if err != nil {
		return nil, err
	}

	debt := 0
	if playerRow.Balance < 0 {
		debt = -int(playerRow.Balance)
	}
	depositAmount := params.Amount
	if depositAmount > debt {
		depositAmount = debt
	}
	tipAmount := params.Amount - depositAmount

	var results []Transaction
	if depositAmount > 0 {
		deposit := params
		deposit.Amount = depositAmount
		created, err := createInTx(ctx, qtx, deposit, now)
		if err != nil {
			return nil, err
		}
		results = append(results, *created)
	}
	if tipAmount > 0 {
		description := "Auto-tip from deposit excess"
		if depositAmount == 0 {
			description = "Auto-tip (player has no debt)"
		}
		tip := params
		tip.TransactionType = TransactionTypeTip
		tip.Amount = tipAmount
		tip.Description = description
		created, err := createInTx(ctx, qtx, tip, now)
		if err != nil {
			return nil, err
		}
		results = append(results, *created)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return results, nil
}

// CreateDepositCouplesMode creates deposit transaction(s) for multiple players (Paar-Modus).
// Total amount is used as a pool: each player gets at most their debt as deposit; the remainder is auto-tip (or surplus deposit if auto-tip disabled).
// All players are locked (in ID order, like the game-day booking) before
// their balances are read, so concurrent bookings cannot be lost.
func (r *Repository) CreateDepositCouplesMode(ctx context.Context, clubID string, playerIDs []string, totalAmount int, description string, txDate time.Time, autoTipEnabled bool) ([]Transaction, error) {
	n := len(playerIDs)
	if n < 2 {
		return nil, fmt.Errorf("couples mode requires at least 2 players, got %d", n)
	}
	if totalAmount <= 0 || totalAmount > MaxAmount {
		return nil, ErrInvalidAmount
	}
	seen := make(map[string]struct{}, n)
	for _, id := range playerIDs {
		if _, ok := seen[id]; ok {
			return nil, ErrDuplicatePlayer
		}
		seen[id] = struct{}{}
	}

	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	qtx := r.queries.WithTx(tx)

	lockOrder := append([]string(nil), playerIDs...)
	sort.Strings(lockOrder)
	byID := make(map[string]db.Player, n)
	for _, id := range lockOrder {
		p, err := lockPlayer(ctx, qtx, id, clubID)
		if err != nil {
			if errors.Is(err, ErrPlayerNotInClub) {
				return nil, fmt.Errorf("player %s does not belong to club: %w", id, err)
			}
			return nil, fmt.Errorf("failed to get player %s: %w", id, err)
		}
		byID[id] = p
	}
	players := make([]db.Player, n)
	for i, id := range playerIDs {
		players[i] = byID[id]
	}

	now := time.Now()
	clubBalanceBefore, clubBalanceAfter, err := changeClubBalance(ctx, qtx, clubID, totalAmount, now)
	if err != nil {
		return nil, err
	}
	dateOnly := txDate.UTC().Truncate(24 * time.Hour).Add(12 * time.Hour)

	playerBalances := make([]int, n)
	for i := range players {
		playerBalances[i] = int(players[i].Balance)
	}

	var results []Transaction
	book := func(i int, txType TransactionType, amount int, desc string) error {
		before := playerBalances[i]
		after := before
		if txType == TransactionTypeDeposit {
			after = before + amount
		}
		pid := players[i].ID
		dbTx, err := qtx.CreateTransaction(ctx, db.CreateTransactionParams{
			ID:                  uuid.New().String(),
			ClubID:              clubID,
			PlayerID:            &pid,
			TransactionType:     string(txType),
			Amount:              int32(amount),
			Description:         sql.NullString{String: desc, Valid: desc != ""},
			PlayerBalanceBefore: toNullInt32(&before),
			PlayerBalanceAfter:  toNullInt32(&after),
			ClubBalanceBefore:   int32(clubBalanceBefore),
			ClubBalanceAfter:    int32(clubBalanceAfter),
			TransactionDate:     dateOnly,
			CreatedAt:           now,
			UpdatedAt:           now,
		})
		if err != nil {
			return fmt.Errorf("failed to create %s transaction: %w", txType, err)
		}
		results = append(results, *dbTransactionToTransaction(dbTx))
		if after != before {
			playerBalances[i] = after
			if err := qtx.UpdatePlayerBalance(ctx, db.UpdatePlayerBalanceParams{
				ID:        pid,
				Balance:   int32(after),
				UpdatedAt: now,
			}); err != nil {
				return fmt.Errorf("failed to update player balance: %w", err)
			}
		}
		return nil
	}

	remaining := totalAmount
	for i := 0; i < n && remaining > 0; i++ {
		debt := 0
		if playerBalances[i] < 0 {
			debt = -playerBalances[i]
		}
		amount := debt
		if amount > remaining {
			amount = remaining
		}
		if amount <= 0 {
			continue
		}
		remaining -= amount
		if err := book(i, TransactionTypeDeposit, amount, description); err != nil {
			return nil, err
		}
	}

	if remaining > 0 {
		if autoTipEnabled {
			if err := book(0, TransactionTypeTip, remaining, "Auto-tip (Paar-Modus Überschuss)"); err != nil {
				return nil, err
			}
		} else {
			restBase := remaining / n
			restRem := remaining % n
			for i := 0; i < n; i++ {
				extra := restBase
				if i < restRem {
					extra++
				}
				if extra <= 0 {
					continue
				}
				if err := book(i, TransactionTypeDeposit, extra, "Guthaben (Paar-Modus Überschuss)"); err != nil {
					return nil, err
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}
	return results, nil
}

// GetByID retrieves a transaction by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*Transaction, error) {
	dbTx, err := r.queries.GetTransactionByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return dbTransactionWithPlayerToTransaction(dbTx), nil
}

// List retrieves a paginated list of transactions
func (r *Repository) List(ctx context.Context, params ListTransactionsParams) (*PaginatedTransactions, error) {
	if params.Limit == 0 {
		params.Limit = 50
	}
	if params.Page < 1 {
		params.Page = 1
	}

	offset := (params.Page - 1) * params.Limit

	if params.TransactionType != nil {
		// Filter by transaction type (server-side for correct pagination)
		total, err := r.queries.CountTransactionsByType(ctx, db.CountTransactionsByTypeParams{
			ClubID:          params.ClubID,
			TransactionType: string(*params.TransactionType),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to count transactions by type: %w", err)
		}
		dbTxs, err := r.queries.ListTransactionsByType(ctx, db.ListTransactionsByTypeParams{
			ClubID:          params.ClubID,
			TransactionType: string(*params.TransactionType),
			Limit:           int32(params.Limit),
			Offset:          int32(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list transactions by type: %w", err)
		}
		transactions := make([]Transaction, len(dbTxs))
		for i, dbTx := range dbTxs {
			transactions[i] = *dbListTypeRowToTransaction(dbTx)
		}
		totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))
		return &PaginatedTransactions{
			Data:       transactions,
			Page:       params.Page,
			Limit:      params.Limit,
			Total:      total,
			TotalPages: totalPages,
		}, nil
	}

	// Get total count
	total, err := r.queries.CountTransactionsByClub(ctx, params.ClubID)
	if err != nil {
		return nil, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Get transactions
	dbTxs, err := r.queries.ListTransactionsByClub(ctx, db.ListTransactionsByClubParams{
		ClubID: params.ClubID,
		Limit:  int32(params.Limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	// Convert to domain models
	transactions := make([]Transaction, len(dbTxs))
	for i, dbTx := range dbTxs {
		transactions[i] = *dbListClubRowToTransaction(dbTx)
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))

	return &PaginatedTransactions{
		Data:       transactions,
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// ListByPlayer retrieves transactions for a specific player
func (r *Repository) ListByPlayer(ctx context.Context, clubID, playerID string, page, limit int) (*PaginatedTransactions, error) {
	if limit == 0 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	// Get total count
	total, err := r.queries.CountTransactionsByPlayer(ctx, db.CountTransactionsByPlayerParams{
		ClubID:   clubID,
		PlayerID: &playerID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Get transactions
	dbTxs, err := r.queries.ListTransactionsByPlayer(ctx, db.ListTransactionsByPlayerParams{
		ClubID:   clubID,
		PlayerID: &playerID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	// Convert to domain models
	transactions := make([]Transaction, len(dbTxs))
	for i, dbTx := range dbTxs {
		transactions[i] = *dbListPlayerRowToTransaction(dbTx)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &PaginatedTransactions{
		Data:       transactions,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// ListByGameDay retrieves all transactions for a game day
func (r *Repository) ListByGameDay(ctx context.Context, gameDayID string) ([]Transaction, error) {
	dbTxs, err := r.queries.ListTransactionsByGameDay(ctx, &gameDayID)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	transactions := make([]Transaction, len(dbTxs))
	for i, dbTx := range dbTxs {
		transactions[i] = *dbListGameDayRowToTransaction(dbTx)
	}

	return transactions, nil
}

// GetByGameDayFee retrieves the transaction for a game day fee
func (r *Repository) GetByGameDayFee(ctx context.Context, gameDayFeeID string) (*Transaction, error) {
	dbTx, err := r.queries.GetTransactionByGameDayFee(ctx, &gameDayFeeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return dbTransactionToTransaction(dbTx), nil
}

// DeleteFeeTransaction deletes a fee or base_fee transaction and reverts player balance.
// Used when updating fee counts so that re-import remains idempotent.
func (r *Repository) DeleteFeeTransaction(ctx context.Context, id string) error {
	tx, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if tx.TransactionType != TransactionTypeFee && tx.TransactionType != TransactionTypeBaseFee {
		return ErrNotManual
	}

	dbTx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer dbTx.Rollback()

	qtx := r.queries.WithTx(dbTx)

	if tx.PlayerID != nil {
		playerEntity, err := r.playerRepo.GetByID(ctx, *tx.PlayerID)
		if err != nil {
			return fmt.Errorf("failed to get player: %w", err)
		}
		// Revert: fee amount is negative (debt), so add it back to reduce debt
		newBalance := playerEntity.Balance - tx.Amount
		err = qtx.UpdatePlayerBalance(ctx, db.UpdatePlayerBalanceParams{
			ID:        playerEntity.ID,
			Balance:   int32(newBalance),
			UpdatedAt: time.Now(),
		})
		if err != nil {
			return fmt.Errorf("failed to revert player balance: %w", err)
		}
	}

	if err := qtx.DeleteTransaction(ctx, id); err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}
	if err := dbTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}

// Delete deletes a manual transaction and recalculates affected player and club balances from remaining transactions.
func (r *Repository) Delete(ctx context.Context, id string) error {
	tx, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !tx.TransactionType.IsManual() {
		return ErrNotManual
	}

	dbTx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer dbTx.Rollback()

	qtx := r.queries.WithTx(dbTx)
	if err := qtx.DeleteTransaction(ctx, id); err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}
	if err := dbTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Recalculate balances so they match the sum of remaining transactions
	if tx.PlayerID != nil {
		if err := r.RecalculatePlayerBalance(ctx, *tx.PlayerID); err != nil {
			return fmt.Errorf("failed to recalculate player balance: %w", err)
		}
	}
	if err := r.RecalculateClubBalance(ctx, tx.ClubID); err != nil {
		return fmt.Errorf("failed to recalculate club balance: %w", err)
	}
	return nil
}

// GetGameDaySummary retrieves transaction summary for a game day
func (r *Repository) GetGameDaySummary(ctx context.Context, gameDayID string) (*GameDayTransactionSummary, error) {
	summary, err := r.queries.GetGameDayTransactionSummary(ctx, &gameDayID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game day summary: %w", err)
	}

	baseFeeTotal := int64(0)
	feeTotal := int64(0)
	if summary.BaseFeeTotal != nil {
		baseFeeTotal = summary.BaseFeeTotal.(int64)
	}
	if summary.FeeTotal != nil {
		feeTotal = summary.FeeTotal.(int64)
	}

	return &GameDayTransactionSummary{
		BaseFeeTotal: int(baseFeeTotal),
		FeeTotal:     int(feeTotal),
		BaseFeeCount: summary.BaseFeeCount,
		FeeCount:     summary.FeeCount,
	}, nil
}

// RecalculatePlayerBalance recomputes the player balance from start_balance + SUM(amount) of transactions that affect player balance (base_fee, fee, deposit). Tip does not change player balance.
func (r *Repository) RecalculatePlayerBalance(ctx context.Context, playerID string) error {
	playerEntity, err := r.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}
	sum, err := r.queries.SumPlayerBalanceDelta(ctx, &playerID)
	if err != nil {
		return fmt.Errorf("failed to sum transactions: %w", err)
	}
	newBalance := playerEntity.StartBalance + int(sum)
	now := time.Now()
	return r.queries.UpdatePlayerBalance(ctx, db.UpdatePlayerBalanceParams{
		ID:        playerID,
		Balance:   int32(newBalance),
		UpdatedAt: now,
	})
}

// RecalculateClubBalance recomputes the club balance from start_balance + SUM(amount) of transactions that affect the till (deposit, tip, expense). Base_fee and fee do not change club cash.
func (r *Repository) RecalculateClubBalance(ctx context.Context, clubID string) error {
	clubEntity, err := r.clubRepo.GetByID(ctx, clubID)
	if err != nil {
		return fmt.Errorf("failed to get club: %w", err)
	}
	sum, err := r.queries.SumAmountByClubCash(ctx, clubID)
	if err != nil {
		return fmt.Errorf("failed to sum transactions: %w", err)
	}
	newBalance := clubEntity.StartBalance + int(sum)
	return r.clubRepo.UpdateBalance(ctx, clubID, newBalance)
}

// Helper functions

func dbTransactionToTransaction(dbTx db.Transaction) *Transaction {
	return &Transaction{
		ID:                  dbTx.ID,
		ClubID:              dbTx.ClubID,
		PlayerID:            dbTx.PlayerID,
		TransactionType:     TransactionType(dbTx.TransactionType),
		Amount:              int(dbTx.Amount),
		Description:         dbTx.Description.String,
		GameDayFeeID:        dbTx.GameDayFeeID,
		GameDayID:           dbTx.GameDayID,
		PlayerBalanceBefore: fromNullInt32(dbTx.PlayerBalanceBefore),
		PlayerBalanceAfter:  fromNullInt32(dbTx.PlayerBalanceAfter),
		ClubBalanceBefore:   int(dbTx.ClubBalanceBefore),
		ClubBalanceAfter:    int(dbTx.ClubBalanceAfter),
		TransactionDate:     dbTx.TransactionDate,
		CreatedAt:           dbTx.CreatedAt,
		UpdatedAt:           dbTx.UpdatedAt,
	}
}

func dbTransactionWithPlayerToTransaction(dbTx db.GetTransactionByIDRow) *Transaction {
	return &Transaction{
		ID:                  dbTx.ID,
		ClubID:              dbTx.ClubID,
		PlayerID:            dbTx.PlayerID,
		TransactionType:     TransactionType(dbTx.TransactionType),
		Amount:              int(dbTx.Amount),
		Description:         dbTx.Description.String,
		GameDayFeeID:        dbTx.GameDayFeeID,
		GameDayID:           dbTx.GameDayID,
		PlayerBalanceBefore: fromNullInt32(dbTx.PlayerBalanceBefore),
		PlayerBalanceAfter:  fromNullInt32(dbTx.PlayerBalanceAfter),
		ClubBalanceBefore:   int(dbTx.ClubBalanceBefore),
		ClubBalanceAfter:    int(dbTx.ClubBalanceAfter),
		TransactionDate:     dbTx.TransactionDate,
		CreatedAt:           dbTx.CreatedAt,
		UpdatedAt:           dbTx.UpdatedAt,
		PlayerName:          fromNullString(dbTx.PlayerName),
	}
}

func toNullInt32(i *int) sql.NullInt32 {
	if i == nil {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{Int32: int32(*i), Valid: true}
}

func fromNullInt32(n sql.NullInt32) *int {
	if !n.Valid {
		return nil
	}
	i := int(n.Int32)
	return &i
}

func fromNullString(n sql.NullString) string {
	if !n.Valid {
		return ""
	}
	return n.String
}

func dbListClubRowToTransaction(dbTx db.ListTransactionsByClubRow) *Transaction {
	return &Transaction{
		ID:                  dbTx.ID,
		ClubID:              dbTx.ClubID,
		PlayerID:            dbTx.PlayerID,
		TransactionType:     TransactionType(dbTx.TransactionType),
		Amount:              int(dbTx.Amount),
		Description:         dbTx.Description.String,
		GameDayFeeID:        dbTx.GameDayFeeID,
		GameDayID:           dbTx.GameDayID,
		PlayerBalanceBefore: fromNullInt32(dbTx.PlayerBalanceBefore),
		PlayerBalanceAfter:  fromNullInt32(dbTx.PlayerBalanceAfter),
		ClubBalanceBefore:   int(dbTx.ClubBalanceBefore),
		ClubBalanceAfter:    int(dbTx.ClubBalanceAfter),
		TransactionDate:     dbTx.TransactionDate,
		CreatedAt:           dbTx.CreatedAt,
		UpdatedAt:           dbTx.UpdatedAt,
		PlayerName:          fromNullString(dbTx.PlayerName),
	}
}

func dbListPlayerRowToTransaction(dbTx db.ListTransactionsByPlayerRow) *Transaction {
	return &Transaction{
		ID:                  dbTx.ID,
		ClubID:              dbTx.ClubID,
		PlayerID:            dbTx.PlayerID,
		TransactionType:     TransactionType(dbTx.TransactionType),
		Amount:              int(dbTx.Amount),
		Description:         dbTx.Description.String,
		GameDayFeeID:        dbTx.GameDayFeeID,
		GameDayID:           dbTx.GameDayID,
		PlayerBalanceBefore: fromNullInt32(dbTx.PlayerBalanceBefore),
		PlayerBalanceAfter:  fromNullInt32(dbTx.PlayerBalanceAfter),
		ClubBalanceBefore:   int(dbTx.ClubBalanceBefore),
		ClubBalanceAfter:    int(dbTx.ClubBalanceAfter),
		TransactionDate:     dbTx.TransactionDate,
		CreatedAt:           dbTx.CreatedAt,
		UpdatedAt:           dbTx.UpdatedAt,
		PlayerName:          fromNullString(dbTx.PlayerName),
	}
}

func dbListTypeRowToTransaction(dbTx db.ListTransactionsByTypeRow) *Transaction {
	return &Transaction{
		ID:                  dbTx.ID,
		ClubID:              dbTx.ClubID,
		PlayerID:            dbTx.PlayerID,
		TransactionType:     TransactionType(dbTx.TransactionType),
		Amount:              int(dbTx.Amount),
		Description:         dbTx.Description.String,
		GameDayFeeID:        dbTx.GameDayFeeID,
		GameDayID:           dbTx.GameDayID,
		PlayerBalanceBefore: fromNullInt32(dbTx.PlayerBalanceBefore),
		PlayerBalanceAfter:  fromNullInt32(dbTx.PlayerBalanceAfter),
		ClubBalanceBefore:   int(dbTx.ClubBalanceBefore),
		ClubBalanceAfter:    int(dbTx.ClubBalanceAfter),
		TransactionDate:     dbTx.TransactionDate,
		CreatedAt:           dbTx.CreatedAt,
		UpdatedAt:           dbTx.UpdatedAt,
		PlayerName:          fromNullString(dbTx.PlayerName),
	}
}

func dbListGameDayRowToTransaction(dbTx db.ListTransactionsByGameDayRow) *Transaction {
	return &Transaction{
		ID:                  dbTx.ID,
		ClubID:              dbTx.ClubID,
		PlayerID:            dbTx.PlayerID,
		TransactionType:     TransactionType(dbTx.TransactionType),
		Amount:              int(dbTx.Amount),
		Description:         dbTx.Description.String,
		GameDayFeeID:        dbTx.GameDayFeeID,
		GameDayID:           dbTx.GameDayID,
		PlayerBalanceBefore: fromNullInt32(dbTx.PlayerBalanceBefore),
		PlayerBalanceAfter:  fromNullInt32(dbTx.PlayerBalanceAfter),
		ClubBalanceBefore:   int(dbTx.ClubBalanceBefore),
		ClubBalanceAfter:    int(dbTx.ClubBalanceAfter),
		TransactionDate:     dbTx.TransactionDate,
		CreatedAt:           dbTx.CreatedAt,
		UpdatedAt:           dbTx.UpdatedAt,
		PlayerName:          fromNullString(dbTx.PlayerName),
	}
}
