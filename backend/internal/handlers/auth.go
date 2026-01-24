package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

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

	if err := h.IssueAuthCookie(c, u.ID); err != nil {
		slog.Error("issue auth cookie", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.Status(fiber.StatusCreated).JSON(UserResponseFromEntity(u))
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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

	if err := h.IssueAuthCookie(c, u.ID); err != nil {
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

