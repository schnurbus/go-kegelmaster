package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/audit"
	"github.com/schnurbus/go-kegelmaster/backend/internal/passkey"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

// ErrorCodeReauthRequired is the "error" value of the 401 response that asks
// an instance admin to confirm their login again (KEG-9).
const ErrorCodeReauthRequired = "reauth_required"

// authClockSkew tolerates an auth_time slightly in the future (several
// backend replicas with minimally different clocks).
const authClockSkew = time.Minute

// recentlyAuthenticated reports whether the session cookie carries an
// auth_time within the configured ADMIN_REAUTH_MINUTES window. Cookies
// without the claim (issued before it existed) count as stale.
func (h *Handler) recentlyAuthenticated(c fiber.Ctx, now time.Time) bool {
	claims, err := h.AuthSvc.ParseToken(strings.TrimSpace(c.Cookies(authCookieName)))
	if err != nil {
		return false
	}
	authTime, ok := claims.AuthenticatedAt()
	if !ok {
		return false
	}
	age := now.Sub(authTime)
	return age >= -authClockSkew && age <= h.Config.AdminReauthWindow()
}

// respondAdminReauthRequired is used by RequireInstanceAdmin once the request
// is known to come from an instance admin's cookie session whose last real
// authentication (password or passkey login, or a re-authentication) is older
// than ADMIN_REAUTH_MINUTES. It answers 401 {"error":"reauth_required"}; the
// frontend then asks for the password or a passkey. This applies whether or
// not the account has a passkey. Non-admins never get here, so they keep
// getting 404 and cannot learn that the area exists.
func (h *Handler) respondAdminReauthRequired(c fiber.Ctx, adminID string) error {
	audit.LogAdminAudit(c, adminID, audit.ActionAdminReauthRequired, "path", c.Path())
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error":   ErrorCodeReauthRequired,
		"message": "Erneute Anmeldung erforderlich, um den Admin-Bereich zu nutzen.",
	})
}

// reauthSessionUser resolves the user of the current browser session for a
// re-authentication. API tokens are rejected: re-authentication refreshes a
// cookie session and must never be reachable with a bearer token.
func (h *Handler) reauthSessionUser(c fiber.Ctx) (user.User, error) {
	if IsAPITokenRequest(c) {
		return user.User{}, fiber.NewError(fiber.StatusForbidden, "Nur über die Weboberfläche möglich")
	}
	ctx, cancel := h.RequestContext()
	defer cancel()
	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return user.User{}, fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}
	return u, nil
}

// reissueAfterReauth issues a new session cookie with a fresh auth_time for
// the current user and keeps the remember-me choice of the existing session.
func (h *Handler) reissueAfterReauth(c fiber.Ctx, userID string) error {
	remember := false
	if claims, err := h.AuthSvc.ParseToken(strings.TrimSpace(c.Cookies(authCookieName))); err == nil {
		remember = h.AuthSvc.IsRememberMe(claims)
	}
	if err := h.IssueAuthCookie(c, userID, remember); err != nil {
		slog.Error("issue auth cookie", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	return nil
}

type reauthRequest struct {
	Password string `json:"password"`
}

// HandleReauth confirms the logged-in user's identity with their password and
// reissues the session cookie with a fresh auth_time (and a new CSRF cookie).
func (h *Handler) HandleReauth(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	u, err := h.reauthSessionUser(c)
	if err != nil {
		return err
	}

	var req reauthRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}
	password := strings.TrimSpace(req.Password)
	if password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Passwort wird benötigt")
	}
	if len(password) > maxPasswordLength {
		return fiber.NewError(fiber.StatusBadRequest, "Passwort darf höchstens 128 Zeichen haben")
	}
	if err := h.AuthSvc.ComparePassword(u.PasswordHash, password); err != nil {
		audit.LogAudit(c, u.ID, audit.ActionUserReauthFailed, "user_id", u.ID, "method", "password")
		return fiber.NewError(fiber.StatusUnauthorized, "Passwort ist falsch")
	}

	if err := h.reissueAfterReauth(c, u.ID); err != nil {
		return err
	}
	audit.LogAudit(c, u.ID, audit.ActionUserReauth, "user_id", u.ID, "method", "password")
	return c.JSON(UserResponseFromEntity(u))
}

// HandlePasskeyReauth finishes a passkey ceremony started with
// POST /api/auth/passkeys/login/begin as a re-authentication: the passkey must
// belong to the logged-in user. On success the session cookie is reissued with
// a fresh auth_time, exactly like HandleReauth.
func (h *Handler) HandlePasskeyReauth(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	if err := h.requirePasskeys(); err != nil {
		return err
	}
	// Consume the ceremony in any case, so it cannot be replayed.
	ceremonyID := h.takePasskeyCeremonyCookie(c)
	u, err := h.reauthSessionUser(c)
	if err != nil {
		return err
	}

	var req finishPasskeyLoginRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil || len(req.Credential) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	userID, cred, err := h.PasskeySvc.FinishLogin(ctx, ceremonyID, req.Credential)
	if err != nil {
		switch {
		case errors.Is(err, passkey.ErrSessionInvalid):
			return fiber.NewError(fiber.StatusBadRequest, "Vorgang abgelaufen. Bitte erneut versuchen.")
		case errors.Is(err, passkey.ErrVerification):
			slog.Info("passkey reauth rejected", "error", err, "user_id", u.ID)
			audit.LogAudit(c, u.ID, audit.ActionUserReauthFailed, "user_id", u.ID, "method", "passkey")
			return fiber.NewError(fiber.StatusUnauthorized, "Bestätigung mit Passkey fehlgeschlagen")
		default:
			slog.Error("finish passkey reauth", "error", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
		}
	}
	if userID != u.ID {
		audit.LogAudit(c, u.ID, audit.ActionUserReauthFailed, "user_id", u.ID, "method", "passkey",
			"reason", "foreign_passkey", "passkey_id", cred.ID)
		return fiber.NewError(fiber.StatusUnauthorized, "Dieser Passkey gehört nicht zum angemeldeten Konto")
	}

	if err := h.reissueAfterReauth(c, u.ID); err != nil {
		return err
	}
	audit.LogAudit(c, u.ID, audit.ActionUserReauth, "user_id", u.ID, "method", "passkey", "passkey_id", cred.ID)
	return c.JSON(UserResponseFromEntity(u))
}
