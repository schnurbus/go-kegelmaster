package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/gameday"
	"github.com/schnurbus/go-kegelmaster/backend/internal/penaltytype"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
)

// ==================== REQUEST/RESPONSE TYPES ====================

type createGameDayRequest struct {
	Date  string `json:"date"` // ISO date format YYYY-MM-DD
	Notes string `json:"notes"`
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
	Date  string `json:"date"`
	Notes string `json:"notes"`
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
	PenaltyTypeID string `json:"penalty_type_id"`
	Count         int    `json:"count"`
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
	hasPermission, err := h.PermissionCheck.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeCreate)
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

	gameDayEntity, err := h.GameDayRepo.Create(ctx, gameday.CreateGameDayParams{
		ClubID: clubID,
		Date:   date,
		Notes:  req.Notes,
	})
	if err != nil {
		slog.Error("create game day", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

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
	hasPermission, err := h.PermissionCheck.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeList)
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
	hasPermission, err := h.PermissionCheck.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeView)
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
	hasPermission, err := h.PermissionCheck.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeUpdate)
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

	updated, err := h.GameDayRepo.Update(ctx, gameday.UpdateGameDayParams{
		ID:    gameDayID,
		Date:  date,
		Notes: req.Notes,
	})
	if err != nil {
		slog.Error("update game day", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(GameDayResponseFromEntity(updated))
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
	hasPermission, err := h.PermissionCheck.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeDelete)
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

	err = h.GameDayRepo.Delete(ctx, gameDayID)
	if err != nil {
		slog.Error("delete game day", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

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
	hasPermission, err := h.PermissionCheck.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeUpdate)
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
	hasPermission, err := h.PermissionCheck.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeUpdate)
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
	hasPermission, err := h.PermissionCheck.HasPermission(ctx, u.ID, clubID, role.EntityTypeGameDays, role.PermissionTypeUpdate)
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

	// Process each fee
	for _, fee := range req.Fees {
		if fee.Count <= 0 {
			// Delete fee if count is 0 or negative
			err := h.GameDayRepo.DeleteFee(ctx, participant.ID, fee.PenaltyTypeID)
			if err != nil {
				slog.Error("delete fee", "error", err)
				// Continue processing other fees
			}
			continue
		}

		// Fetch current penalty type to snapshot
		pt, err := h.PenaltyTypeRepo.GetByID(ctx, fee.PenaltyTypeID)
		if err != nil {
			if errors.Is(err, penaltytype.ErrNotFound) {
				return fiber.NewError(fiber.StatusBadRequest, "Strafentyp "+fee.PenaltyTypeID+" nicht gefunden")
			}
			slog.Error("get penalty type", "error", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
		}

		// Verify penalty type belongs to club
		if pt.ClubID != clubID {
			return fiber.NewError(fiber.StatusBadRequest, "Strafentyp gehört nicht zu diesem Klub")
		}

		// Verify penalty type is active (not deleted)
		if pt.DeletedAt != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Strafentyp "+pt.Name+" ist inaktiv")
		}

		// Upsert fee with snapshot
		_, err = h.GameDayRepo.UpsertFee(ctx, gameday.UpsertFeeParams{
			ParticipantID:          participant.ID,
			PenaltyTypeID:          pt.ID,
			PenaltyTypeName:        pt.Name,
			PenaltyTypeDescription: pt.Description,
			PenaltyTypePrice:       pt.Price,
			Count:                  fee.Count,
		})
		if err != nil {
			slog.Error("upsert fee", "error", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
		}
	}

	// Return updated participant with fees
	fees, err := h.GameDayRepo.GetFeesByParticipant(ctx, participant.ID)
	if err != nil {
		slog.Error("get fees", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(FeesResponseFromEntities(fees))
}
