package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
)

type createPlayerRequest struct {
	Name         string  `json:"name"`
	Balance      int     `json:"balance"`
	StartBalance int     `json:"start_balance"`
	UserID       *string `json:"user_id"`
	RoleID       *string `json:"role_id"`
	Gender       *string `json:"gender"` // male, female, or nil
}

func (r createPlayerRequest) validate() error {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return errors.New("Name ist erforderlich")
	}
	if r.RoleID == nil || *r.RoleID == "" {
		return errors.New("Rolle ist erforderlich")
	}
	return nil
}

func (h *Handler) HandleCreatePlayer(c fiber.Ctx) error {
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

	// Check permission: owner OR has create permission for players
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypePlayers, role.PermissionTypeCreate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Player zu erstellen")
	}

	var req createPlayerRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if err := req.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	playerEntity, err := h.PlayerRepo.Create(ctx, player.CreatePlayerParams{
		ClubID:       clubID,
		Name:         strings.TrimSpace(req.Name),
		Balance:      req.Balance,
		StartBalance: req.StartBalance,
		UserID:       req.UserID,
		RoleID:       req.RoleID,
		Gender:       req.Gender,
	})
	if err != nil {
		slog.Error("create player", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.Status(fiber.StatusCreated).JSON(PlayerResponseFromEntity(playerEntity))
}

func (h *Handler) HandleGetPlayers(c fiber.Ctx) error {
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

	// Check permission: owner OR has list permission for players
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypePlayers, role.PermissionTypeList)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Player anzuzeigen")
	}

	players, err := h.PlayerRepo.GetByClubID(ctx, clubID)
	if err != nil {
		slog.Error("get players", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(PlayersResponseFromEntities(players))
}

func (h *Handler) HandleGetPlayer(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("clubId")
	playerID := c.Params("id")
	if clubID == "" || playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Player-ID sind erforderlich")
	}

	// Check permission: owner OR has view permission for players
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypePlayers, role.PermissionTypeView)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Player anzuzeigen")
	}

	playerEntity, err := h.PlayerRepo.GetByID(ctx, playerID)
	if err != nil {
		if errors.Is(err, player.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Player nicht gefunden")
		}
		slog.Error("get player", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Verify player belongs to club
	if playerEntity.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Player nicht gefunden")
	}

	return c.JSON(PlayerResponseFromEntity(playerEntity))
}

func (h *Handler) HandleGetMyPlayer(c fiber.Ctx) error {
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

	// Get player for the current user in the specified club
	playerEntity, err := h.PlayerRepo.GetByUserIDAndClubID(ctx, u.ID, clubID)
	if err != nil {
		if errors.Is(err, player.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Kein Spieler in diesem Club")
		}
		slog.Error("get my player", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Verify player belongs to club
	if playerEntity.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Kein Spieler in diesem Club")
	}

	return c.JSON(PlayerResponseFromEntity(playerEntity))
}

type updatePlayerRequest struct {
	Name         string  `json:"name"`
	Balance      int     `json:"balance"`
	StartBalance int     `json:"start_balance"`
	UserID       *string `json:"user_id"`
	RoleID       *string `json:"role_id"`
	Gender       *string `json:"gender"` // male, female, or nil
}

func (r updatePlayerRequest) validate() error {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return errors.New("Name ist erforderlich")
	}
	if r.RoleID == nil || *r.RoleID == "" {
		return errors.New("Rolle ist erforderlich")
	}
	return nil
}

func (h *Handler) HandleUpdatePlayer(c fiber.Ctx) error {
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
	playerID := c.Params("id")
	if clubID == "" || playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Player-ID sind erforderlich")
	}

	// Check permission: owner OR has update permission for players
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypePlayers, role.PermissionTypeUpdate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Player zu aktualisieren")
	}

	// Verify player exists and belongs to club
	existingPlayer, err := h.PlayerRepo.GetByID(ctx, playerID)
	if err != nil {
		if errors.Is(err, player.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Player nicht gefunden")
		}
		slog.Error("get player", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if existingPlayer.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Player nicht gefunden")
	}

	var req updatePlayerRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if err := req.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	updatedPlayer, err := h.PlayerRepo.Update(ctx, player.UpdatePlayerParams{
		ID:           playerID,
		Name:         strings.TrimSpace(req.Name),
		Balance:      req.Balance,
		StartBalance: req.StartBalance,
		UserID:       req.UserID,
		RoleID:       req.RoleID,
		Gender:       req.Gender,
	})
	if err != nil {
		if errors.Is(err, player.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Player nicht gefunden")
		}
		slog.Error("update player", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(PlayerResponseFromEntity(updatedPlayer))
}

func (h *Handler) HandleDeletePlayer(c fiber.Ctx) error {
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
	playerID := c.Params("id")
	if clubID == "" || playerID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Player-ID sind erforderlich")
	}

	// Check permission: owner OR has delete permission for players
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypePlayers, role.PermissionTypeDelete)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Player zu löschen")
	}

	// Verify player exists and belongs to club
	existingPlayer, err := h.PlayerRepo.GetByID(ctx, playerID)
	if err != nil {
		if errors.Is(err, player.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Player nicht gefunden")
		}
		slog.Error("get player", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if existingPlayer.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Player nicht gefunden")
	}

	if err := h.PlayerRepo.Delete(ctx, playerID); err != nil {
		if errors.Is(err, player.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Player nicht gefunden")
		}
		slog.Error("delete player", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
