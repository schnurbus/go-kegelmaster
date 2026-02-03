package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
	"github.com/schnurbus/go-kegelmaster/backend/internal/transaction"
)

// HandleListTransactions lists all transactions for a club with pagination
func (h *Handler) HandleListTransactions(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("clubId")
	if clubID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID ist erforderlich")
	}

	// Check permission
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeTransactions, role.PermissionTypeList)
	if err != nil {
		slog.Error("permission check failed", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung")
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	typeFilter := c.Query("type", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	listParams := transaction.ListTransactionsParams{
		ClubID: clubID,
		Page:   page,
		Limit:  limit,
	}
	if typeFilter != "" {
		t := transaction.TransactionType(typeFilter)
		if t.IsValid() {
			listParams.TransactionType = &t
		}
	}

	// Get transactions
	result, err := h.TransactionRepo.List(ctx, listParams)
	if err != nil {
		slog.Error("list transactions", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(PaginatedTransactionsResponseFromEntity(result))
}

// HandleGetTransaction gets a single transaction by ID
func (h *Handler) HandleGetTransaction(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("clubId")
	txID := c.Params("id")

	if clubID == "" || txID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Transaction-ID sind erforderlich")
	}

	// Check permission
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeTransactions, role.PermissionTypeView)
	if err != nil {
		slog.Error("permission check failed", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung")
	}

	// Get transaction
	tx, err := h.TransactionRepo.GetByID(ctx, txID)
	if err != nil {
		if errors.Is(err, transaction.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Transaktion nicht gefunden")
		}
		slog.Error("get transaction", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Verify it belongs to the club
	if tx.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Transaktion nicht gefunden")
	}

	return c.JSON(TransactionResponseFromEntity(*tx))
}

type createTransactionRequest struct {
	PlayerID        *string `json:"player_id,omitempty"`
	TransactionType string  `json:"transaction_type"`
	Amount          int     `json:"amount"`
	Description     string  `json:"description,omitempty"`
	TransactionDate *string `json:"transaction_date,omitempty"` // YYYY-MM-DD, optional
}

func (r createTransactionRequest) validate() error {
	// Validate transaction type
	txType := transaction.TransactionType(r.TransactionType)
	if !txType.IsValid() {
		return errors.New("Ungültiger Transaktionstyp")
	}

	// Manual transactions only
	if !txType.IsManual() {
		return errors.New("Nur manuelle Transaktionen können erstellt werden")
	}

	// Validate amount (always positive for user input)
	if r.Amount <= 0 {
		return errors.New("Betrag muss positiv sein")
	}

	// Description is mandatory
	if strings.TrimSpace(r.Description) == "" {
		return errors.New("Beschreibung ist erforderlich")
	}

	// Deposits and tips require a player
	if (txType == transaction.TransactionTypeDeposit || txType == transaction.TransactionTypeTip) && r.PlayerID == nil {
		return errors.New("Einzahlungen und Trinkgeld benötigen einen Spieler")
	}

	// If transaction_date is set, must be valid YYYY-MM-DD
	if r.TransactionDate != nil && strings.TrimSpace(*r.TransactionDate) != "" {
		if _, err := time.Parse("2006-01-02", *r.TransactionDate); err != nil {
			return errors.New("Transaktionsdatum muss im Format YYYY-MM-DD sein")
		}
	}

	return nil
}

// HandleCreateTransaction creates a new manual transaction (deposit, tip, expense)
func (h *Handler) HandleCreateTransaction(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("clubId")
	if clubID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID ist erforderlich")
	}

	// Check permission
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeTransactions, role.PermissionTypeCreate)
	if err != nil {
		slog.Error("permission check failed", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung")
	}

	// Parse request
	var req createTransactionRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if err := req.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Get club to check auto-tip setting
	clubEntity, err := h.ClubRepo.GetByID(ctx, clubID)
	if err != nil {
		slog.Error("get club", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Create transaction(s) - may create 2 if auto-tip splits deposit
	txType := transaction.TransactionType(req.TransactionType)
	txDate := time.Now().UTC().Truncate(24 * time.Hour)
	if req.TransactionDate != nil && strings.TrimSpace(*req.TransactionDate) != "" {
		if parsed, err := time.Parse("2006-01-02", *req.TransactionDate); err == nil {
			// Noon UTC so calendar day is preserved when stored as DATE
			txDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 12, 0, 0, 0, time.UTC)
		}
	}
	params := transaction.CreateTransactionParams{
		ClubID:           clubID,
		PlayerID:         req.PlayerID,
		TransactionType:  txType,
		Amount:           req.Amount,
		Description:      strings.TrimSpace(req.Description),
		TransactionDate:  txDate,
	}

	// Convert expense amounts to negative (user enters positive, backend stores negative)
	if txType == transaction.TransactionTypeExpense {
		params.Amount = -params.Amount
	}

	var transactions []transaction.Transaction
	if txType == transaction.TransactionTypeDeposit {
		// Use auto-tip logic for deposits
		transactions, err = h.TransactionRepo.CreateWithAutoTip(ctx, params, clubEntity.AutoTipEnabled)
	} else {
		// Regular creation for tips and expenses
		tx, err := h.TransactionRepo.Create(ctx, params)
		if err == nil {
			transactions = []transaction.Transaction{*tx}
		}
	}

	if err != nil {
		slog.Error("create transaction", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Erstellen der Transaktion")
	}

	// Return array of created transactions
	responses := TransactionsResponseFromEntities(transactions)

	// If only one transaction, return it directly
	if len(responses) == 1 {
		return c.Status(fiber.StatusCreated).JSON(responses[0])
	}

	// If multiple (deposit + auto-tip), return array
	return c.Status(fiber.StatusCreated).JSON(responses)
}

// HandleDeleteTransaction deletes a manual transaction
func (h *Handler) HandleDeleteTransaction(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("clubId")
	txID := c.Params("id")

	if clubID == "" || txID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Transaction-ID sind erforderlich")
	}

	// Check permission
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeTransactions, role.PermissionTypeDelete)
	if err != nil {
		slog.Error("permission check failed", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung")
	}

	// Get transaction to verify it belongs to the club
	tx, err := h.TransactionRepo.GetByID(ctx, txID)
	if err != nil {
		if errors.Is(err, transaction.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Transaktion nicht gefunden")
		}
		slog.Error("get transaction", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if tx.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Transaktion nicht gefunden")
	}

	// Delete transaction
	if err := h.TransactionRepo.Delete(ctx, txID); err != nil {
		if errors.Is(err, transaction.ErrNotManual) {
			return fiber.NewError(fiber.StatusBadRequest, "Nur manuelle Transaktionen können gelöscht werden")
		}
		slog.Error("delete transaction", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Löschen der Transaktion")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// HandleListPlayerTransactions lists all transactions for a specific player
func (h *Handler) HandleListPlayerTransactions(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("clubId")
	playerID := c.Params("playerId")

	if clubID == "" || playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Player-ID sind erforderlich")
	}

	// Check permission
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeTransactions, role.PermissionTypeList)
	if err != nil {
		slog.Error("permission check failed", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung")
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// Get transactions
	result, err := h.TransactionRepo.ListByPlayer(ctx, clubID, playerID, page, limit)
	if err != nil {
		slog.Error("list player transactions", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(PaginatedTransactionsResponseFromEntity(result))
}

// HandleListGameDayTransactions lists all transactions for a specific game day
func (h *Handler) HandleListGameDayTransactions(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("clubId")
	gameDayID := c.Params("gamedayId")

	if clubID == "" || gameDayID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und GameDay-ID sind erforderlich")
	}

	// Check permission
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeTransactions, role.PermissionTypeList)
	if err != nil {
		slog.Error("permission check failed", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung")
	}

	// Get transactions
	transactions, err := h.TransactionRepo.ListByGameDay(ctx, gameDayID)
	if err != nil {
		slog.Error("list gameday transactions", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(TransactionsResponseFromEntities(transactions))
}

// HandleGetGameDayTransactionSummary gets the transaction summary for a game day
func (h *Handler) HandleGetGameDayTransactionSummary(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("clubId")
	gameDayID := c.Params("gamedayId")

	if clubID == "" || gameDayID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und GameDay-ID sind erforderlich")
	}

	// Check permission
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeView)
	if err != nil {
		slog.Error("permission check failed", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung")
	}

	// Get summary
	summary, err := h.TransactionRepo.GetGameDaySummary(ctx, gameDayID)
	if err != nil {
		slog.Error("get gameday transaction summary", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(GameDayTransactionSummaryResponseFromEntity(summary))
}
