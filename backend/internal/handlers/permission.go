package handlers

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
)

// PermissionEntry is a single permission for the "my permissions" response.
type PermissionEntry struct {
	EntityType     string `json:"entity_type"`
	PermissionType string `json:"permission_type"`
}

// MyPermissionsResponse is the response for GET /api/clubs/:clubId/permissions/me
type MyPermissionsResponse struct {
	IsOwner     bool              `json:"is_owner"`
	Permissions []PermissionEntry `json:"permissions"`
}

// HandleGetMyPermissions returns the current user's permissions for a club.
// Owner gets all permissions; others get permissions from their player's role.
func (h *Handler) HandleGetMyPermissions(c fiber.Ctx) error {
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

	isOwner, err := h.PermissionChecker.IsClubOwner(ctx, u.ID, clubID)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check club owner", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if isOwner {
		// Owner has all permissions: build full list
		perms := make([]PermissionEntry, 0)
		for _, et := range role.ValidEntityTypes() {
			for _, pt := range role.ValidPermissionTypes() {
				perms = append(perms, PermissionEntry{
					EntityType:     string(et),
					PermissionType: string(pt),
				})
			}
		}
		return c.JSON(MyPermissionsResponse{IsOwner: true, Permissions: perms})
	}

	// Not owner: collect permissions from role
	playerEntity, err := h.PlayerRepo.GetByUserIDAndClubID(ctx, u.ID, clubID)
	if err != nil {
		if errors.Is(err, player.ErrNotFound) {
			return c.JSON(MyPermissionsResponse{IsOwner: false, Permissions: []PermissionEntry{}})
		}
		slog.Error("get my player", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if playerEntity.RoleID == nil {
		return c.JSON(MyPermissionsResponse{IsOwner: false, Permissions: []PermissionEntry{}})
	}

	// One query for the role's permissions instead of a HasPermission call
	// (three queries each) for every entity/permission combination.
	rolePerms, err := h.RoleRepo.GetPermissions(ctx, *playerEntity.RoleID)
	if err != nil {
		slog.Error("get role permissions", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	granted := make(map[[2]string]struct{}, len(rolePerms))
	for _, rp := range rolePerms {
		granted[[2]string{string(rp.EntityType), string(rp.PermissionType)}] = struct{}{}
	}
	perms := make([]PermissionEntry, 0, len(granted))
	for _, et := range role.ValidEntityTypes() {
		for _, pt := range role.ValidPermissionTypes() {
			if _, ok := granted[[2]string{string(et), string(pt)}]; ok {
				perms = append(perms, PermissionEntry{
					EntityType:     string(et),
					PermissionType: string(pt),
				})
			}
		}
	}
	return c.JSON(MyPermissionsResponse{IsOwner: false, Permissions: perms})
}
