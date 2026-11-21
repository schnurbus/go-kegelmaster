package handlers

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/audit"
	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
)

// Kennungen der Schritte der Ersteinrichtung, in der angezeigten Reihenfolge.
const (
	SetupStepClub         = "club"
	SetupStepPenaltyTypes = "penalty_types"
	SetupStepRoles        = "roles"
	SetupStepPlayers      = "players"
	SetupStepFirstGameDay = "first_game_day"
	SetupStepOwnPlayer    = "own_player"
	SetupStepInvitations  = "invitations"
	SetupStepCompetitions = "competitions"
)

type SetupStepResponse struct {
	ID       string `json:"id"`
	Required bool   `json:"required"`
	Done     bool   `json:"done"`
}

// ClubSetupResponse is the state of the Ersteinrichtung of a club. It is
// derived from the club's data on every request and never stored.
type ClubSetupResponse struct {
	// Dismissed: the owner hid the card.
	Dismissed bool `json:"dismissed"`
	// Complete: all required steps are done.
	Complete bool                `json:"complete"`
	Steps    []SetupStepResponse `json:"steps"`
}

func clubSetupFromCounts(cl club.Club, n club.SetupCounts) ClubSetupResponse {
	steps := []SetupStepResponse{
		{ID: SetupStepClub, Required: true, Done: true},
		{ID: SetupStepPenaltyTypes, Required: true, Done: n.PenaltyTypes > 0},
		{ID: SetupStepRoles, Required: true, Done: n.Roles > 0},
		{ID: SetupStepPlayers, Required: true, Done: n.Players > 0},
		// Only a published game day books anything; drafts do not count.
		{ID: SetupStepFirstGameDay, Required: true, Done: n.PublishedGameDays > 0},
		{ID: SetupStepOwnPlayer, Required: false, Done: n.OwnPlayers > 0},
		{ID: SetupStepInvitations, Required: false, Done: n.Invitations > 0},
		{ID: SetupStepCompetitions, Required: false, Done: n.Competitions > 0},
	}
	complete := true
	for _, s := range steps {
		if s.Required && !s.Done {
			complete = false
		}
	}
	return ClubSetupResponse{Dismissed: cl.SetupDismissedAt != nil, Complete: complete, Steps: steps}
}

// ownedClub loads the club from the :id parameter and checks that the
// signed-in user is its owner (Besitzer). Errors are *fiber.Error.
func (h *Handler) ownedClub(c fiber.Ctx, userID string) (club.Club, error) {
	ctx, cancel := h.RequestContext()
	defer cancel()

	clubID := c.Params("id")
	if clubID == "" {
		return club.Club{}, fiber.NewError(fiber.StatusBadRequest, "Club-ID ist erforderlich")
	}
	cl, err := h.ClubRepo.GetByID(ctx, clubID)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return club.Club{}, fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("get club", "error", err)
		return club.Club{}, fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if cl.UserID != userID {
		return club.Club{}, fiber.NewError(fiber.StatusForbidden, "Nur der Besitzer sieht die Ersteinrichtung")
	}
	return cl, nil
}

// HandleGetClubSetup returns the state of the Ersteinrichtung (owner only).
func (h *Handler) HandleGetClubSetup(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}
	cl, err := h.ownedClub(c, u.ID)
	if err != nil {
		return err
	}
	counts, err := h.ClubRepo.SetupCounts(ctx, cl.ID, u.ID)
	if err != nil {
		slog.Error("club setup counts", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	return c.JSON(clubSetupFromCounts(cl, counts))
}

// HandleDismissClubSetup hides the Ersteinrichtung card of a club (owner only).
func (h *Handler) HandleDismissClubSetup(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}
	cl, err := h.ownedClub(c, u.ID)
	if err != nil {
		return err
	}
	if err := h.ClubRepo.DismissSetup(ctx, cl.ID); err != nil {
		slog.Error("dismiss club setup", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	audit.LogAudit(c, u.ID, audit.ActionClubSetupDismissed, "club_id", cl.ID)
	return c.SendStatus(fiber.StatusNoContent)
}
