package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
)

type createClubRequest struct {
	Name           string `json:"name"`
	Balance        int    `json:"balance"`
	BaseFee        int    `json:"base_fee"`
	AutoTipEnabled *bool  `json:"auto_tip_enabled,omitempty"` // Optional, defaults to true
}

func (r createClubRequest) validate() error {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return errors.New("Name ist erforderlich")
	}
	if r.Balance < 0 {
		return errors.New("Balance darf nicht negativ sein")
	}
	if r.BaseFee < 0 {
		return errors.New("BaseFee darf nicht negativ sein")
	}
	return nil
}

func (h *Handler) HandleCreateClub(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	var req createClubRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if err := req.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Default auto_tip_enabled to true if not provided
	autoTipEnabled := true
	if req.AutoTipEnabled != nil {
		autoTipEnabled = *req.AutoTipEnabled
	}

	clubEntity, err := h.ClubRepo.Create(ctx, club.CreateClubParams{
		Name:           strings.TrimSpace(req.Name),
		Balance:        req.Balance,
		BaseFee:        req.BaseFee,
		AutoTipEnabled: autoTipEnabled,
		UserID:         u.ID,
	})
	if err != nil {
		slog.Error("create club", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.Status(fiber.StatusCreated).JSON(ClubResponseFromEntity(clubEntity))
}

func (h *Handler) HandleGetClubs(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubs, err := h.ClubRepo.GetForUser(ctx, u.ID)
	if err != nil {
		slog.Error("get clubs", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(ClubsResponseFromEntities(clubs))
}

func (h *Handler) HandleGetClub(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	if _, err := h.UserFromCookie(ctx, c); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("id")
	if clubID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID ist erforderlich")
	}

	clubEntity, err := h.ClubRepo.GetByID(ctx, clubID)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("get club", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(ClubResponseFromEntity(clubEntity))
}

type updateClubRequest struct {
	Name           string `json:"name"`
	Balance        int    `json:"balance"`
	BaseFee        int    `json:"base_fee"`
	AutoTipEnabled bool   `json:"auto_tip_enabled"`
}

func (r updateClubRequest) validate() error {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return errors.New("Name ist erforderlich")
	}
	if r.Balance < 0 {
		return errors.New("Balance darf nicht negativ sein")
	}
	if r.BaseFee < 0 {
		return errors.New("BaseFee darf nicht negativ sein")
	}
	return nil
}

func (h *Handler) HandleUpdateClub(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("id")
	if clubID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID ist erforderlich")
	}

	// Prüfe, ob der Club existiert und der User der Owner ist
	existingClub, err := h.ClubRepo.GetByID(ctx, clubID)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("get club", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Owner-Check: Nur der Owner darf updaten
	if existingClub.UserID != u.ID {
		return fiber.NewError(fiber.StatusForbidden, "Nur der Owner darf diesen Club aktualisieren")
	}

	var req updateClubRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if err := req.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	updatedClub, err := h.ClubRepo.Update(ctx, club.UpdateClubParams{
		ID:             clubID,
		Name:           strings.TrimSpace(req.Name),
		Balance:        req.Balance,
		BaseFee:        req.BaseFee,
		AutoTipEnabled: req.AutoTipEnabled,
	})
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("update club", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(ClubResponseFromEntity(updatedClub))
}

func (h *Handler) HandleDeleteClub(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("id")
	if clubID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID ist erforderlich")
	}

	// Prüfe, ob der Club existiert und der User der Owner ist
	existingClub, err := h.ClubRepo.GetByID(ctx, clubID)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("get club", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Owner-Check: Nur der Owner darf löschen
	if existingClub.UserID != u.ID {
		return fiber.NewError(fiber.StatusForbidden, "Nur der Owner darf diesen Club löschen")
	}

	if err := h.ClubRepo.Delete(ctx, clubID); err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("delete club", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
