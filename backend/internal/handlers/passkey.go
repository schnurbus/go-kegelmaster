package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/audit"
	"github.com/schnurbus/go-kegelmaster/backend/internal/passkey"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

const (
	maxPasskeysPerUser   = 20
	maxPasskeyNameLength = 100
	defaultPasskeyName   = "Passkey"

	// passkeyCeremonyCookie carries the ID of the pending WebAuthn ceremony
	// from "begin" to "finish". The challenge itself stays in the database.
	passkeyCeremonyCookie = "passkey_ceremony"
	passkeyCookiePath     = "/api/auth/passkeys"
)

// PasskeyResponse describes a passkey without key material.
type PasskeyResponse struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Transports []string `json:"transports"`
	// Synced is true for passkeys that the authenticator backs up (e.g. iCloud
	// Keychain, Google Password Manager); false for device-bound security keys.
	Synced     bool       `json:"synced"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

func passkeyResponseFromEntity(c passkey.Credential) PasskeyResponse {
	return PasskeyResponse{
		ID:         c.ID,
		Name:       c.Name,
		Transports: c.Transports,
		Synced:     c.BackupState,
		LastUsedAt: c.LastUsedAt,
		CreatedAt:  c.CreatedAt,
	}
}

type beginPasskeyRegistrationRequest struct {
	CurrentPassword string `json:"current_password"`
}

type finishPasskeyRegistrationRequest struct {
	Name       string          `json:"name"`
	Credential json.RawMessage `json:"credential"`
}

type finishPasskeyLoginRequest struct {
	Credential json.RawMessage `json:"credential"`
	RememberMe bool            `json:"remember_me"`
}

type renamePasskeyRequest struct {
	Name string `json:"name"`
}

func (h *Handler) requirePasskeys() error {
	if h.PasskeySvc == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Passkeys sind nicht verfügbar")
	}
	return nil
}

// sessionUserForPasskeyManagement authenticates passkey management requests.
// Like API tokens, passkeys can only be managed from a browser session: a
// leaked API token must not be enough to add a permanent way into the account.
func (h *Handler) sessionUserForPasskeyManagement(c fiber.Ctx) (user.User, error) {
	if IsAPITokenRequest(c) {
		return user.User{}, fiber.NewError(fiber.StatusForbidden, "Passkeys können nur über die Weboberfläche verwaltet werden")
	}
	if err := h.requirePasskeys(); err != nil {
		return user.User{}, err
	}
	ctx, cancel := h.RequestContext()
	defer cancel()
	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return user.User{}, fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}
	return u, nil
}

func (h *Handler) setPasskeyCeremonyCookie(c fiber.Ctx, id string) {
	c.Cookie(&fiber.Cookie{
		Name:     passkeyCeremonyCookie,
		Value:    id,
		HTTPOnly: true,
		Secure:   h.Config.AppEnv == "production",
		SameSite: fiber.CookieSameSiteStrictMode,
		Path:     passkeyCookiePath,
		MaxAge:   int(passkey.CeremonyTimeout.Seconds()),
	})
}

// takePasskeyCeremonyCookie returns the pending ceremony ID and clears the
// cookie; the ceremony is single-use either way.
func (h *Handler) takePasskeyCeremonyCookie(c fiber.Ctx) string {
	id := strings.TrimSpace(c.Cookies(passkeyCeremonyCookie))
	c.Cookie(&fiber.Cookie{
		Name:     passkeyCeremonyCookie,
		Value:    "",
		HTTPOnly: true,
		Secure:   h.Config.AppEnv == "production",
		SameSite: fiber.CookieSameSiteStrictMode,
		Path:     passkeyCookiePath,
		Expires:  time.Unix(0, 0),
	})
	return id
}

func validatePasskeyName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		name = defaultPasskeyName
	}
	if utf8.RuneCountInString(name) > maxPasskeyNameLength {
		return "", fiber.NewError(fiber.StatusBadRequest, "Name darf höchstens 100 Zeichen haben")
	}
	return name, nil
}

// HandleBeginPasskeyRegistration starts adding a passkey. The current password
// is required, so a hijacked session alone cannot plant a passkey.
func (h *Handler) HandleBeginPasskeyRegistration(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	u, err := h.sessionUserForPasskeyManagement(c)
	if err != nil {
		return err
	}

	var req beginPasskeyRegistrationRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}
	current := strings.TrimSpace(req.CurrentPassword)
	if current == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Aktuelles Passwort wird benötigt")
	}
	if len(current) > maxPasswordLength {
		return fiber.NewError(fiber.StatusBadRequest, "Passwort darf höchstens 128 Zeichen haben")
	}
	if err := h.AuthSvc.ComparePassword(u.PasswordHash, current); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Aktuelles Passwort ist falsch")
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	count, err := h.PasskeySvc.CountByUserID(ctx, u.ID)
	if err != nil {
		slog.Error("count passkeys", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if count >= maxPasskeysPerUser {
		return fiber.NewError(fiber.StatusConflict, "Maximal 20 Passkeys pro Benutzer möglich")
	}

	options, ceremonyID, err := h.PasskeySvc.BeginRegistration(ctx, u.ID, u.Email)
	if err != nil {
		slog.Error("begin passkey registration", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	h.setPasskeyCeremonyCookie(c, ceremonyID)
	return c.JSON(options)
}

// HandleFinishPasskeyRegistration verifies the authenticator response and stores the passkey.
func (h *Handler) HandleFinishPasskeyRegistration(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	u, err := h.sessionUserForPasskeyManagement(c)
	if err != nil {
		return err
	}
	ceremonyID := h.takePasskeyCeremonyCookie(c)

	var req finishPasskeyRegistrationRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil || len(req.Credential) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}
	name, err := validatePasskeyName(req.Name)
	if err != nil {
		return err
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	cred, err := h.PasskeySvc.FinishRegistration(ctx, ceremonyID, u.ID, u.Email, name, req.Credential)
	if err != nil {
		switch {
		case errors.Is(err, passkey.ErrSessionInvalid):
			return fiber.NewError(fiber.StatusBadRequest, "Vorgang abgelaufen. Bitte erneut versuchen.")
		case errors.Is(err, passkey.ErrVerification):
			slog.Info("passkey registration rejected", "error", err, "user_id", u.ID)
			return fiber.NewError(fiber.StatusBadRequest, "Passkey konnte nicht überprüft werden")
		case errors.Is(err, passkey.ErrAlreadyRegistered):
			return fiber.NewError(fiber.StatusConflict, "Dieser Passkey ist bereits registriert")
		default:
			slog.Error("finish passkey registration", "error", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
		}
	}

	audit.LogAudit(c, u.ID, audit.ActionPasskeyCreated, "passkey_id", cred.ID)
	return c.Status(fiber.StatusCreated).JSON(passkeyResponseFromEntity(cred))
}

// HandleBeginPasskeyLogin starts a sign-in with any passkey of this site.
func (h *Handler) HandleBeginPasskeyLogin(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	if err := h.requirePasskeys(); err != nil {
		return err
	}
	ctx, cancel := h.RequestContext()
	defer cancel()

	options, ceremonyID, err := h.PasskeySvc.BeginLogin(ctx)
	if err != nil {
		slog.Error("begin passkey login", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	h.setPasskeyCeremonyCookie(c, ceremonyID)
	return c.JSON(options)
}

// HandleFinishPasskeyLogin verifies the assertion and signs the user in.
func (h *Handler) HandleFinishPasskeyLogin(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	if err := h.requirePasskeys(); err != nil {
		return err
	}
	ceremonyID := h.takePasskeyCeremonyCookie(c)

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
			return fiber.NewError(fiber.StatusBadRequest, "Anmeldung abgelaufen. Bitte erneut versuchen.")
		case errors.Is(err, passkey.ErrVerification):
			slog.Info("passkey login rejected", "error", err)
			return fiber.NewError(fiber.StatusUnauthorized, "Anmeldung mit Passkey fehlgeschlagen")
		default:
			slog.Error("finish passkey login", "error", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
		}
	}

	u, err := h.UserRepo.GetByID(ctx, userID)
	if err != nil {
		slog.Error("lookup passkey user", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if err := h.IssueAuthCookie(c, u.ID, req.RememberMe); err != nil {
		slog.Error("issue auth cookie", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionUserLoginPasskey, "user_id", u.ID, "passkey_id", cred.ID)
	return c.JSON(UserResponseFromEntity(u))
}

func (h *Handler) HandleListPasskeys(c fiber.Ctx) error {
	u, err := h.sessionUserForPasskeyManagement(c)
	if err != nil {
		return err
	}
	ctx, cancel := h.RequestContext()
	defer cancel()

	creds, err := h.PasskeySvc.ListByUserID(ctx, u.ID)
	if err != nil {
		slog.Error("list passkeys", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	out := make([]PasskeyResponse, 0, len(creds))
	for _, cred := range creds {
		out = append(out, passkeyResponseFromEntity(cred))
	}
	return c.JSON(out)
}

func (h *Handler) HandleRenamePasskey(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	u, err := h.sessionUserForPasskeyManagement(c)
	if err != nil {
		return err
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Passkey-ID")
	}
	var req renamePasskeyRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}
	if strings.TrimSpace(req.Name) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Name ist erforderlich")
	}
	name, err := validatePasskeyName(req.Name)
	if err != nil {
		return err
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	cred, err := h.PasskeySvc.Rename(ctx, id, u.ID, name)
	if err != nil {
		if errors.Is(err, passkey.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Passkey nicht gefunden")
		}
		slog.Error("rename passkey", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionPasskeyRenamed, "passkey_id", id)
	return c.JSON(passkeyResponseFromEntity(cred))
}

func (h *Handler) HandleDeletePasskey(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	u, err := h.sessionUserForPasskeyManagement(c)
	if err != nil {
		return err
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Passkey-ID")
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	if err := h.PasskeySvc.Delete(ctx, id, u.ID); err != nil {
		if errors.Is(err, passkey.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Passkey nicht gefunden")
		}
		slog.Error("delete passkey", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionPasskeyDeleted, "passkey_id", id)
	return c.SendStatus(fiber.StatusNoContent)
}
