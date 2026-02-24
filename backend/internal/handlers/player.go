package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/audit"
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
	Gender       *string `json:"gender"`   // male, female, or nil
	Inactive     *bool   `json:"inactive"` // optional, default false
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

	inactive := false
	if req.Inactive != nil {
		inactive = *req.Inactive
	}
	playerEntity, err := h.PlayerRepo.Create(ctx, player.CreatePlayerParams{
		ClubID:       clubID,
		Name:         strings.TrimSpace(req.Name),
		Balance:      req.Balance,
		StartBalance: req.StartBalance,
		UserID:       req.UserID,
		RoleID:       req.RoleID,
		Gender:       req.Gender,
		Inactive:     inactive,
	})
	if err != nil {
		slog.Error("create player", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionPlayerCreated, "club_id", clubID, "player_id", playerEntity.ID)
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

	responses := PlayersResponseFromEntities(players)
	for i := range responses {
		h.enrichPlayerResponseWithPairBalance(ctx, players[i], &responses[i])
	}
	return c.JSON(responses)
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

	// Allow access to own player; otherwise require view permission
	isOwnPlayer := playerEntity.UserID != nil && *playerEntity.UserID == u.ID
	if !isOwnPlayer {
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
	}

	resp := PlayerResponseFromEntity(playerEntity)
	h.enrichPlayerResponseWithPairBalance(ctx, playerEntity, &resp)
	return c.JSON(resp)
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

	resp := PlayerResponseFromEntity(playerEntity)
	h.enrichPlayerResponseWithPairBalance(ctx, playerEntity, &resp)
	return c.JSON(resp)
}

// enrichPlayerResponseWithPairBalance sets resp.PairBalance to own + partner balance when player has a partner (Paar-Modus).
func (h *Handler) enrichPlayerResponseWithPairBalance(ctx context.Context, p player.Player, resp *PlayerResponse) {
	if p.PartnerID == nil {
		return
	}
	partner, err := h.PlayerRepo.GetByID(ctx, *p.PartnerID)
	if err != nil || partner.ClubID != p.ClubID {
		return
	}
	sum := p.Balance + partner.Balance
	resp.PairBalance = &sum
}

// HandleGetMyPenaltyHistory returns penalty counts per game day for the current user's player (dashboard chart).
// Query param since (YYYY-MM-DD) optional; default is 1 year ago.
func (h *Handler) HandleGetMyPenaltyHistory(c fiber.Ctx) error {
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

	playerEntity, err := h.PlayerRepo.GetByUserIDAndClubID(ctx, u.ID, clubID)
	if err != nil {
		if errors.Is(err, player.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Kein Spieler in diesem Club")
		}
		slog.Error("get my player for penalty history", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if playerEntity.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Kein Spieler in diesem Club")
	}

	since := time.Now().UTC().AddDate(-1, 0, 0) // default 1 year ago
	if sinceParam := c.Query("since"); sinceParam != "" {
		if t, err := time.Parse("2006-01-02", sinceParam); err == nil {
			since = t
		}
	}

	days, err := h.GameDayRepo.GetPlayerPenaltyHistory(ctx, clubID, playerEntity.ID, since)
	if err != nil {
		slog.Error("get player penalty history", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(PenaltyHistoryResponseFromEntities(days))
}

// HandleGetMyCompetitionHistory returns competition values per game day for the current user's player (dashboard chart).
// Query param since (YYYY-MM-DD) optional; default is 1 year ago.
func (h *Handler) HandleGetMyCompetitionHistory(c fiber.Ctx) error {
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

	playerEntity, err := h.PlayerRepo.GetByUserIDAndClubID(ctx, u.ID, clubID)
	if err != nil {
		if errors.Is(err, player.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Kein Spieler in diesem Club")
		}
		slog.Error("get my player for competition history", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if playerEntity.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Kein Spieler in diesem Club")
	}

	since := time.Now().UTC().AddDate(-1, 0, 0) // default 1 year ago
	if sinceParam := c.Query("since"); sinceParam != "" {
		if t, err := time.Parse("2006-01-02", sinceParam); err == nil {
			since = t
		}
	}

	days, err := h.GameDayRepo.GetPlayerCompetitionHistory(ctx, clubID, playerEntity.ID, since)
	if err != nil {
		slog.Error("get player competition history", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(CompetitionHistoryResponseFromEntities(days))
}

type updatePlayerRequest struct {
	Name         string  `json:"name"`
	Balance      int     `json:"balance"`
	StartBalance int     `json:"start_balance"`
	UserID       *string `json:"user_id"`
	RoleID       *string `json:"role_id"`
	Gender       *string `json:"gender"`    // male, female, or nil
	Inactive     *bool   `json:"inactive"`  // optional, keep existing if nil
	PartnerID    *string `json:"partner_id"` // optional; null or omit = keep, uuid = set, empty = clear
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

	// Bestehende UserID/RoleID beibehalten, wenn Request sie nicht mitschickt (verhindert versehentliches NULL)
	userID := req.UserID
	if userID == nil {
		userID = existingPlayer.UserID
	}
	roleID := req.RoleID
	if roleID == nil || *roleID == "" {
		roleID = existingPlayer.RoleID
	}
	inactive := existingPlayer.Inactive
	if req.Inactive != nil {
		inactive = *req.Inactive
	}

	// Partner (Paar-Modus): validate same club and not self; nil = keep, empty string = clear
	partnerID := existingPlayer.PartnerID
	if req.PartnerID != nil {
		if *req.PartnerID == "" {
			partnerID = nil
		} else {
			if *req.PartnerID == playerID {
				return fiber.NewError(fiber.StatusBadRequest, "Spieler kann nicht sich selbst als Partner haben")
			}
			partner, err := h.PlayerRepo.GetByID(ctx, *req.PartnerID)
			if err != nil {
				if errors.Is(err, player.ErrNotFound) {
					return fiber.NewError(fiber.StatusBadRequest, "Partner-Spieler nicht gefunden")
				}
				slog.Error("get partner player", "error", err)
				return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
			}
			if partner.ClubID != clubID {
				return fiber.NewError(fiber.StatusBadRequest, "Partner muss dem gleichen Club angehören")
			}
			partnerID = req.PartnerID
		}
	}

	updatedPlayer, err := h.PlayerRepo.Update(ctx, player.UpdatePlayerParams{
		ID:           playerID,
		Name:         strings.TrimSpace(req.Name),
		Balance:      req.Balance,
		StartBalance: req.StartBalance,
		UserID:       userID,
		RoleID:       roleID,
		Gender:       req.Gender,
		Inactive:     &inactive,
		PartnerID:    partnerID,
	})
	if err != nil {
		if errors.Is(err, player.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Player nicht gefunden")
		}
		slog.Error("update player", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionPlayerUpdated, "club_id", clubID, "player_id", playerID)
	resp := PlayerResponseFromEntity(updatedPlayer)
	h.enrichPlayerResponseWithPairBalance(ctx, updatedPlayer, &resp)
	return c.JSON(resp)
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

	audit.LogAudit(c, u.ID, audit.ActionPlayerDeleted, "club_id", clubID, "player_id", playerID)
	return c.SendStatus(fiber.StatusNoContent)
}

// HandleRecalculatePlayerBalance recalculates player balance from start_balance + SUM(amount) of all transactions. Allowed for own player or with player-update permission.
func (h *Handler) HandleRecalculatePlayerBalance(c fiber.Ctx) error {
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

	// Allow: own player OR has player-update permission (includes club owner)
	isOwnPlayer := existingPlayer.UserID != nil && *existingPlayer.UserID == u.ID
	hasUpdatePermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypePlayers, role.PermissionTypeUpdate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !isOwnPlayer && !hasUpdatePermission {
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, die Player-Balance neu zu berechnen")
	}

	if err := h.TransactionRepo.RecalculatePlayerBalance(ctx, playerID); err != nil {
		slog.Error("recalculate player balance", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	updatedPlayer, err := h.PlayerRepo.GetByID(ctx, playerID)
	if err != nil {
		slog.Error("get player after recalc", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionPlayerBalanceRecalc, "club_id", clubID, "player_id", playerID)
	return c.JSON(PlayerResponseFromEntity(updatedPlayer))
}
