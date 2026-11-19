package handlers

import (
	"bytes"
	"encoding/csv"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/admin"
	"github.com/schnurbus/go-kegelmaster/backend/internal/audit"
)

// LocalsAdminUserID holds the instance admin's user ID once RequireInstanceAdmin passed.
const LocalsAdminUserID = "admin_user_id"

// RequireInstanceAdmin guards /api/admin. Only cookie sessions of users with
// the instance admin flag pass. Everything else gets 404 (not 401/403), so the
// area cannot be discovered:
//   - requests with an "Authorization: Bearer" header (API tokens, and thereby
//     MCP, which calls the API in-process with the caller's token), even if a
//     valid admin cookie is sent along,
//   - unauthenticated requests and invalid/revoked sessions,
//   - users without the instance admin flag.
//
// It is deliberately independent of the club PermissionChecker.
func (h *Handler) RequireInstanceAdmin(c fiber.Ctx) error {
	if h.AdminRepo == nil || IsAPITokenRequest(c) {
		return fiber.ErrNotFound
	}
	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil || !u.IsInstanceAdmin {
		return fiber.ErrNotFound
	}
	c.Locals(LocalsAdminUserID, u.ID)
	// Operator data must not end up in shared or browser caches.
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.Next()
}

func adminUserID(c fiber.Ctx) string {
	id, _ := c.Locals(LocalsAdminUserID).(string)
	return id
}

func adminListParams(c fiber.Ctx) admin.ListParams {
	return admin.ListParams{
		Search: c.Query("search"),
		Filter: c.Query("filter"),
		Sort:   c.Query("sort"),
		Order:  c.Query("order"),
		Page:   c.Query("page"),
		Limit:  c.Query("limit"),
	}
}

func adminValidationError(err error) error {
	var ve admin.ValidationError
	if errors.As(err, &ve) {
		return fiber.NewError(fiber.StatusBadRequest, ve.Message)
	}
	return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
}

// HandleAdminStats returns the dashboard key figures and anomalies.
func (h *Handler) HandleAdminStats(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	stats, err := h.AdminRepo.Stats(ctx)
	if err != nil {
		slog.Error("admin stats", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	audit.LogAdminAudit(c, adminUserID(c), audit.ActionAdminStatsViewed)
	return c.JSON(stats)
}

// HandleAdminListUsers lists accounts with search, filter, sorting and paging.
func (h *Handler) HandleAdminListUsers(c fiber.Ctx) error {
	q, err := admin.ParseUserListQuery(adminListParams(c))
	if err != nil {
		return adminValidationError(err)
	}
	ctx, cancel := h.RequestContext()
	defer cancel()

	page, err := h.AdminRepo.ListUsers(ctx, q)
	if err != nil {
		slog.Error("admin list users", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	audit.LogAdminAudit(c, adminUserID(c), audit.ActionAdminUsersListed,
		"search", q.Search, "filter", q.Filter, "page", q.Page)
	return c.JSON(page)
}

// HandleAdminGetUser returns one account with clubs, player links, API tokens
// (without hash) and passkeys.
func (h *Handler) HandleAdminGetUser(c fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Benutzer nicht gefunden")
	}
	ctx, cancel := h.RequestContext()
	defer cancel()

	detail, err := h.AdminRepo.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, admin.ErrUserNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Benutzer nicht gefunden")
		}
		slog.Error("admin get user", "error", err, "target_user_id", id)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	audit.LogAdminAudit(c, adminUserID(c), audit.ActionAdminUserViewed, "target_user_id", id)
	return c.JSON(detail)
}

// HandleAdminListClubs lists clubs with owner, player counts and last activity.
func (h *Handler) HandleAdminListClubs(c fiber.Ctx) error {
	q, err := admin.ParseClubListQuery(adminListParams(c))
	if err != nil {
		return adminValidationError(err)
	}
	ctx, cancel := h.RequestContext()
	defer cancel()

	page, err := h.AdminRepo.ListClubs(ctx, q)
	if err != nil {
		slog.Error("admin list clubs", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	audit.LogAdminAudit(c, adminUserID(c), audit.ActionAdminClubsListed,
		"search", q.Search, "filter", q.Filter, "page", q.Page)
	return c.JSON(page)
}

// HandleAdminGetClub returns a club with owner, players, open invitations and
// roles, but no business data.
func (h *Handler) HandleAdminGetClub(c fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Klub nicht gefunden")
	}
	ctx, cancel := h.RequestContext()
	defer cancel()

	detail, err := h.AdminRepo.GetClub(ctx, id)
	if err != nil {
		if errors.Is(err, admin.ErrClubNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Klub nicht gefunden")
		}
		slog.Error("admin get club", "error", err, "club_id", id)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	audit.LogAdminAudit(c, adminUserID(c), audit.ActionAdminClubViewed, "club_id", id)
	return c.JSON(detail)
}

// HandleAdminExportMemberships returns all user↔club links as CSV
// (user_email, club_name, is_owner, player_name, role_name).
func (h *Handler) HandleAdminExportMemberships(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	rows, err := h.AdminRepo.Memberships(ctx)
	if err != nil {
		slog.Error("admin export memberships", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"user_email", "club_name", "is_owner", "player_name", "role_name"})
	for _, m := range rows {
		_ = w.Write([]string{
			csvSafe(m.UserEmail),
			csvSafe(m.ClubName),
			strconv.FormatBool(m.IsOwner),
			csvSafe(m.PlayerName),
			csvSafe(m.RoleName),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		slog.Error("admin export memberships: write csv", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	audit.LogAdminAudit(c, adminUserID(c), audit.ActionAdminExport, "export", "memberships", "rows", len(rows))
	filename := "memberships-" + time.Now().UTC().Format("2006-01-02") + ".csv"
	c.Set(fiber.HeaderContentType, "text/csv; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+filename+`"`)
	return c.Send(buf.Bytes())
}

// csvSafe neutralizes values a spreadsheet would interpret as a formula
// (CSV injection): club and player names are chosen by users.
func csvSafe(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}
