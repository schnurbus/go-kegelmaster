package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/audit"
	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/penaltytype"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
)

func (h *Handler) HandleGetPenaltyTypes(c fiber.Ctx) error {
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

	// Check permission: owner OR has list permission for penalty_types
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypePenaltyTypes, role.PermissionTypeList)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Strafentypen einzusehen")
	}

	penaltyTypes, err := h.PenaltyTypeRepo.GetByClubID(ctx, clubID)
	if err != nil {
		slog.Error("get penalty types", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(PenaltyTypesResponseFromEntities(penaltyTypes))
}

func (h *Handler) HandleGetPenaltyType(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("clubId")
	penaltyTypeID := c.Params("id")
	if clubID == "" || penaltyTypeID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Strafentyp-ID sind erforderlich")
	}

	// Check permission: owner OR has view permission for penalty_types
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypePenaltyTypes, role.PermissionTypeView)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Strafentyp einzusehen")
	}

	penaltyType, err := h.PenaltyTypeRepo.GetByID(ctx, penaltyTypeID)
	if err != nil {
		if errors.Is(err, penaltytype.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Strafentyp nicht gefunden")
		}
		slog.Error("get penalty type", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Verify penalty type belongs to club
	if penaltyType.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Strafentyp nicht gefunden")
	}

	return c.JSON(PenaltyTypeResponseFromEntity(penaltyType))
}

type createPenaltyTypeRequest struct {
	Name                  string `json:"name"`
	Description           string `json:"description"`
	Price                 int    `json:"price"`
	DisplayOrder          *int   `json:"display_order,omitempty"`
	AllowsDecimalQuantity bool   `json:"allows_decimal_quantity"`
}

func (r createPenaltyTypeRequest) validate() error {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return errors.New("Name ist erforderlich")
	}
	if r.Price < 0 {
		return errors.New("Preis muss >= 0 sein")
	}
	return nil
}

func (h *Handler) HandleCreatePenaltyType(c fiber.Ctx) error {
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

	// Check permission: owner OR has create permission for penalty_types
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypePenaltyTypes, role.PermissionTypeCreate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Strafentyp zu erstellen")
	}

	var req createPenaltyTypeRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if err := req.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	penaltyType, err := h.PenaltyTypeRepo.Create(ctx, penaltytype.CreatePenaltyTypeParams{
		ClubID:                clubID,
		Name:                  strings.TrimSpace(req.Name),
		Description:           strings.TrimSpace(req.Description),
		Price:                 req.Price,
		DisplayOrder:          req.DisplayOrder,
		AllowsDecimalQuantity: req.AllowsDecimalQuantity,
	})
	if err != nil {
		slog.Error("create penalty type", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionPenaltyTypeCreated, "club_id", clubID, "penalty_type_id", penaltyType.ID)
	return c.Status(fiber.StatusCreated).JSON(PenaltyTypeResponseFromEntity(penaltyType))
}

type updatePenaltyTypeRequest struct {
	Name                  string `json:"name"`
	Description           string `json:"description"`
	Price                 int    `json:"price"`
	AllowsDecimalQuantity bool   `json:"allows_decimal_quantity"`
}

func (r updatePenaltyTypeRequest) validate() error {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return errors.New("Name ist erforderlich")
	}
	if r.Price < 0 {
		return errors.New("Preis muss >= 0 sein")
	}
	return nil
}

func (h *Handler) HandleUpdatePenaltyType(c fiber.Ctx) error {
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
	penaltyTypeID := c.Params("id")
	if clubID == "" || penaltyTypeID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Strafentyp-ID sind erforderlich")
	}

	// Check permission: owner OR has update permission for penalty_types
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypePenaltyTypes, role.PermissionTypeUpdate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Strafentyp zu aktualisieren")
	}

	// Verify penalty type exists and belongs to club
	existingPenaltyType, err := h.PenaltyTypeRepo.GetByID(ctx, penaltyTypeID)
	if err != nil {
		if errors.Is(err, penaltytype.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Strafentyp nicht gefunden")
		}
		slog.Error("get penalty type", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if existingPenaltyType.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Strafentyp nicht gefunden")
	}

	var req updatePenaltyTypeRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if err := req.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Replace creates a new version
	updatedPenaltyType, err := h.PenaltyTypeRepo.Replace(ctx, penaltytype.ReplacePenaltyTypeParams{
		ID:                    penaltyTypeID,
		Name:                  strings.TrimSpace(req.Name),
		Description:           strings.TrimSpace(req.Description),
		Price:                 req.Price,
		AllowsDecimalQuantity: req.AllowsDecimalQuantity,
	})
	if err != nil {
		if errors.Is(err, penaltytype.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Strafentyp nicht gefunden")
		}
		slog.Error("replace penalty type", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionPenaltyTypeUpdated, "club_id", clubID, "penalty_type_id", penaltyTypeID)
	return c.JSON(PenaltyTypeResponseFromEntity(updatedPenaltyType))
}

func (h *Handler) HandleDeletePenaltyType(c fiber.Ctx) error {
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
	penaltyTypeID := c.Params("id")
	if clubID == "" || penaltyTypeID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Strafentyp-ID sind erforderlich")
	}

	// Check permission: owner OR has delete permission for penalty_types
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypePenaltyTypes, role.PermissionTypeDelete)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Strafentyp zu löschen")
	}

	// Verify penalty type exists and belongs to club
	existingPenaltyType, err := h.PenaltyTypeRepo.GetByID(ctx, penaltyTypeID)
	if err != nil {
		if errors.Is(err, penaltytype.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Strafentyp nicht gefunden")
		}
		slog.Error("get penalty type", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if existingPenaltyType.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Strafentyp nicht gefunden")
	}

	if err := h.PenaltyTypeRepo.Delete(ctx, penaltyTypeID); err != nil {
		if errors.Is(err, penaltytype.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Strafentyp nicht gefunden")
		}
		slog.Error("delete penalty type", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionPenaltyTypeDeleted, "club_id", clubID, "penalty_type_id", penaltyTypeID)
	return c.SendStatus(fiber.StatusNoContent)
}

type updateDisplayOrderRequest struct {
	DisplayOrder int `json:"display_order"`
}

func (h *Handler) HandleUpdatePenaltyTypeDisplayOrder(c fiber.Ctx) error {
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
	penaltyTypeID := c.Params("id")
	if clubID == "" || penaltyTypeID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Strafentyp-ID sind erforderlich")
	}

	// Check permission: owner OR has update permission for penalty_types
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypePenaltyTypes, role.PermissionTypeUpdate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Strafentyp zu aktualisieren")
	}

	// Verify penalty type exists and belongs to club
	existingPenaltyType, err := h.PenaltyTypeRepo.GetByID(ctx, penaltyTypeID)
	if err != nil {
		if errors.Is(err, penaltytype.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Strafentyp nicht gefunden")
		}
		slog.Error("get penalty type", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if existingPenaltyType.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Strafentyp nicht gefunden")
	}

	var req updateDisplayOrderRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if req.DisplayOrder < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Display Order muss >= 0 sein")
	}

	err = h.PenaltyTypeRepo.UpdateDisplayOrder(ctx, penaltytype.UpdateDisplayOrderParams{
		ID:           penaltyTypeID,
		DisplayOrder: req.DisplayOrder,
	})
	if err != nil {
		if errors.Is(err, penaltytype.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Strafentyp nicht gefunden")
		}
		slog.Error("update display order", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionPenaltyTypeOrder, "club_id", clubID, "penalty_type_id", penaltyTypeID)
	return c.SendStatus(fiber.StatusNoContent)
}

