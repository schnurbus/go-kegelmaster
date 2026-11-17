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

	"github.com/schnurbus/go-kegelmaster/backend/internal/apitoken"
	"github.com/schnurbus/go-kegelmaster/backend/internal/audit"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

const (
	maxAPITokensPerUser   = 20
	maxAPITokenNameLength = 100
	maxAPITokenExpiryDays = 365
)

// APITokenResponse describes an API token without its secret.
type APITokenResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	TokenPrefix string     `json:"token_prefix"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreatedAPITokenResponse is returned once on creation and includes the secret.
type CreatedAPITokenResponse struct {
	APITokenResponse
	Token string `json:"token"`
}

func apiTokenResponseFromEntity(t apitoken.Token) APITokenResponse {
	return APITokenResponse{
		ID:          t.ID,
		Name:        t.Name,
		TokenPrefix: t.Prefix,
		ExpiresAt:   t.ExpiresAt,
		LastUsedAt:  t.LastUsedAt,
		CreatedAt:   t.CreatedAt,
	}
}

type createAPITokenRequest struct {
	Name          string `json:"name"`
	ExpiresInDays *int   `json:"expires_in_days"`
}

// sessionUserForTokenManagement authenticates token management requests.
// Tokens can only be managed from a browser session, so a leaked token cannot
// be used to mint further tokens or to hide itself by deleting others.
func (h *Handler) sessionUserForTokenManagement(c fiber.Ctx) (user.User, error) {
	if IsAPITokenRequest(c) {
		return user.User{}, fiber.NewError(fiber.StatusForbidden, "API-Tokens können nur über die Weboberfläche verwaltet werden")
	}
	if h.APITokenRepo == nil {
		return user.User{}, fiber.NewError(fiber.StatusServiceUnavailable, "API-Tokens sind nicht verfügbar")
	}
	ctx, cancel := h.RequestContext()
	defer cancel()
	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return user.User{}, fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}
	return u, nil
}

func (h *Handler) HandleListAPITokens(c fiber.Ctx) error {
	u, err := h.sessionUserForTokenManagement(c)
	if err != nil {
		return err
	}
	ctx, cancel := h.RequestContext()
	defer cancel()

	tokens, err := h.APITokenRepo.ListByUserID(ctx, u.ID)
	if err != nil {
		slog.Error("list api tokens", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	out := make([]APITokenResponse, 0, len(tokens))
	for _, t := range tokens {
		out = append(out, apiTokenResponseFromEntity(t))
	}
	return c.JSON(out)
}

func (h *Handler) HandleCreateAPIToken(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	u, err := h.sessionUserForTokenManagement(c)
	if err != nil {
		return err
	}

	var req createAPITokenRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Name ist erforderlich")
	}
	if utf8.RuneCountInString(name) > maxAPITokenNameLength {
		return fiber.NewError(fiber.StatusBadRequest, "Name darf höchstens 100 Zeichen haben")
	}
	var expiresAt *time.Time
	if req.ExpiresInDays != nil {
		days := *req.ExpiresInDays
		if days < 1 || days > maxAPITokenExpiryDays {
			return fiber.NewError(fiber.StatusBadRequest, "Gültigkeit muss zwischen 1 und 365 Tagen liegen")
		}
		t := time.Now().UTC().AddDate(0, 0, days)
		expiresAt = &t
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	count, err := h.APITokenRepo.CountByUserID(ctx, u.ID)
	if err != nil {
		slog.Error("count api tokens", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if count >= maxAPITokensPerUser {
		return fiber.NewError(fiber.StatusConflict, "Maximal 20 API-Tokens pro Benutzer möglich")
	}

	t, plain, err := h.APITokenRepo.Create(ctx, u.ID, name, expiresAt)
	if err != nil {
		slog.Error("create api token", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionAPITokenCreated, "token_id", t.ID, "token_prefix", t.Prefix)
	return c.Status(fiber.StatusCreated).JSON(CreatedAPITokenResponse{
		APITokenResponse: apiTokenResponseFromEntity(t),
		Token:            plain,
	})
}

func (h *Handler) HandleDeleteAPIToken(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	u, err := h.sessionUserForTokenManagement(c)
	if err != nil {
		return err
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Token-ID")
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	if err := h.APITokenRepo.Delete(ctx, id, u.ID); err != nil {
		if errors.Is(err, apitoken.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "API-Token nicht gefunden")
		}
		slog.Error("delete api token", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAudit(c, u.ID, audit.ActionAPITokenDeleted, "token_id", id)
	return c.SendStatus(fiber.StatusNoContent)
}
