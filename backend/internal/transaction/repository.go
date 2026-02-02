package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
)

// Repository provides access to transaction storage
type Repository struct {
	queries    *db.Queries
	database   *sql.DB
	playerRepo *player.Repository
	clubRepo   *club.Repository
}

// NewRepository creates a new transaction repository
func NewRepository(database *sql.DB, playerRepo *player.Repository, clubRepo *club.Repository) *Repository {
	return &Repository{
		queries:    db.New(database),
		database:   database,
		playerRepo: playerRepo,
		clubRepo:   clubRepo,
	}
}

// Create creates a new transaction and updates balances
func (r *Repository) Create(ctx context.Context, params CreateTransactionParams) (*Transaction, error) {
	// Validate transaction type
	if !params.TransactionType.IsValid() {
		return nil, ErrInvalidType
	}

	// Validate amount
	if params.Amount == 0 {
		return nil, ErrInvalidAmount
	}

	// Get current balances
	var playerBalanceBefore, playerBalanceAfter *int
	var clubBalanceBefore, clubBalanceAfter int

	// Get club
	clubEntity, err := r.clubRepo.GetByID(ctx, params.ClubID)
	if err != nil {
		return nil, fmt.Errorf("failed to get club: %w", err)
	}
	clubBalanceBefore = clubEntity.Balance

	// Get player if applicable
	var playerEntity player.Player
	var hasPlayer bool
	if params.PlayerID != nil {
		playerEntity, err = r.playerRepo.GetByID(ctx, *params.PlayerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get player: %w", err)
		}
		hasPlayer = true
		before := playerEntity.Balance
		playerBalanceBefore = &before
	}

	// Calculate new balances based on transaction type
	switch params.TransactionType {
	case TransactionTypeBaseFee, TransactionTypeFee:
		// Fees are negative (debt)
		// Player balance decreases (more debt)
		if hasPlayer {
			after := playerEntity.Balance + params.Amount // Amount is negative
			playerBalanceAfter = &after
			playerEntity.Balance = after
		}
		// Club balance unchanged for fees

	case TransactionTypeDeposit:
		// Deposit is positive
		// Player balance increases (less debt)
		// Club balance increases (money received)
		if hasPlayer {
			after := playerEntity.Balance - params.Amount // Subtract positive amount to reduce debt
			playerBalanceAfter = &after
			playerEntity.Balance = after
		}
		clubBalanceAfter = clubBalanceBefore + params.Amount

	case TransactionTypeTip:
		// Tip is positive
		// Player balance unchanged
		// Club balance increases
		if hasPlayer {
			after := playerEntity.Balance
			playerBalanceAfter = &after
		}
		clubBalanceAfter = clubBalanceBefore + params.Amount

	case TransactionTypeExpense:
		// Expense is negative
		// Player balance unchanged
		// Club balance decreases
		clubBalanceAfter = clubBalanceBefore + params.Amount // Amount is negative
	}

	// If club balance didn't change, set it to before value
	if clubBalanceAfter == 0 {
		clubBalanceAfter = clubBalanceBefore
	}

	// Begin transaction
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	// Create transaction record
	now := time.Now()
	txID := uuid.New().String()

	dbTx, err := qtx.CreateTransaction(ctx, db.CreateTransactionParams{
		ID:                  txID,
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
		CreatedAt:           now,
		UpdatedAt:           now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Update player balance if applicable
	if hasPlayer {
		_, err = qtx.UpdatePlayer(ctx, db.UpdatePlayerParams{
			ID:           playerEntity.ID,
			Name:         playerEntity.Name,
			Balance:      int32(playerEntity.Balance),
			StartBalance: int32(playerEntity.StartBalance),
			UpdatedAt:    now,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update player balance: %w", err)
		}
	}

	// Update club balance if changed
	if clubBalanceAfter != clubBalanceBefore {
		_, err = qtx.UpdateClub(ctx, db.UpdateClubParams{
			ID:             clubEntity.ID,
			Name:           clubEntity.Name,
			Balance:        int32(clubBalanceAfter),
			BaseFee:        int32(clubEntity.BaseFee),
			AutoTipEnabled: clubEntity.AutoTipEnabled,
			UpdatedAt:      now,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update club balance: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Convert to domain model
	return dbTransactionToTransaction(dbTx), nil
}

// CreateWithAutoTip creates a deposit transaction with auto-tip logic
func (r *Repository) CreateWithAutoTip(ctx context.Context, params CreateTransactionParams, autoTipEnabled bool) ([]Transaction, error) {
	// Only applies to deposits
	if params.TransactionType != TransactionTypeDeposit {
		tx, err := r.Create(ctx, params)
		if err != nil {
			return nil, err
		}
		return []Transaction{*tx}, nil
	}

	// Validate we have a player
	if params.PlayerID == nil {
		return nil, fmt.Errorf("player_id required for deposits")
	}

	// Get player
	playerEntity, err := r.playerRepo.GetByID(ctx, *params.PlayerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	// Check if auto-tip should trigger
	if !autoTipEnabled {
		// No auto-tip, create normal deposit
		tx, err := r.Create(ctx, params)
		if err != nil {
			return nil, err
		}
		return []Transaction{*tx}, nil
	}

	// Auto-tip is enabled
	transactions := []Transaction{}

	if playerEntity.Balance >= 0 {
		// Player has no debt, entire amount becomes tip
		tipParams := CreateTransactionParams{
			ClubID:          params.ClubID,
			PlayerID:        params.PlayerID,
			TransactionType: TransactionTypeTip,
			Amount:          params.Amount,
			Description:     "Auto-tip (player has no debt)",
			GameDayID:       params.GameDayID,
		}
		tipTx, err := r.Create(ctx, tipParams)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, *tipTx)
		return transactions, nil
	}

	// Player has debt (balance is negative)
	debt := -playerEntity.Balance // Make it positive
	if params.Amount > debt {
		// Split into deposit + tip
		depositAmount := debt
		tipAmount := params.Amount - debt

		// Create deposit transaction
		depositParams := CreateTransactionParams{
			ClubID:          params.ClubID,
			PlayerID:        params.PlayerID,
			TransactionType: TransactionTypeDeposit,
			Amount:          depositAmount,
			Description:     params.Description,
			GameDayID:       params.GameDayID,
		}
		depositTx, err := r.Create(ctx, depositParams)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, *depositTx)

		// Create tip transaction
		tipParams := CreateTransactionParams{
			ClubID:          params.ClubID,
			PlayerID:        params.PlayerID,
			TransactionType: TransactionTypeTip,
			Amount:          tipAmount,
			Description:     "Auto-tip from deposit excess",
			GameDayID:       params.GameDayID,
		}
		tipTx, err := r.Create(ctx, tipParams)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, *tipTx)

		return transactions, nil
	}

	// Deposit doesn't exceed debt, create normal deposit
	tx, err := r.Create(ctx, params)
	if err != nil {
		return nil, err
	}
	return []Transaction{*tx}, nil
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
		var gender sql.NullString
		if playerEntity.Gender != nil && *playerEntity.Gender != "" {
			gender = sql.NullString{String: *playerEntity.Gender, Valid: true}
		}
		_, err = qtx.UpdatePlayer(ctx, db.UpdatePlayerParams{
			ID:           playerEntity.ID,
			Name:         playerEntity.Name,
			Balance:      int32(newBalance),
			StartBalance: int32(playerEntity.StartBalance),
			UserID:       playerEntity.UserID,
			RoleID:       playerEntity.RoleID,
			Gender:       gender,
			UpdatedAt:    time.Now(),
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

// Delete deletes a transaction and reverts balances
func (r *Repository) Delete(ctx context.Context, id string) error {
	// Get transaction first
	tx, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if it's a manual transaction
	if !tx.TransactionType.IsManual() {
		return ErrNotManual
	}

	// Begin database transaction
	dbTx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer dbTx.Rollback()

	qtx := r.queries.WithTx(dbTx)

	// Revert balances
	// For deposits: player balance increases (more debt), club balance decreases
	// For tips: club balance decreases
	// For expenses: club balance increases

	if tx.PlayerID != nil {
		playerEntity, err := r.playerRepo.GetByID(ctx, *tx.PlayerID)
		if err != nil {
			return fmt.Errorf("failed to get player: %w", err)
		}

		var newBalance int
		if tx.TransactionType == TransactionTypeDeposit {
			// Revert deposit: add back the amount (increase debt)
			newBalance = playerEntity.Balance + tx.Amount
		}
		// Tips and expenses don't affect player balance

		if tx.TransactionType == TransactionTypeDeposit {
			_, err = qtx.UpdatePlayer(ctx, db.UpdatePlayerParams{
				ID:           playerEntity.ID,
				Name:         playerEntity.Name,
				Balance:      int32(newBalance),
				StartBalance: int32(playerEntity.StartBalance),
				UpdatedAt:    time.Now(),
			})
			if err != nil {
				return fmt.Errorf("failed to update player balance: %w", err)
			}
		}
	}

	// Update club balance
	clubEntity, err := r.clubRepo.GetByID(ctx, tx.ClubID)
	if err != nil {
		return fmt.Errorf("failed to get club: %w", err)
	}

	var newClubBalance int
	switch tx.TransactionType {
	case TransactionTypeDeposit, TransactionTypeTip:
		// Revert: subtract the amount
		newClubBalance = clubEntity.Balance - tx.Amount
	case TransactionTypeExpense:
		// Revert: subtract the negative amount (add back)
		newClubBalance = clubEntity.Balance - tx.Amount
	}

	_, err = qtx.UpdateClub(ctx, db.UpdateClubParams{
		ID:             clubEntity.ID,
		Name:           clubEntity.Name,
		Balance:        int32(newClubBalance),
		BaseFee:        int32(clubEntity.BaseFee),
		AutoTipEnabled: clubEntity.AutoTipEnabled,
		UpdatedAt:      time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to update club balance: %w", err)
	}

	// Delete transaction
	err = qtx.DeleteTransaction(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	// Commit
	if err := dbTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
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
		CreatedAt:           dbTx.CreatedAt,
		UpdatedAt:           dbTx.UpdatedAt,
		PlayerName:          fromNullString(dbTx.PlayerName),
	}
}
