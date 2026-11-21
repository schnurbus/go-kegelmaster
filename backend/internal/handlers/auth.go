package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/audit"
	"github.com/schnurbus/go-kegelmaster/backend/internal/passwordreset"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

const maxPasswordLength = 128

var (
	dummyHashOnce sync.Once
	dummyHash     string
)

// dummyPasswordHash returns a valid bcrypt hash of a random value. Login
// compares against it for unknown e-mails to keep response times uniform.
func (h *Handler) dummyPasswordHash() string {
	dummyHashOnce.Do(func() {
		hash, err := h.AuthSvc.HashPassword(uuid.NewString())
		if err != nil {
			slog.Error("generate dummy password hash", "error", err)
			return
		}
		dummyHash = hash
	})
	return dummyHash
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	// InviteToken is required when REGISTRATION_MODE=invite_only.
	InviteToken string `json:"invite_token"`
}

func (r registerRequest) validate() error {
	email := strings.TrimSpace(strings.ToLower(r.Email))
	password := strings.TrimSpace(r.Password)

	if email == "" || !strings.Contains(email, "@") {
		return errors.New("ungültige E-Mail-Adresse")
	}
	if len(password) < 8 {
		return errors.New("Passwort muss mindestens 8 Zeichen lang sein")
	}
	if len(password) > maxPasswordLength {
		return errors.New("Passwort darf höchstens 128 Zeichen haben")
	}
	return nil
}

func (h *Handler) HandleRegister(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	if err := h.registrationClosedError(); err != nil {
		return err
	}

	var req registerRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if err := req.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)

	ctx, cancel := h.RequestContext()
	defer cancel()

	if err := h.checkRegistrationAllowed(ctx, email, req.InviteToken); err != nil {
		return err
	}

	hash, err := h.AuthSvc.HashPassword(password)
	if err != nil {
		slog.Error("hash password", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	u, err := h.UserRepo.Create(ctx, user.CreateUserParams{
		Email:        email,
		PasswordHash: hash,
	})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrEmailConflict):
			return fiber.NewError(fiber.StatusConflict, "E-Mail ist bereits registriert")
		default:
			slog.Error("create user", "error", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
		}
	}

	// New users stay logged in (remember-me style) after register
	if err := h.IssueAuthCookie(c, u.ID, true); err != nil {
		slog.Error("issue auth cookie", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionUserRegistered, "user_id", u.ID)
	return c.Status(fiber.StatusCreated).JSON(UserResponseFromEntity(u))
}

type loginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

func (h *Handler) HandleLogin(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}

	var req loginRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "E-Mail und Passwort werden benötigt")
	}
	if len(strings.TrimSpace(req.Password)) > maxPasswordLength {
		return fiber.NewError(fiber.StatusBadRequest, "Passwort darf höchstens 128 Zeichen haben")
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserRepo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			// Spend the same bcrypt time as for a real account, so response
			// timing does not reveal which e-mail addresses are registered.
			_ = h.AuthSvc.ComparePassword(h.dummyPasswordHash(), strings.TrimSpace(req.Password))
			return fiber.NewError(fiber.StatusUnauthorized, "Ungültige Anmeldedaten")
		}
		slog.Error("lookup user", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if err := h.AuthSvc.ComparePassword(u.PasswordHash, strings.TrimSpace(req.Password)); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Ungültige Anmeldedaten")
	}
	// Checked only after the password, so the response does not reveal to
	// strangers that the account exists and is disabled.
	if u.Disabled() {
		audit.LogAudit(c, u.ID, audit.ActionUserLoginRejected, "user_id", u.ID, "reason", "disabled")
		return fiber.NewError(fiber.StatusForbidden, errMsgAccountDisabled)
	}

	if err := h.IssueAuthCookie(c, u.ID, req.RememberMe); err != nil {
		slog.Error("issue auth cookie", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	u = h.recordLogin(ctx, u)

	audit.LogAudit(c, u.ID, audit.ActionUserLogin, "user_id", u.ID)
	return c.JSON(UserResponseFromEntity(u))
}

// recordLogin stores the login time for the operator overview. A failure is
// only logged: it must never block a login.
func (h *Handler) recordLogin(ctx context.Context, u user.User) user.User {
	now := time.Now().UTC()
	if err := h.UserRepo.RecordLogin(ctx, u.ID, now); err != nil {
		slog.Warn("record last login", "error", err, "user_id", u.ID)
		return u
	}
	u.LastLoginAt = &now
	return u
}

func (h *Handler) HandleLogout(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}

	h.ClearAuthCookie(c)
	h.ClearCSRFCookie(c)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) HandleCSRFCookie(c fiber.Ctx) error {
	token := uuid.NewString()
	h.SetCSRFCookie(c, token)
	return c.JSON(fiber.Map{"csrf_token": token})
}

func (h *Handler) HandleCurrentUser(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}
	return c.JSON(UserResponseFromEntity(u))
}

type updateCurrentUserRequest struct {
	NotifyGameDaySettled *bool `json:"notify_gameday_settled"`
	// TourSeen / SetupHintSeen: true marks the Rundgang or the hint at the
	// Ersteinrichtung card as seen. There is no way back: a manually started
	// Rundgang ignores the flags.
	TourSeen      bool `json:"tour_seen"`
	SetupHintSeen bool `json:"setup_hint_seen"`
}

// HandleUpdateCurrentUser changes the signed-in user's account settings.
// Omitted fields keep their value.
func (h *Handler) HandleUpdateCurrentUser(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}

	var req updateCurrentUserRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	markTour := req.TourSeen && u.TourSeenAt == nil
	markHint := req.SetupHintSeen && u.SetupHintSeenAt == nil
	if markTour || markHint {
		u, err = h.UserRepo.MarkTourFlags(ctx, u.ID, markTour, markHint)
		if err != nil {
			if errors.Is(err, user.ErrNotFound) {
				return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
			}
			slog.Error("mark tour flags", "error", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
		}
	}

	if req.NotifyGameDaySettled == nil || *req.NotifyGameDaySettled == u.NotifyGameDaySettled {
		return c.JSON(UserResponseFromEntity(u))
	}

	updated, err := h.UserRepo.UpdateNotifyGameDaySettled(ctx, u.ID, *req.NotifyGameDaySettled)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
		}
		slog.Error("update user settings", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionUserSettingsUpdated, "user_id", u.ID, "notify_gameday_settled", updated.NotifyGameDaySettled)
	return c.JSON(UserResponseFromEntity(updated))
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

func (h *Handler) HandleForgotPassword(c fiber.Ctx) error {
	var req forgotPasswordRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || !strings.Contains(email, "@") {
		return fiber.NewError(fiber.StatusBadRequest, "Gültige E-Mail-Adresse ist erforderlich")
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
				"message": "Falls ein Konto mit dieser E-Mail existiert, wurde ein Link zum Zurücksetzen des Passworts gesendet.",
			})
		}
		slog.Error("lookup user for password reset", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if err := h.PasswordResetRepo.DeleteByUserID(ctx, u.ID); err != nil {
		slog.Error("delete old password reset tokens", "error", err)
	}
	token := uuid.NewString()
	expiresAt := time.Now().UTC().Add(time.Duration(h.Config.PasswordResetTokenExpiryMin) * time.Minute)
	if _, err := h.PasswordResetRepo.Create(ctx, u.ID, token, expiresAt); err != nil {
		slog.Error("create password reset token", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	baseURL := strings.TrimSuffix(h.Config.BaseURL, "/")
	resetLink := baseURL + "/reset-password?token=" + token
	if err := h.EmailSvc.SendPasswordResetEmail(u.Email, resetLink); err != nil {
		slog.Error("send password reset email", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "E-Mail konnte nicht gesendet werden. Bitte später erneut versuchen.")
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Falls ein Konto mit dieser E-Mail existiert, wurde ein Link zum Zurücksetzen des Passworts gesendet.",
	})
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (h *Handler) HandleResetPassword(c fiber.Ctx) error {
	var req resetPasswordRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}
	token := strings.TrimSpace(req.Token)
	password := strings.TrimSpace(req.NewPassword)
	if token == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Token ist erforderlich")
	}
	if len(password) < 8 {
		return fiber.NewError(fiber.StatusBadRequest, "Passwort muss mindestens 8 Zeichen lang sein")
	}
	if len(password) > maxPasswordLength {
		return fiber.NewError(fiber.StatusBadRequest, "Passwort darf höchstens 128 Zeichen haben")
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	hash, err := h.AuthSvc.HashPassword(password)
	if err != nil {
		slog.Error("hash password", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Consume deletes the token in the same statement, so a link works only once.
	pr, err := h.PasswordResetRepo.Consume(ctx, token)
	if err != nil {
		if errors.Is(err, passwordreset.ErrNotFound) {
			return fiber.NewError(fiber.StatusBadRequest, "Link ungültig oder abgelaufen. Bitte fordern Sie einen neuen Link an.")
		}
		slog.Error("consume password reset token", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	now := time.Now().UTC()
	if now.After(pr.ExpiresAt) {
		return fiber.NewError(fiber.StatusBadRequest, "Link ungültig oder abgelaufen. Bitte fordern Sie einen neuen Link an.")
	}

	// Recording the change time logs out every existing session of this user.
	if err := h.UserRepo.UpdatePassword(ctx, pr.UserID, hash, now); err != nil {
		slog.Error("update user password", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Tokens created before now are already rejected (see userFromAPIToken);
	// deleting them keeps the token list in the UI truthful.
	if h.APITokenRepo != nil {
		if err := h.APITokenRepo.DeleteByUserID(ctx, pr.UserID); err != nil {
			slog.Error("delete api tokens after password reset", "error", err)
		}
	}
	// A reset is the recovery path after a compromise: passkeys an attacker
	// may have added must not survive it. (A normal password change keeps them.)
	if h.PasskeySvc != nil {
		if err := h.PasskeySvc.DeleteByUserID(ctx, pr.UserID); err != nil {
			slog.Error("delete passkeys after password reset", "error", err)
		}
	}

	audit.LogAudit(c, pr.UserID, audit.ActionUserPasswordReset, "user_id", pr.UserID)
	return c.SendStatus(fiber.StatusNoContent)
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// HandleChangePassword lets a signed-in user set a new password. It requires
// the current password, logs out all other sessions, revokes all API tokens
// and reissues the auth cookie so the current browser stays signed in.
func (h *Handler) HandleChangePassword(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	// A leaked API token must not be enough to take over the account.
	if IsAPITokenRequest(c) {
		return fiber.NewError(fiber.StatusForbidden, "Passwort kann nicht mit einem API-Token geändert werden")
	}

	var req changePasswordRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}
	current := strings.TrimSpace(req.CurrentPassword)
	password := strings.TrimSpace(req.NewPassword)
	if current == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Aktuelles Passwort wird benötigt")
	}
	if len(current) > maxPasswordLength {
		return fiber.NewError(fiber.StatusBadRequest, "Passwort darf höchstens 128 Zeichen haben")
	}
	if len(password) < 8 {
		return fiber.NewError(fiber.StatusBadRequest, "Passwort muss mindestens 8 Zeichen lang sein")
	}
	if len(password) > maxPasswordLength {
		return fiber.NewError(fiber.StatusBadRequest, "Passwort darf höchstens 128 Zeichen haben")
	}
	if password == current {
		return fiber.NewError(fiber.StatusBadRequest, "Das neue Passwort muss sich vom aktuellen unterscheiden")
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	if err := h.AuthSvc.ComparePassword(u.PasswordHash, current); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Aktuelles Passwort ist falsch")
	}

	hash, err := h.AuthSvc.HashPassword(password)
	if err != nil {
		slog.Error("hash password", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Recording the change time logs out every existing session of this user,
	// including this one; the cookie is reissued below.
	now := time.Now().UTC()
	if err := h.UserRepo.UpdatePassword(ctx, u.ID, hash, now); err != nil {
		slog.Error("update user password", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if h.APITokenRepo != nil {
		if err := h.APITokenRepo.DeleteByUserID(ctx, u.ID); err != nil {
			slog.Error("delete api tokens after password change", "error", err)
		}
	}

	remember := false
	if claims, err := h.AuthSvc.ParseToken(strings.TrimSpace(c.Cookies(authCookieName))); err == nil {
		remember = h.AuthSvc.IsRememberMe(claims)
	}
	if err := h.IssueAuthCookie(c, u.ID, remember); err != nil {
		slog.Error("issue auth cookie", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionUserPasswordChanged, "user_id", u.ID)
	u.PasswordHash = hash
	u.PasswordChangedAt = &now
	u.UpdatedAt = now
	return c.JSON(UserResponseFromEntity(u))
}
