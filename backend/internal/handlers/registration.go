package handlers

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/config"
	"github.com/schnurbus/go-kegelmaster/backend/internal/invitation"
)

// registrationClosedError returns a 403 when REGISTRATION_MODE=closed, so
// the request is rejected before its body is even looked at.
func (h *Handler) registrationClosedError() error {
	if h.Config.EffectiveRegistrationMode() == config.RegistrationModeClosed {
		return fiber.NewError(fiber.StatusForbidden, "Die Registrierung ist auf dieser Instanz deaktiviert")
	}
	return nil
}

// checkRegistrationAllowed enforces REGISTRATION_MODE for POST /api/auth/register.
// In "invite_only" mode the request needs a valid, unexpired and not yet
// accepted invitation token whose e-mail matches the one being registered.
// The invitation itself is accepted afterwards via the usual
// POST /api/invitations/:token/accept.
func (h *Handler) checkRegistrationAllowed(ctx context.Context, email, inviteToken string) error {
	switch h.Config.EffectiveRegistrationMode() {
	case config.RegistrationModeOpen:
		return nil
	case config.RegistrationModeInviteOnly:
		// handled below
	default: // closed (invalid values are rejected at startup)
		return fiber.NewError(fiber.StatusForbidden, "Die Registrierung ist auf dieser Instanz deaktiviert")
	}

	inviteToken = strings.TrimSpace(inviteToken)
	if inviteToken == "" || h.InvitationRepo == nil {
		return fiber.NewError(fiber.StatusForbidden, "Registrierung nur mit Einladung möglich")
	}

	inv, err := h.InvitationRepo.GetByToken(ctx, inviteToken)
	if err != nil {
		switch {
		case errors.Is(err, invitation.ErrNotFound):
			return fiber.NewError(fiber.StatusForbidden, "Registrierung nur mit gültiger Einladung möglich")
		case errors.Is(err, invitation.ErrExpired):
			return fiber.NewError(fiber.StatusForbidden, "Die Einladung ist abgelaufen")
		case errors.Is(err, invitation.ErrAlreadyAccepted):
			return fiber.NewError(fiber.StatusForbidden, "Die Einladung wurde bereits angenommen")
		}
		slog.Error("get invitation for registration", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if !strings.EqualFold(strings.TrimSpace(inv.Email), email) {
		return fiber.NewError(fiber.StatusForbidden, "Die E-Mail-Adresse passt nicht zur Einladung")
	}
	return nil
}
