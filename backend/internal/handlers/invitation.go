package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/invitation"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
)

type invitePlayerRequest struct {
	Email string `json:"email"`
}

func (r invitePlayerRequest) validate() error {
	email := strings.TrimSpace(strings.ToLower(r.Email))
	if email == "" {
		return errors.New("E-Mail ist erforderlich")
	}
	// Basic email validation
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return errors.New("Ungültige E-Mail-Adresse")
	}
	return nil
}

// HandleInvitePlayer sends an invitation email to link a user to a player.
func (h *Handler) HandleInvitePlayer(c fiber.Ctx) error {
	slog.Info("handle invite player request started")
	
	if err := h.ValidateCSRF(c); err != nil {
		slog.Warn("CSRF validation failed", "error", err)
		return err
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		slog.Warn("user not authenticated", "error", err)
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("clubId")
	playerID := c.Params("id")
	slog.Info("invite player request", "user_id", u.ID, "club_id", clubID, "player_id", playerID)
	
	if clubID == "" || playerID == "" {
		slog.Warn("missing parameters", "club_id", clubID, "player_id", playerID)
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID und Player-ID sind erforderlich")
	}

	// Check permission: owner OR has update permission for players
	slog.Debug("checking permissions", "user_id", u.ID, "club_id", clubID)
	hasPermission, err := h.PermissionChecker.HasPermission(ctx, u.ID, clubID, role.EntityTypePlayers, role.PermissionTypeUpdate)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			slog.Warn("club not found", "club_id", clubID)
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("check permission failed", "error", err, "user_id", u.ID, "club_id", clubID)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if !hasPermission {
		slog.Warn("permission denied", "user_id", u.ID, "club_id", clubID, "player_id", playerID)
		return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, Player einzuladen")
	}
	slog.Debug("permission granted", "user_id", u.ID, "club_id", clubID)

	// Verify player exists and belongs to club
	slog.Debug("fetching player", "player_id", playerID)
	existingPlayer, err := h.PlayerRepo.GetByID(ctx, playerID)
	if err != nil {
		if errors.Is(err, player.ErrNotFound) {
			slog.Warn("player not found", "player_id", playerID)
			return fiber.NewError(fiber.StatusNotFound, "Player nicht gefunden")
		}
		slog.Error("get player failed", "error", err, "player_id", playerID)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if existingPlayer.ClubID != clubID {
		slog.Warn("player club mismatch", "player_id", playerID, "player_club", existingPlayer.ClubID, "requested_club", clubID)
		return fiber.NewError(fiber.StatusNotFound, "Player nicht gefunden")
	}

	// Check if player already has a user
	if existingPlayer.UserID != nil {
		slog.Warn("player already has user", "player_id", playerID, "user_id", *existingPlayer.UserID)
		return fiber.NewError(fiber.StatusBadRequest, "Player hat bereits einen zugewiesenen User")
	}

	var req invitePlayerRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		slog.Warn("invalid request body", "error", err)
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if err := req.validate(); err != nil {
		slog.Warn("request validation failed", "error", err, "email", req.Email)
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	
	slog.Info("creating invitation", "player_id", playerID, "email", req.Email)

	// Create invitation
	inv, err := h.InvitationRepo.Create(ctx, invitation.CreateInvitationParams{
		PlayerID: playerID,
		Email:    strings.TrimSpace(strings.ToLower(req.Email)),
	})
	if err != nil {
		slog.Error("create invitation failed", "error", err, "player_id", playerID, "email", req.Email)
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Erstellen der Einladung")
	}
	slog.Info("invitation created", "invitation_id", inv.ID, "player_id", playerID, "email", inv.Email, "expires_at", inv.ExpiresAt)

	// Get club name for email (optional, for better email content)
	clubEntity, err := h.ClubRepo.GetByID(ctx, clubID)
	if err != nil {
		slog.Warn("could not fetch club for email", "error", err, "club_id", clubID)
		// Continue anyway, we can send email without club name
	} else {
		slog.Debug("club fetched for email", "club_name", clubEntity.Name)
	}

	// Build invitation link
	invitationLink := h.Config.BaseURL + "/invite/" + inv.Token
	slog.Debug("invitation link generated", "link", invitationLink)

	// Send email
	slog.Info("sending invitation email", "email", inv.Email, "player", existingPlayer.Name, "link", invitationLink)
	if h.EmailSvc == nil {
		slog.Error("email service not initialized", "invitation_id", inv.ID)
		return c.JSON(fiber.Map{
			"invitation_id": inv.ID,
			"email":         inv.Email,
			"expires_at":    inv.ExpiresAt,
			"message":       "Einladung wurde erstellt, aber E-Mail konnte nicht versendet werden",
			"warning":       "E-Mail-Service ist nicht konfiguriert. Bitte überprüfen Sie RESEND_API_KEY und RESEND_FROM_EMAIL",
		})
	}
	
	if err := h.EmailSvc.SendInvitationEmail(
		inv.Email,
		existingPlayer.Name,
		invitationLink,
	); err != nil {
		slog.Error("send invitation email failed", 
			"error", err, 
			"invitation_id", inv.ID,
			"email", inv.Email,
			"player", existingPlayer.Name,
			"link", invitationLink)
		// Don't fail the request, invitation was created
		// Return success but include a warning in the response
		return c.JSON(fiber.Map{
			"invitation_id": inv.ID,
			"email":         inv.Email,
			"expires_at":    inv.ExpiresAt,
			"message":       "Einladung wurde erstellt, aber E-Mail konnte nicht versendet werden",
			"warning":       err.Error(),
		})
	}

	slog.Info("invite player request completed successfully", 
		"invitation_id", inv.ID, 
		"player_id", playerID, 
		"email", inv.Email)
	
	return c.JSON(fiber.Map{
		"invitation_id": inv.ID,
		"email":         inv.Email,
		"expires_at":    inv.ExpiresAt,
		"message":       "Einladung wurde erfolgreich versendet",
	})
}

// HandleGetInvitation retrieves invitation details by token.
func (h *Handler) HandleGetInvitation(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	token := c.Params("token")
	tokenPreview := token
	if len(token) > 8 {
		tokenPreview = token[:8] + "..."
	}
	slog.Info("get invitation request", "token", tokenPreview)
	
	if token == "" {
		slog.Warn("empty token in request")
		return fiber.NewError(fiber.StatusBadRequest, "Token ist erforderlich")
	}

	inv, err := h.InvitationRepo.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, invitation.ErrNotFound) {
			slog.Warn("invitation not found", "token", tokenPreview)
			return fiber.NewError(fiber.StatusNotFound, "Einladung nicht gefunden")
		}
		if errors.Is(err, invitation.ErrExpired) {
			slog.Warn("invitation expired", "invitation_id", inv.ID)
			return fiber.NewError(fiber.StatusGone, "Einladung ist abgelaufen")
		}
		if errors.Is(err, invitation.ErrAlreadyAccepted) {
			slog.Warn("invitation already accepted", "invitation_id", inv.ID)
			return fiber.NewError(fiber.StatusGone, "Einladung wurde bereits akzeptiert")
		}
		slog.Error("get invitation failed", "error", err, "token", tokenPreview)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	slog.Info("invitation found", "invitation_id", inv.ID, "player_id", inv.PlayerID, "email", inv.Email)

	// Get player details
	playerEntity, err := h.PlayerRepo.GetByID(ctx, inv.PlayerID)
	if err != nil {
		slog.Error("get player failed", "error", err, "player_id", inv.PlayerID)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Get club details
	clubEntity, err := h.ClubRepo.GetByID(ctx, playerEntity.ClubID)
	if err != nil {
		slog.Error("get club failed", "error", err, "club_id", playerEntity.ClubID)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	slog.Info("invitation details retrieved", "invitation_id", inv.ID, "player", playerEntity.Name, "club", clubEntity.Name)
	return c.JSON(fiber.Map{
		"player_id":   inv.PlayerID,
		"player_name": playerEntity.Name,
		"club_id":     clubEntity.ID,
		"club_name":   clubEntity.Name,
		"email":       inv.Email,
		"expires_at":  inv.ExpiresAt,
	})
}

// HandleAcceptInvitation accepts an invitation and links the player to the current user.
func (h *Handler) HandleAcceptInvitation(c fiber.Ctx) error {
	slog.Info("handle accept invitation request started")
	
	if err := h.ValidateCSRF(c); err != nil {
		slog.Warn("CSRF validation failed", "error", err)
		return err
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		slog.Warn("user not authenticated", "error", err)
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	token := c.Params("token")
	tokenPreview := token
	if len(token) > 8 {
		tokenPreview = token[:8] + "..."
	}
	slog.Info("accept invitation request", "user_id", u.ID, "token", tokenPreview)
	
	if token == "" {
		slog.Warn("empty token in accept request")
		return fiber.NewError(fiber.StatusBadRequest, "Token ist erforderlich")
	}

	// Get invitation to validate
	inv, err := h.InvitationRepo.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, invitation.ErrNotFound) {
			slog.Warn("invitation not found for accept", "token", tokenPreview)
			return fiber.NewError(fiber.StatusNotFound, "Einladung nicht gefunden")
		}
		if errors.Is(err, invitation.ErrExpired) {
			slog.Warn("invitation expired for accept", "invitation_id", inv.ID)
			return fiber.NewError(fiber.StatusGone, "Einladung ist abgelaufen")
		}
		if errors.Is(err, invitation.ErrAlreadyAccepted) {
			slog.Warn("invitation already accepted", "invitation_id", inv.ID)
			return fiber.NewError(fiber.StatusGone, "Einladung wurde bereits akzeptiert")
		}
		slog.Error("get invitation failed for accept", "error", err, "token", tokenPreview)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	slog.Info("invitation validated", "invitation_id", inv.ID, "player_id", inv.PlayerID, "email", inv.Email)

	// Verify email matches (optional security check)
	// In a real scenario, you might want to verify the user's email matches the invitation email
	// For now, we'll allow any authenticated user to accept
	slog.Debug("accepting invitation", "invitation_id", inv.ID, "user_id", u.ID, "player_id", inv.PlayerID)

	// Accept invitation (this will update the player's user_id)
	if err := h.InvitationRepo.Accept(ctx, inv.ID, u.ID); err != nil {
		slog.Error("accept invitation failed", 
			"error", err, 
			"invitation_id", inv.ID, 
			"user_id", u.ID,
			"player_id", inv.PlayerID)
		return fiber.NewError(fiber.StatusInternalServerError, "Fehler beim Akzeptieren der Einladung")
	}

	slog.Info("invitation accepted successfully", 
		"invitation_id", inv.ID, 
		"user_id", u.ID,
		"player_id", inv.PlayerID,
		"email", inv.Email)
	
	return c.JSON(fiber.Map{
		"player_id": inv.PlayerID,
		"message":   "Einladung erfolgreich akzeptiert",
	})
}
