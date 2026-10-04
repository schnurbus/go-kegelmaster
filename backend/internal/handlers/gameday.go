package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/audit"
	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/competition"
	"github.com/schnurbus/go-kegelmaster/backend/internal/gameday"
	"github.com/schnurbus/go-kegelmaster/backend/internal/penaltytype"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
	"github.com/schnurbus/go-kegelmaster/backend/internal/transaction"
)

// ==================== REQUEST/RESPONSE TYPES ====================

type createGameDayRequest struct {
	Date    string `json:"date"` // ISO date format YYYY-MM-DD
	Notes   string `json:"notes"`
	IsDraft *bool  `json:"is_draft"` // optional; default true (vorläufig)
}

func (r createGameDayRequest) validate() error {
	if strings.TrimSpace(r.Date) == "" {
		return errors.New("Datum ist erforderlich")
	}

	_, err := time.Parse("2006-01-02", r.Date)
	if err != nil {
		return errors.New("Ungültiges Datumsformat (erwartet YYYY-MM-DD)")
	}

	return nil
}

type updateGameDayRequest struct {
	Date    string `json:"date"`
	Notes   string `json:"notes"`
	IsDraft *bool  `json:"is_draft"` // optional; when omitted keep existing value
}

func (r updateGameDayRequest) validate() error {
	return createGameDayRequest{Date: r.Date, Notes: r.Notes}.validate()
}

type addParticipantRequest struct {
	PlayerID string `json:"player_id"`
}

type updateFeesRequest struct {
	Fees []feeInput `json:"fees"`
}

type feeInput struct {
	PenaltyTypeID string  `json:"penalty_type_id"`
	Count         float64 `json:"count"` // decimal allowed when penalty type allows_decimal_quantity
}

// ==================== HANDLERS ====================

func (h *Handler) HandleCreateGameDay(c fiber.Ctx) error {
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
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeCreate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Spieltage zu erstellen")
	}

	var req createGameDayRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if err := req.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	date, _ := time.Parse("2006-01-02", req.Date)

	isDraft := true
	if req.IsDraft != nil {
		isDraft = *req.IsDraft
	}

	var gameDayEntity gameday.GameDay
	if isDraft {
		gameDayEntity, err = h.GameDayRepo.Create(ctx, gameday.CreateGameDayParams{
			ClubID:  clubID,
			Date:    date,
			Notes:   req.Notes,
			IsDraft: true,
		})
		if err != nil {
			slog.Error("create game day", "error", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
		}
	} else {
		// Publishing books every base fee in the same transaction as the game day.
		// A failure rolls the insert back, so the day cannot exist with missing fees.
		gameDayEntity, err = h.TransactionRepo.SaveGameDayWithCharges(ctx, transaction.SaveGameDayWithChargesParams{
			ClubID:     clubID,
			Date:       date,
			Notes:      req.Notes,
			IsDraft:    false,
			ChargeFees: true,
		})
		if err != nil {
			return gameDayChargeError(err)
		}
	}

	audit.LogAudit(c, u.ID, audit.ActionGameDayCreated, "club_id", clubID, "gameday_id", gameDayEntity.ID)
	return c.Status(fiber.StatusCreated).JSON(GameDayResponseFromEntity(gameDayEntity))
}

func (h *Handler) HandleGetGameDays(c fiber.Ctx) error {
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
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeList)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Spieltage einzusehen")
	}

	gameDays, err := h.GameDayRepo.GetByClubID(ctx, clubID)
	if err != nil {
		slog.Error("get game days", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(GameDaysResponseFromEntities(gameDays))
}

func (h *Handler) HandleGetGameDaySummaries(c fiber.Ctx) error {
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
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeList)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Spieltage einzusehen")
	}

	summaries, err := h.GameDayRepo.GetSummariesByClubID(ctx, clubID)
	if err != nil {
		slog.Error("get game day summaries", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(GameDaySummariesResponseFromEntities(summaries))
}

func (h *Handler) HandleGetGameDay(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("clubId")
	gameDayID := c.Params("id")
	if clubID == "" || gameDayID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Spieltag-ID sind erforderlich")
	}

	// Check permission
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeView)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Spieltag einzusehen")
	}

	// Get with full details
	detail, err := h.GameDayRepo.GetGameDayWithDetails(ctx, gameDayID)
	if err != nil {
		if errors.Is(err, gameday.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Spieltag nicht gefunden")
		}
		slog.Error("get game day", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Verify belongs to club
	if detail.GameDay.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Spieltag nicht gefunden")
	}

	return c.JSON(GameDayDetailResponseFromEntity(detail))
}

func (h *Handler) HandleUpdateGameDay(c fiber.Ctx) error {
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
	gameDayID := c.Params("id")
	if clubID == "" || gameDayID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Spieltag-ID sind erforderlich")
	}

	// Check permission
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeUpdate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Spieltag zu bearbeiten")
	}

	var req updateGameDayRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if err := req.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Verify game day exists and belongs to club
	existing, err := h.GameDayRepo.GetByID(ctx, gameDayID)
	if err != nil {
		if errors.Is(err, gameday.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Spieltag nicht gefunden")
		}
		slog.Error("get game day", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if existing.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Spieltag nicht gefunden")
	}

	date, _ := time.Parse("2006-01-02", req.Date)

	isDraft := existing.IsDraft
	if req.IsDraft != nil {
		isDraft = *req.IsDraft
	}

	publishing := existing.IsDraft && !isDraft
	var updated gameday.GameDay
	if publishing {
		// Status change and every fee share one transaction. On error the day stays a draft.
		updated, err = h.TransactionRepo.SaveGameDayWithCharges(ctx, transaction.SaveGameDayWithChargesParams{
			ID:         gameDayID,
			ClubID:     clubID,
			Date:       date,
			Notes:      req.Notes,
			IsDraft:    false,
			ChargeFees: true,
		})
		if err != nil {
			return gameDayChargeError(err)
		}
	} else {
		updated, err = h.GameDayRepo.Update(ctx, gameday.UpdateGameDayParams{
			ID:      gameDayID,
			Date:    date,
			Notes:   req.Notes,
			IsDraft: isDraft,
		})
		if err != nil {
			slog.Error("update game day", "error", err)
			if errors.Is(err, gameday.ErrNotFound) {
				return fiber.NewError(fiber.StatusNotFound, "Spieltag nicht gefunden")
			}
			return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
		}
	}

	audit.LogAudit(c, u.ID, audit.ActionGameDayUpdated, "club_id", clubID, "gameday_id", gameDayID)
	return c.JSON(GameDayResponseFromEntity(updated))
}

func gameDayChargeError(err error) error {
	var chargeErr *transaction.ChargeError
	if errors.As(err, &chargeErr) {
		return fiber.NewError(fiber.StatusBadRequest, chargeErr.Message)
	}
	if errors.Is(err, gameday.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "Spieltag nicht gefunden")
	}
	slog.Error("save game day with charges", "error", err)
	return fiber.NewError(fiber.StatusInternalServerError, "Die Gebühren konnten nicht vollständig gebucht werden. Es wurde nichts geändert.")
}

func (h *Handler) HandleDeleteGameDay(c fiber.Ctx) error {
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
	gameDayID := c.Params("id")
	if clubID == "" || gameDayID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Spieltag-ID sind erforderlich")
	}

	// Check permission
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeDelete)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Spieltag zu löschen")
	}

	// Verify exists and belongs to club
	existing, err := h.GameDayRepo.GetByID(ctx, gameDayID)
	if err != nil {
		if errors.Is(err, gameday.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Spieltag nicht gefunden")
		}
		slog.Error("get game day", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if existing.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Spieltag nicht gefunden")
	}

	// Collect affected player IDs before delete (CASCADE will remove transactions)
	txs, err := h.TransactionRepo.ListByGameDay(ctx, gameDayID)
	if err != nil {
		slog.Error("list transactions by game day", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	affectedPlayerIDs := make(map[string]struct{})
	for _, tx := range txs {
		if tx.PlayerID != nil && *tx.PlayerID != "" {
			affectedPlayerIDs[*tx.PlayerID] = struct{}{}
		}
	}

	err = h.GameDayRepo.Delete(ctx, gameDayID)
	if err != nil {
		slog.Error("delete game day", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Recalculate balances after CASCADE removed transactions
	for playerID := range affectedPlayerIDs {
		if err := h.TransactionRepo.RecalculatePlayerBalance(ctx, playerID); err != nil {
			slog.Error("recalculate player balance after game day delete", "player_id", playerID, "error", err)
			// Continue with other players and club
		}
	}
	if err := h.TransactionRepo.RecalculateClubBalance(ctx, existing.ClubID); err != nil {
		slog.Error("recalculate club balance after game day delete", "error", err)
		// Already returned 204; balance will be wrong until next recalc
	}

	audit.LogAudit(c, u.ID, audit.ActionGameDayDeleted, "club_id", clubID, "gameday_id", gameDayID)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) HandleAddParticipant(c fiber.Ctx) error {
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
	gameDayID := c.Params("id")
	if clubID == "" || gameDayID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Spieltag-ID sind erforderlich")
	}

	// Check permission
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeUpdate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Teilnehmer hinzuzufügen")
	}

	var req addParticipantRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if req.PlayerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Spieler-ID ist erforderlich")
	}

	// Verify game day exists and belongs to club
	gd, err := h.GameDayRepo.GetByID(ctx, gameDayID)
	if err != nil {
		if errors.Is(err, gameday.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Spieltag nicht gefunden")
		}
		slog.Error("get game day", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if gd.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Spieltag nicht gefunden")
	}

	participant, err := h.GameDayRepo.AddParticipant(ctx, gameDayID, req.PlayerID)
	if err != nil {
		slog.Error("add participant", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionParticipantAdded, "club_id", clubID, "gameday_id", gameDayID, "player_id", req.PlayerID)
	return c.Status(fiber.StatusCreated).JSON(ParticipantResponseFromEntity(participant))
}

func (h *Handler) HandleRemoveParticipant(c fiber.Ctx) error {
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
	gameDayID := c.Params("id")
	playerID := c.Params("playerId")
	if clubID == "" || gameDayID == "" || playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID, Spieltag-ID und Spieler-ID sind erforderlich")
	}

	// Check permission
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeUpdate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Teilnehmer zu entfernen")
	}

	// Verify game day exists and belongs to club
	gd, err := h.GameDayRepo.GetByID(ctx, gameDayID)
	if err != nil {
		if errors.Is(err, gameday.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Spieltag nicht gefunden")
		}
		slog.Error("get game day", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if gd.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Spieltag nicht gefunden")
	}

	err = h.GameDayRepo.RemoveParticipant(ctx, gameDayID, playerID)
	if err != nil {
		if errors.Is(err, gameday.ErrParticipantNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Teilnehmer nicht gefunden")
		}
		slog.Error("remove participant", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionParticipantRemoved, "club_id", clubID, "gameday_id", gameDayID, "player_id", playerID)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) HandleUpdateFees(c fiber.Ctx) error {
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
	gameDayID := c.Params("id")
	playerID := c.Params("playerId")
	if clubID == "" || gameDayID == "" || playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID, Spieltag-ID und Spieler-ID sind erforderlich")
	}

	// Check permission
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeUpdate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Strafen zu bearbeiten")
	}

	var req updateFeesRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	// Verify game day exists and belongs to club
	gd, err := h.GameDayRepo.GetByID(ctx, gameDayID)
	if err != nil {
		if errors.Is(err, gameday.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Spieltag nicht gefunden")
		}
		slog.Error("get game day", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if gd.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Spieltag nicht gefunden")
	}

	// Get participant
	participant, err := h.GameDayRepo.GetParticipantByGameDayAndPlayer(ctx, gameDayID, playerID)
	if err != nil {
		if errors.Is(err, gameday.ErrParticipantNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Teilnehmer nicht gefunden")
		}
		slog.Error("get participant", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	changes, err := h.participantFeeChanges(ctx, clubID, req.Fees)
	if err != nil {
		return err
	}

	// One transaction per participant. The previous per-fee commit loop exceeded
	// the request deadline when every player was saved at once.
	if err := h.TransactionRepo.SyncParticipantFees(ctx, transaction.SyncParticipantFeesParams{
		ClubID:             clubID,
		PlayerID:           playerID,
		GameDayID:          gameDayID,
		ParticipantID:      participant.ID,
		TransactionDate:    gd.Date,
		RecordTransactions: !gd.IsDraft,
		Fees:               changes,
	}); err != nil {
		slog.Error("sync participant fees", "error", err, "player_id", playerID, "gameday_id", gameDayID)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Return updated participant with fees
	fees, err := h.GameDayRepo.GetFeesByParticipant(ctx, participant.ID)
	if err != nil {
		slog.Error("get fees", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionFeesUpdated, "club_id", clubID, "gameday_id", gameDayID, "player_id", playerID)
	return c.JSON(FeesResponseFromEntities(fees))
}

func (h *Handler) participantFeeChanges(ctx context.Context, clubID string, fees []feeInput) ([]transaction.ParticipantFeeChange, error) {
	ordered := make([]feeInput, 0, len(fees))
	index := make(map[string]int, len(fees))
	needsTypes := false
	for _, fee := range fees {
		id := strings.TrimSpace(fee.PenaltyTypeID)
		if id == "" {
			return nil, fiber.NewError(fiber.StatusBadRequest, "Strafentyp-ID ist erforderlich")
		}
		fee.PenaltyTypeID = id
		if i, ok := index[id]; ok {
			ordered[i] = fee
		} else {
			index[id] = len(ordered)
			ordered = append(ordered, fee)
		}
		if fee.Count > 0 && !math.IsNaN(fee.Count) && !math.IsInf(fee.Count, 0) {
			needsTypes = true
		}
	}

	byID := map[string]penaltytype.PenaltyType{}
	if needsTypes {
		penaltyTypes, err := h.PenaltyTypeRepo.GetByClubID(ctx, clubID)
		if err != nil {
			slog.Error("list penalty types", "error", err)
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
		}
		byID = make(map[string]penaltytype.PenaltyType, len(penaltyTypes))
		for _, pt := range penaltyTypes {
			byID[pt.ID] = pt
		}
	}

	changes := make([]transaction.ParticipantFeeChange, 0, len(ordered))
	for _, fee := range ordered {
		if fee.Count <= 0 || math.IsNaN(fee.Count) || math.IsInf(fee.Count, 0) {
			changes = append(changes, transaction.ParticipantFeeChange{
				PenaltyTypeID: fee.PenaltyTypeID,
				Count:         0,
			})
			continue
		}

		pt, ok := byID[fee.PenaltyTypeID]
		if !ok {
			loaded, err := h.PenaltyTypeRepo.GetByID(ctx, fee.PenaltyTypeID)
			if err != nil {
				if errors.Is(err, penaltytype.ErrNotFound) {
					return nil, fiber.NewError(fiber.StatusBadRequest, "Strafentyp "+fee.PenaltyTypeID+" nicht gefunden")
				}
				slog.Error("get penalty type", "error", err)
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
			}
			pt = loaded
		}
		if pt.ClubID != clubID {
			return nil, fiber.NewError(fiber.StatusBadRequest, "Strafentyp gehört nicht zu diesem Klub")
		}
		if pt.DeletedAt != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "Strafentyp "+pt.Name+" ist inaktiv")
		}

		quantityScale := 1
		if pt.AllowsDecimalQuantity {
			quantityScale = 100
		}
		countStored := int(math.Round(fee.Count * float64(quantityScale)))
		if countStored <= 0 {
			continue
		}
		changes = append(changes, transaction.ParticipantFeeChange{
			PenaltyTypeID:          pt.ID,
			PenaltyTypeName:        pt.Name,
			PenaltyTypeDescription: pt.Description,
			PenaltyTypePrice:       pt.Price,
			Count:                  countStored,
			QuantityScale:          quantityScale,
		})
	}
	return changes, nil
}

type updateCompetitionValuesRequest struct {
	Values []competitionValueInput `json:"values"`
}

type competitionValueInput struct {
	CompetitionID string `json:"competition_id"`
	Value         int    `json:"value"`
}

func (h *Handler) HandleUpdateCompetitionValues(c fiber.Ctx) error {
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
	gameDayID := c.Params("id")
	playerID := c.Params("playerId")
	if clubID == "" || gameDayID == "" || playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID, Spieltag-ID und Spieler-ID sind erforderlich")
	}

	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeUpdate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Wettbewerbs-Werte zu bearbeiten")
	}

	gd, err := h.GameDayRepo.GetByID(ctx, gameDayID)
	if err != nil {
		if errors.Is(err, gameday.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Spieltag nicht gefunden")
		}
		slog.Error("get game day", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if gd.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Spieltag nicht gefunden")
	}

	participant, err := h.GameDayRepo.GetParticipantByGameDayAndPlayer(ctx, gameDayID, playerID)
	if err != nil {
		if errors.Is(err, gameday.ErrParticipantNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Teilnehmer nicht gefunden")
		}
		slog.Error("get participant", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	var req updateCompetitionValuesRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	for _, v := range req.Values {
		_, err := h.CompetitionRepo.GetByID(ctx, v.CompetitionID)
		if err != nil {
			if errors.Is(err, competition.ErrNotFound) {
				return fiber.NewError(fiber.StatusBadRequest, "Wettbewerb "+v.CompetitionID+" nicht gefunden")
			}
			slog.Error("get competition", "error", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
		}

		_, err = h.GameDayRepo.UpsertCompetitionValue(ctx, gameday.UpsertCompetitionValueParams{
			ParticipantID: participant.ID,
			CompetitionID: v.CompetitionID,
			Value:         v.Value,
		})
		if err != nil {
			slog.Error("upsert competition value", "error", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
		}
	}

	competitionValues, err := h.GameDayRepo.GetCompetitionValuesByParticipant(ctx, participant.ID)
	if err != nil {
		slog.Error("get competition values", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionCompetitionValues, "club_id", clubID, "gameday_id", gameDayID, "player_id", playerID)
	return c.JSON(CompetitionValuesResponseFromEntities(competitionValues))
}
