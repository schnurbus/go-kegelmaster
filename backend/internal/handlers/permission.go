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

	perms := make([]PermissionEntry, 0)
	for _, et := range role.ValidEntityTypes() {
		for _, pt := range role.ValidPermissionTypes() {
			has, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, et, pt)
			if err != nil {
				slog.Error("check permission", "entity_type", et, "permission_type", pt, "error", err)
				continue
			}
			if has {
				perms = append(perms, PermissionEntry{
					EntityType:     string(et),
					PermissionType: string(pt),
				})
			}
		}
	}
	return c.JSON(MyPermissionsResponse{IsOwner: false, Permissions: perms})
}
