package handlers

import (
	"context"
	"crypto/subtle"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/apitoken"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

const (
	authCookieName = "auth_token"
	csrfCookieName = "csrf_token"

	// LocalsUserID holds the authenticated user's ID once a handler resolved it
	// (used by the request log for API-token requests, which carry no JWT).
	LocalsUserID = "auth_user_id"
	// LocalsAPITokenID is set when the request was authenticated with an API token.
	LocalsAPITokenID = "auth_api_token_id"
)

// bearerToken returns the token from an "Authorization: Bearer <token>" header.
func bearerToken(c fiber.Ctx) (string, bool) {
	header := strings.TrimSpace(c.Get(fiber.HeaderAuthorization))
	if len(header) < 7 || !strings.EqualFold(header[:7], "bearer ") {
		return "", false
	}
	return strings.TrimSpace(header[7:]), true
}

// IsAPITokenRequest reports whether the request authenticates with a bearer token.
func IsAPITokenRequest(c fiber.Ctx) bool {
	_, ok := bearerToken(c)
	return ok
}

// RequestContext creates a context with timeout for request handling.
func (h *Handler) RequestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// ValidateCSRF validates the CSRF token from header and cookie.
//
// Requests authenticated with an API token are exempt: CSRF exploits cookies a
// browser attaches automatically, while a bearer header has to be set
// explicitly, and cross-origin pages cannot send it (CORS does not allow the
// Authorization header). UserFromCookie never falls back to the cookie when a
// bearer header is present, so the exemption cannot be abused with a bogus token.
func (h *Handler) ValidateCSRF(c fiber.Ctx) error {
	if IsAPITokenRequest(c) {
		return nil
	}
	header := strings.TrimSpace(c.Get("X-CSRF-Token"))
	cookie := strings.TrimSpace(c.Cookies(csrfCookieName))
	if header == "" || cookie == "" || subtle.ConstantTimeCompare([]byte(header), []byte(cookie)) != 1 {
		return fiber.NewError(fiber.StatusForbidden, "CSRF-Token ungültig")
	}
	return nil
}

var (
	errAPITokensDisabled = errors.New("api tokens not configured")
	errAPITokenInvalid   = errors.New("api token invalid")
)

// UserFromCookie resolves the authenticated user: from the API token when an
// "Authorization: Bearer" header is present, otherwise from the auth cookie.
func (h *Handler) UserFromCookie(ctx context.Context, c fiber.Ctx) (user.User, error) {
	if plain, ok := bearerToken(c); ok {
		return h.userFromAPIToken(ctx, c, plain)
	}

	token := strings.TrimSpace(c.Cookies(authCookieName))
	if token == "" {
		return user.User{}, errors.New("missing auth cookie")
	}

	claims, err := h.AuthSvc.ParseToken(token)
	if err != nil {
		return user.User{}, err
	}

	u, err := h.UserRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return user.User{}, err
	}
	if u.Disabled() {
		return user.User{}, errAccountDisabled
	}
	// Tokens issued before the last password reset or the last session
	// revocation by an instance admin are no longer accepted.
	if claims.IssuedAt == nil || !u.SessionValid(claims.IssuedAt.Time) {
		return user.User{}, errSessionRevoked
	}
	c.Locals(LocalsUserID, u.ID)
	return u, nil
}

func (h *Handler) userFromAPIToken(ctx context.Context, c fiber.Ctx, plain string) (user.User, error) {
	if h.APITokenRepo == nil {
		return user.User{}, errAPITokensDisabled
	}
	now := time.Now().UTC()
	t, err := h.APITokenRepo.Authenticate(ctx, plain, now)
	if err != nil {
		if errors.Is(err, apitoken.ErrNotFound) || errors.Is(err, apitoken.ErrExpired) {
			return user.User{}, errAPITokenInvalid
		}
		return user.User{}, err
	}
	u, err := h.UserRepo.GetByID(ctx, t.UserID)
	if err != nil {
		return user.User{}, err
	}
	if u.Disabled() {
		return user.User{}, errAccountDisabled
	}
	// Like sessions, tokens created before the last password reset are revoked.
	// (An admin session revocation does not affect API tokens; the admin
	// revokes them individually.)
	// Unlike JWTs, token timestamps are exact, so no second truncation here.
	if u.PasswordChangedAt != nil && t.CreatedAt.Before(*u.PasswordChangedAt) {
		return user.User{}, errSessionRevoked
	}
	if err := h.APITokenRepo.Touch(ctx, t.ID, now); err != nil {
		slog.Warn("record api token use", "error", err, "token_id", t.ID)
	}
	c.Locals(LocalsUserID, u.ID)
	c.Locals(LocalsAPITokenID, t.ID)
	return u, nil
}

// errMsgAccountDisabled is shown when a disabled account tries to log in.
const errMsgAccountDisabled = "Dieses Konto ist gesperrt. Bitte wende dich an den Betreiber."

var (
	errSessionRevoked  = errors.New("session issued before last password change or session revocation")
	errAccountDisabled = errors.New("account disabled by instance admin")
)

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
