package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/competition"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
)

func (h *Handler) HandleGetCompetitions(c fiber.Ctx) error {
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

	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeCompetitions, role.PermissionTypeList)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Wettbewerbe einzusehen")
	}

	competitions, err := h.CompetitionRepo.GetByClubID(ctx, clubID)
	if err != nil {
		slog.Error("get competitions", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(CompetitionsResponseFromEntities(competitions))
}

func (h *Handler) HandleGetCompetition(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("clubId")
	competitionID := c.Params("id")
	if clubID == "" || competitionID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Wettbewerbs-ID sind erforderlich")
	}

	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeCompetitions, role.PermissionTypeView)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Wettbewerb einzusehen")
	}

	comp, err := h.CompetitionRepo.GetByID(ctx, competitionID)
	if err != nil {
		if errors.Is(err, competition.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Wettbewerb nicht gefunden")
		}
		slog.Error("get competition", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if comp.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Wettbewerb nicht gefunden")
	}

	return c.JSON(CompetitionResponseFromEntity(comp))
}

type createCompetitionRequest struct {
	Name             string `json:"name"`
	ScoringType      string `json:"scoring_type"`       // winner, loser, both
	IsGenderSpecific bool   `json:"is_gender_specific"`
	DisplayOrder     *int   `json:"display_order,omitempty"`
}

func (r createCompetitionRequest) validate() error {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return errors.New("Name ist erforderlich")
	}
	switch r.ScoringType {
	case "winner", "loser", "both":
		// valid
	default:
		return errors.New("Wertung muss winner, loser oder both sein")
	}
	return nil
}

func (h *Handler) HandleCreateCompetition(c fiber.Ctx) error {
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

	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeCompetitions, role.PermissionTypeCreate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Wettbewerb zu erstellen")
	}

	var req createCompetitionRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if err := req.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	comp, err := h.CompetitionRepo.Create(ctx, competition.CreateCompetitionParams{
		ClubID:           clubID,
		Name:             strings.TrimSpace(req.Name),
		ScoringType:      req.ScoringType,
		IsGenderSpecific: req.IsGenderSpecific,
		DisplayOrder:     req.DisplayOrder,
	})
	if err != nil {
		slog.Error("create competition", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.Status(fiber.StatusCreated).JSON(CompetitionResponseFromEntity(comp))
}

type updateCompetitionRequest struct {
	Name             string `json:"name"`
	ScoringType      string `json:"scoring_type"`
	IsGenderSpecific bool   `json:"is_gender_specific"`
	DisplayOrder     int    `json:"display_order"`
}

func (r updateCompetitionRequest) validate() error {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return errors.New("Name ist erforderlich")
	}
	switch r.ScoringType {
	case "winner", "loser", "both":
		// valid
	default:
		return errors.New("Wertung muss winner, loser oder both sein")
	}
	return nil
}

func (h *Handler) HandleUpdateCompetition(c fiber.Ctx) error {
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
	competitionID := c.Params("id")
	if clubID == "" || competitionID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Wettbewerbs-ID sind erforderlich")
	}

	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeCompetitions, role.PermissionTypeUpdate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Wettbewerb zu aktualisieren")
	}

	existing, err := h.CompetitionRepo.GetByID(ctx, competitionID)
	if err != nil {
		if errors.Is(err, competition.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Wettbewerb nicht gefunden")
		}
		slog.Error("get competition", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if existing.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Wettbewerb nicht gefunden")
	}

	var req updateCompetitionRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if err := req.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	comp, err := h.CompetitionRepo.Update(ctx, competition.UpdateCompetitionParams{
		ID:               competitionID,
		Name:             strings.TrimSpace(req.Name),
		ScoringType:      req.ScoringType,
		IsGenderSpecific: req.IsGenderSpecific,
		DisplayOrder:     req.DisplayOrder,
	})
	if err != nil {
		if errors.Is(err, competition.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Wettbewerb nicht gefunden")
		}
		slog.Error("update competition", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(CompetitionResponseFromEntity(comp))
}

func (h *Handler) HandleDeleteCompetition(c fiber.Ctx) error {
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
	competitionID := c.Params("id")
	if clubID == "" || competitionID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Wettbewerbs-ID sind erforderlich")
	}

	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeCompetitions, role.PermissionTypeDelete)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Wettbewerb zu löschen")
	}

	existing, err := h.CompetitionRepo.GetByID(ctx, competitionID)
	if err != nil {
		if errors.Is(err, competition.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Wettbewerb nicht gefunden")
		}
		slog.Error("get competition", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if existing.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Wettbewerb nicht gefunden")
	}

	if err := h.CompetitionRepo.Delete(ctx, competitionID); err != nil {
		if errors.Is(err, competition.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Wettbewerb nicht gefunden")
		}
		slog.Error("delete competition", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
