package transaction

import (
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("transaction not found")
	ErrInvalidAmount = errors.New("amount must be positive")
	ErrInvalidType   = errors.New("invalid transaction type")
	ErrNotManual     = errors.New("only manual transactions can be deleted")
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeBaseFee TransactionType = "base_fee"
	TransactionTypeFee     TransactionType = "fee"
	TransactionTypeDeposit TransactionType = "deposit"
	TransactionTypeTip     TransactionType = "tip"
	TransactionTypeExpense TransactionType = "expense"
)

// IsValid checks if the transaction type is valid
func (t TransactionType) IsValid() bool {
	switch t {
	case TransactionTypeBaseFee, TransactionTypeFee, TransactionTypeDeposit,
		TransactionTypeTip, TransactionTypeExpense:
		return true
	}
	return false
}

// IsManual returns true if the transaction was manually created
func (t TransactionType) IsManual() bool {
	return t == TransactionTypeDeposit || t == TransactionTypeTip || t == TransactionTypeExpense
}

// Transaction represents a financial transaction
type Transaction struct {
	ID                  string
	ClubID              string
	PlayerID            *string
	TransactionType     TransactionType
	Amount              int // Cents
	Description         string
	GameDayFeeID        *string
	GameDayID           *string
	PlayerBalanceBefore *int
	PlayerBalanceAfter  *int
	ClubBalanceBefore   int
	ClubBalanceAfter    int
	TransactionDate     time.Time // effective date (game day date for fee/base_fee)
	CreatedAt           time.Time
	UpdatedAt           time.Time

	// Joined fields (not stored)
	PlayerName string
}

// CreateTransactionParams for creating a new transaction
type CreateTransactionParams struct {
	ClubID           string
	PlayerID         *string
	TransactionType  TransactionType
	Amount           int
	Description      string
	GameDayFeeID     *string
	GameDayID        *string
	TransactionDate  time.Time // for fee/base_fee = game day date; for manual = effective date
}

// ListTransactionsParams for querying transactions
type ListTransactionsParams struct {
	ClubID          string
	PlayerID        *string
	GameDayID       *string
	TransactionType *TransactionType
	Page            int
	Limit           int
}

// PaginatedTransactions represents a paginated list of transactions
type PaginatedTransactions struct {
	Data       []Transaction
	Page       int
	Limit      int
	Total      int64
	TotalPages int
}

// BalanceValidation represents balance validation results
type BalanceValidation struct {
	ClubID            string
	PlayerID          *string
	StoredBalance     int
	CalculatedBalance int
	Discrepancy       int
	TransactionCount  int
}

// GameDayTransactionSummary represents transaction summary for a game day
type GameDayTransactionSummary struct {
	BaseFeeTotal int
	FeeTotal     int
	BaseFeeCount int64
	FeeCount     int64
}
