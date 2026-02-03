package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/passwordreset"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

const maxPasswordLength = 128

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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

	var req registerRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if err := req.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)

	hash, err := h.AuthSvc.HashPassword(password)
	if err != nil {
		slog.Error("hash password", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

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
			return fiber.NewError(fiber.StatusUnauthorized, "Ungültige Anmeldedaten")
		}
		slog.Error("lookup user", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if err := h.AuthSvc.ComparePassword(u.PasswordHash, strings.TrimSpace(req.Password)); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Ungültige Anmeldedaten")
	}

	if err := h.IssueAuthCookie(c, u.ID, req.RememberMe); err != nil {
		slog.Error("issue auth cookie", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(UserResponseFromEntity(u))
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

	pr, err := h.PasswordResetRepo.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, passwordreset.ErrNotFound) {
			return fiber.NewError(fiber.StatusBadRequest, "Link ungültig oder abgelaufen. Bitte fordern Sie einen neuen Link an.")
		}
		slog.Error("get password reset token", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if time.Now().UTC().After(pr.ExpiresAt) {
		_ = h.PasswordResetRepo.DeleteByToken(ctx, token)
		return fiber.NewError(fiber.StatusBadRequest, "Link ungültig oder abgelaufen. Bitte fordern Sie einen neuen Link an.")
	}

	hash, err := h.AuthSvc.HashPassword(password)
	if err != nil {
		slog.Error("hash password", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if err := h.UserRepo.UpdatePassword(ctx, pr.UserID, hash); err != nil {
		slog.Error("update user password", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if err := h.PasswordResetRepo.DeleteByToken(ctx, token); err != nil {
		slog.Error("delete password reset token", "error", err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

