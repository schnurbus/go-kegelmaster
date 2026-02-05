package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
)

func (h *Handler) HandleGetRoles(c fiber.Ctx) error {
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

	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeRoles, role.PermissionTypeList)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check roles list permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Rollen einzusehen")
	}

	roles, err := h.RoleRepo.GetByClubID(ctx, clubID)
	if err != nil {
		slog.Error("get roles", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	playerCountByRole, err := h.PlayerRepo.CountByClubIDGroupByRoleID(ctx, clubID)
	if err != nil {
		slog.Error("count players by role", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Load permissions for each role
	rolesWithPerms := make([]RoleResponse, 0, len(roles))
	for _, r := range roles {
		perms, err := h.RoleRepo.GetPermissions(ctx, r.ID)
		if err != nil {
			slog.Error("get permissions", "error", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
		}
		count := playerCountByRole[r.ID]
		rolesWithPerms = append(rolesWithPerms, RoleResponseFromEntity(r, perms, count))
	}

	return c.JSON(rolesWithPerms)
}

func (h *Handler) HandleGetRole(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("clubId")
	roleID := c.Params("id")
	if clubID == "" || roleID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Rollen-ID sind erforderlich")
	}

	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeRoles, role.PermissionTypeView)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check roles view permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Rollen einzusehen")
	}

	roleEntity, err := h.RoleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, role.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Rolle nicht gefunden")
		}
		slog.Error("get role", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Verify role belongs to club
	if roleEntity.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Rolle nicht gefunden")
	}

	perms, err := h.RoleRepo.GetPermissions(ctx, roleID)
	if err != nil {
		slog.Error("get permissions", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(RoleResponseFromEntity(roleEntity, perms, 0))
}

type createRoleRequest struct {
	Name        string                    `json:"name"`
	PaysBaseFee bool                      `json:"pays_base_fee"`
	Permissions []createPermissionRequest `json:"permissions,omitempty"`
}

type createPermissionRequest struct {
	EntityType     string `json:"entity_type"`
	PermissionType string `json:"permission_type"`
}

func (r createRoleRequest) validate() error {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return errors.New("Name ist erforderlich")
	}
	return nil
}

func (h *Handler) HandleCreateRole(c fiber.Ctx) error {
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

	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeRoles, role.PermissionTypeCreate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check roles create permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Rollen zu erstellen")
	}

	var req createRoleRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if err := req.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	roleEntity, err := h.RoleRepo.Create(ctx, role.CreateRoleParams{
		ClubID:      clubID,
		Name:        strings.TrimSpace(req.Name),
		PaysBaseFee: req.PaysBaseFee,
	})
	if err != nil {
		slog.Error("create role", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Add permissions if provided
	perms := make([]role.RolePermission, 0)
	for _, permReq := range req.Permissions {
		entityType := role.EntityType(permReq.EntityType)
		permissionType := role.PermissionType(permReq.PermissionType)

		if !entityType.IsValid() {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Ungültiger Entity-Type: %s", permReq.EntityType))
		}
		if !permissionType.IsValid() {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Ungültiger Permission-Type: %s", permReq.PermissionType))
		}

		perm, err := h.RoleRepo.AddPermission(ctx, role.AddPermissionParams{
			RoleID:         roleEntity.ID,
			EntityType:     entityType,
			PermissionType: permissionType,
		})
		if err != nil {
			if errors.Is(err, role.ErrPermissionExist) {
				continue // Skip duplicate permissions
			}
			slog.Error("add permission", "error", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
		}
		perms = append(perms, perm)
	}

	return c.Status(fiber.StatusCreated).JSON(RoleResponseFromEntity(roleEntity, perms, 0))
}

type updateRoleRequest struct {
	Name        string `json:"name"`
	PaysBaseFee bool   `json:"pays_base_fee"`
}

func (r updateRoleRequest) validate() error {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return errors.New("Name ist erforderlich")
	}
	return nil
}

func (h *Handler) HandleUpdateRole(c fiber.Ctx) error {
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
	roleID := c.Params("id")
	if clubID == "" || roleID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Rollen-ID sind erforderlich")
	}

	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeRoles, role.PermissionTypeUpdate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check roles update permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Rollen zu aktualisieren")
	}

	// Verify role exists and belongs to club
	existingRole, err := h.RoleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, role.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Rolle nicht gefunden")
		}
		slog.Error("get role", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if existingRole.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Rolle nicht gefunden")
	}

	var req updateRoleRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if err := req.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	updatedRole, err := h.RoleRepo.Update(ctx, role.UpdateRoleParams{
		ID:          roleID,
		Name:        strings.TrimSpace(req.Name),
		PaysBaseFee: req.PaysBaseFee,
	})
	if err != nil {
		if errors.Is(err, role.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Rolle nicht gefunden")
		}
		slog.Error("update role", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	perms, err := h.RoleRepo.GetPermissions(ctx, roleID)
	if err != nil {
		slog.Error("get permissions", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(RoleResponseFromEntity(updatedRole, perms, 0))
}

func (h *Handler) HandleDeleteRole(c fiber.Ctx) error {
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
	roleID := c.Params("id")
	if clubID == "" || roleID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Rollen-ID sind erforderlich")
	}

	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeRoles, role.PermissionTypeDelete)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check roles delete permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Rollen zu löschen")
	}

	// Verify role exists and belongs to club
	existingRole, err := h.RoleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, role.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Rolle nicht gefunden")
		}
		slog.Error("get role", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if existingRole.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Rolle nicht gefunden")
	}

	if err := h.RoleRepo.Delete(ctx, roleID); err != nil {
		if errors.Is(err, role.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Rolle nicht gefunden")
		}
		slog.Error("delete role", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

type addPermissionRequest struct {
	EntityType     string `json:"entity_type"`
	PermissionType string `json:"permission_type"`
}

func (h *Handler) HandleAddPermission(c fiber.Ctx) error {
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
	roleID := c.Params("id")
	if clubID == "" || roleID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Rollen-ID sind erforderlich")
	}

	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeRoles, role.PermissionTypeUpdate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check roles update permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Rollen zu bearbeiten")
	}

	// Verify role exists and belongs to club
	existingRole, err := h.RoleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, role.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Rolle nicht gefunden")
		}
		slog.Error("get role", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if existingRole.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Rolle nicht gefunden")
	}

	var req addPermissionRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	entityType := role.EntityType(req.EntityType)
	permissionType := role.PermissionType(req.PermissionType)

	if !entityType.IsValid() {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Ungültiger Entity-Type: %s", req.EntityType))
	}
	if !permissionType.IsValid() {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Ungültiger Permission-Type: %s", req.PermissionType))
	}

	perm, err := h.RoleRepo.AddPermission(ctx, role.AddPermissionParams{
		RoleID:         roleID,
		EntityType:     entityType,
		PermissionType: permissionType,
	})
	if err != nil {
		if errors.Is(err, role.ErrPermissionExist) {
			return fiber.NewError(fiber.StatusConflict, "Berechtigung existiert bereits")
		}
		slog.Error("add permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.Status(fiber.StatusCreated).JSON(PermissionResponseFromEntity(perm))
}

type removePermissionRequest struct {
	EntityType     string `json:"entity_type"`
	PermissionType string `json:"permission_type"`
}

func (h *Handler) HandleRemovePermission(c fiber.Ctx) error {
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
	roleID := c.Params("id")
	if clubID == "" || roleID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Rollen-ID sind erforderlich")
	}

	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypeRoles, role.PermissionTypeUpdate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check roles update permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Rollen zu bearbeiten")
	}

	// Verify role exists and belongs to club
	existingRole, err := h.RoleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, role.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Rolle nicht gefunden")
		}
		slog.Error("get role", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if existingRole.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Rolle nicht gefunden")
	}

	var req removePermissionRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	entityType := role.EntityType(req.EntityType)
	permissionType := role.PermissionType(req.PermissionType)

	if !entityType.IsValid() {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Ungültiger Entity-Type: %s", req.EntityType))
	}
	if !permissionType.IsValid() {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Ungültiger Permission-Type: %s", req.PermissionType))
	}

	if err := h.RoleRepo.RemovePermission(ctx, role.RemovePermissionParams{
		RoleID:         roleID,
		EntityType:     entityType,
		PermissionType: permissionType,
	}); err != nil {
		if errors.Is(err, role.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Berechtigung nicht gefunden")
		}
		slog.Error("remove permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

