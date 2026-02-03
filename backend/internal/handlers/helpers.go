package handlers

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

const (
	authCookieName = "auth_token"
	csrfCookieName = "csrf_token"
)

// RequestContext creates a context with timeout for request handling.
func (h *Handler) RequestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// ValidateCSRF validates the CSRF token from header and cookie.
func (h *Handler) ValidateCSRF(c fiber.Ctx) error {
	header := strings.TrimSpace(c.Get("X-CSRF-Token"))
	cookie := strings.TrimSpace(c.Cookies(csrfCookieName))
	if header == "" || cookie == "" || header != cookie {
		return fiber.NewError(fiber.StatusForbidden, "CSRF-Token ungültig")
	}
	return nil
}

// UserFromCookie extracts and validates the user from the auth cookie.
func (h *Handler) UserFromCookie(ctx context.Context, c fiber.Ctx) (user.User, error) {
	token := strings.TrimSpace(c.Cookies(authCookieName))
	if token == "" {
		return user.User{}, errors.New("missing auth cookie")
	}

	claims, err := h.AuthSvc.ParseToken(token)
	if err != nil {
		return user.User{}, err
	}

	return h.UserRepo.GetByID(ctx, claims.UserID)
}

// IssueAuthCookie sets the authentication cookie for the user.
// When rememberMe is false, the cookie has no Expires (session cookie); when true, it lasts RememberMeDays.
func (h *Handler) IssueAuthCookie(c fiber.Ctx, userID string, rememberMe bool) error {
	token, err := h.AuthSvc.GenerateToken(userID, rememberMe)
	if err != nil {
		return err
	}

	sameSite := fiber.CookieSameSiteLaxMode
	if h.Config.AppEnv == "production" {
		sameSite = fiber.CookieSameSiteStrictMode
	}

	cookie := &fiber.Cookie{
		Name:     authCookieName,
		Value:    token,
		HTTPOnly: true,
		Secure:   h.Config.AppEnv == "production",
		SameSite: sameSite,
		Path:     "/",
	}
	if rememberMe {
		cookie.Expires = time.Now().Add(h.AuthSvc.TokenTTL(true))
	}
	// when !rememberMe, Expires is zero → session cookie (browser discards on close)
	c.Cookie(cookie)

	h.SetCSRFCookie(c, uuid.NewString())
	return nil
}

// ClearAuthCookie removes the authentication cookie.
func (h *Handler) ClearAuthCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     authCookieName,
		Value:    "",
		HTTPOnly: true,
		Expires:  time.Unix(0, 0),
		Path:     "/",
	})
}

// SetCSRFCookie sets the CSRF token cookie.
func (h *Handler) SetCSRFCookie(c fiber.Ctx, token string) {
	sameSite := fiber.CookieSameSiteLaxMode
	if h.Config.AppEnv == "production" {
		sameSite = fiber.CookieSameSiteStrictMode
	}

	c.Cookie(&fiber.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		HTTPOnly: false,
		Secure:   h.Config.AppEnv == "production",
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: sameSite,
		Path:     "/",
	})
}

// ClearCSRFCookie removes the CSRF token cookie.
func (h *Handler) ClearCSRFCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     csrfCookieName,
		Value:    "",
		HTTPOnly: false,
		Expires:  time.Unix(0, 0),
		Path:     "/",
	})
}

