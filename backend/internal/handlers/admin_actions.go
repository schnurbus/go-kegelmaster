package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/schnurbus/go-kegelmaster/backend/internal/admin"
	"github.com/schnurbus/go-kegelmaster/backend/internal/apitoken"
	"github.com/schnurbus/go-kegelmaster/backend/internal/audit"
	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

// Instance admin interventions (operator overview, phase 3). All of them run
// behind RequireInstanceAdmin (cookie session of an instance admin, 404
// otherwise), validate the CSRF token and are audited with admin=true.
//
// Guards beyond the admin check:
//   - an admin cannot disable or delete the own account,
//   - other instance admins cannot be disabled or deleted through the web UI:
//     like granting, removing an admin stays a CLI task (`kegelmaster admin
//     revoke`), so a compromised admin account cannot lock out the others and
//     the last admin can never be removed via the web.

const (
	errMsgAdminUserNotFound = "Benutzer nicht gefunden"
	errMsgAdminClubNotFound = "Klub nicht gefunden"
)

// adminTargetUser loads the account addressed by :id. Invalid IDs are 404.
func (h *Handler) adminTargetUser(ctx context.Context, c fiber.Ctx) (user.User, error) {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return user.User{}, fiber.NewError(fiber.StatusNotFound, errMsgAdminUserNotFound)
	}
	u, err := h.UserRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return user.User{}, fiber.NewError(fiber.StatusNotFound, errMsgAdminUserNotFound)
		}
		slog.Error("admin get target user", "error", err, "target_user_id", id)
		return user.User{}, fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	return u, nil
}

// guardAdminTarget refuses actions that would lock out the acting admin or
// another instance admin.
func guardAdminTarget(c fiber.Ctx, target user.User, selfMsg string) error {
	if target.ID == adminUserID(c) {
		return fiber.NewError(fiber.StatusBadRequest, selfMsg)
	}
	if target.IsInstanceAdmin {
		return fiber.NewError(fiber.StatusConflict,
			"Instanz-Admins können nicht über die Weboberfläche gesperrt oder gelöscht werden. Bitte zuerst die Admin-Rolle per CLI entziehen (kegelmaster admin revoke).")
	}
	return nil
}

func adminRepoError(c fiber.Ctx, err error, op string) error {
	switch {
	case errors.Is(err, admin.ErrUserNotFound):
		return fiber.NewError(fiber.StatusNotFound, errMsgAdminUserNotFound)
	case errors.Is(err, admin.ErrClubNotFound):
		return fiber.NewError(fiber.StatusNotFound, errMsgAdminClubNotFound)
	}
	slog.Error(op, "error", err, "admin_user_id", adminUserID(c))
	return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
}

// HandleAdminDisableUser disables an account: login, cookie sessions and API
// tokens are rejected until it is enabled again. Its browser sessions are
// revoked as well, so enabling it later does not revive them.
func (h *Handler) HandleAdminDisableUser(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	ctx, cancel := h.RequestContext()
	defer cancel()

	target, err := h.adminTargetUser(ctx, c)
	if err != nil {
		return err
	}
	if err := guardAdminTarget(c, target, "Das eigene Konto kann nicht gesperrt werden"); err != nil {
		return err
	}
	if err := h.AdminRepo.DisableUser(ctx, target.ID, time.Now().UTC()); err != nil {
		return adminRepoError(c, err, "admin disable user")
	}
	audit.LogAdminAudit(c, adminUserID(c), audit.ActionAdminUserDisabled, "target_user_id", target.ID)
	return c.SendStatus(fiber.StatusNoContent)
}

// HandleAdminEnableUser lifts the disabling. The user has to log in again.
func (h *Handler) HandleAdminEnableUser(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	ctx, cancel := h.RequestContext()
	defer cancel()

	target, err := h.adminTargetUser(ctx, c)
	if err != nil {
		return err
	}
	if err := h.AdminRepo.EnableUser(ctx, target.ID, time.Now().UTC()); err != nil {
		return adminRepoError(c, err, "admin enable user")
	}
	audit.LogAdminAudit(c, adminUserID(c), audit.ActionAdminUserEnabled, "target_user_id", target.ID)
	return c.SendStatus(fiber.StatusNoContent)
}

// HandleAdminRevokeSessions logs the account out of every browser: sessions
// issued before now are rejected. API tokens are revoked individually.
func (h *Handler) HandleAdminRevokeSessions(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	ctx, cancel := h.RequestContext()
	defer cancel()

	target, err := h.adminTargetUser(ctx, c)
	if err != nil {
		return err
	}
	if target.ID == adminUserID(c) {
		return fiber.NewError(fiber.StatusBadRequest, "Die eigenen Sitzungen können hier nicht widerrufen werden. Bitte stattdessen das Passwort ändern.")
	}
	if err := h.AdminRepo.RevokeSessions(ctx, target.ID, time.Now().UTC()); err != nil {
		return adminRepoError(c, err, "admin revoke sessions")
	}
	audit.LogAdminAudit(c, adminUserID(c), audit.ActionAdminSessionsRevoked, "target_user_id", target.ID)
	return c.SendStatus(fiber.StatusNoContent)
}

// HandleAdminRevokeAPIToken deletes one API token of the account.
func (h *Handler) HandleAdminRevokeAPIToken(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	userID := c.Params("id")
	tokenID := c.Params("tokenId")
	if _, err := uuid.Parse(userID); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "API-Token nicht gefunden")
	}
	if _, err := uuid.Parse(tokenID); err != nil || h.APITokenRepo == nil {
		return fiber.NewError(fiber.StatusNotFound, "API-Token nicht gefunden")
	}
	ctx, cancel := h.RequestContext()
	defer cancel()

	if err := h.APITokenRepo.Delete(ctx, tokenID, userID); err != nil {
		if errors.Is(err, apitoken.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "API-Token nicht gefunden")
		}
		slog.Error("admin revoke api token", "error", err, "token_id", tokenID)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	audit.LogAdminAudit(c, adminUserID(c), audit.ActionAdminTokenRevoked, "target_user_id", userID, "token_id", tokenID)
	return c.SendStatus(fiber.StatusNoContent)
}

type adminDeleteUserRequest struct {
	// ConfirmEmail must repeat the account's e-mail address.
	ConfirmEmail string `json:"confirm_email"`
}

// HandleAdminDeleteUser deletes an account after the admin repeated its
// e-mail address. Owners must transfer their clubs first (409), otherwise the
// cascade on clubs.user_id would delete the clubs with all their data. Linked
// players stay in their clubs (unlinked), balances are not touched.
func (h *Handler) HandleAdminDeleteUser(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	var req adminDeleteUserRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}
	ctx, cancel := h.RequestContext()
	defer cancel()

	target, err := h.adminTargetUser(ctx, c)
	if err != nil {
		return err
	}
	if err := guardAdminTarget(c, target, "Das eigene Konto kann nicht gelöscht werden"); err != nil {
		return err
	}
	if !strings.EqualFold(strings.TrimSpace(req.ConfirmEmail), target.Email) {
		return fiber.NewError(fiber.StatusBadRequest, "Die eingegebene E-Mail-Adresse stimmt nicht überein")
	}
	if err := h.AdminRepo.DeleteUser(ctx, target.ID); err != nil {
		if errors.Is(err, admin.ErrUserOwnsClubs) {
			return fiber.NewError(fiber.StatusConflict, "Der Benutzer ist noch Owner von Klubs. Bitte zuerst den Owner dieser Klubs übertragen.")
		}
		return adminRepoError(c, err, "admin delete user")
	}
	audit.LogAdminAudit(c, adminUserID(c), audit.ActionAdminUserDeleted, "target_user_id", target.ID, "target_email", target.Email)
	return c.SendStatus(fiber.StatusNoContent)
}

// adminTargetClub loads the club addressed by :id. Invalid IDs are 404.
func (h *Handler) adminTargetClub(ctx context.Context, c fiber.Ctx) (club.Club, error) {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return club.Club{}, fiber.NewError(fiber.StatusNotFound, errMsgAdminClubNotFound)
	}
	cl, err := h.ClubRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return club.Club{}, fiber.NewError(fiber.StatusNotFound, errMsgAdminClubNotFound)
		}
		slog.Error("admin get target club", "error", err, "club_id", id)
		return club.Club{}, fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	return cl, nil
}

type adminTransferClubOwnerRequest struct {
	NewOwnerEmail string `json:"new_owner_email"`
}

// HandleAdminTransferClubOwner makes another account the owner of a club,
// e.g. when the owner is no longer reachable or is about to be deleted. Same
// logic as HandleTransferClubOwner, without owner and password check. The
// admin does not become a member of the club.
func (h *Handler) HandleAdminTransferClubOwner(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	var req adminTransferClubOwnerRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}
	email := strings.TrimSpace(req.NewOwnerEmail)
	if email == "" || !strings.Contains(email, "@") {
		return fiber.NewError(fiber.StatusBadRequest, "Gültige E-Mail-Adresse des neuen Eigentümers ist erforderlich")
	}
	ctx, cancel := h.RequestContext()
	defer cancel()

	cl, err := h.adminTargetClub(ctx, c)
	if err != nil {
		return err
	}
	updated, newOwner, err := h.transferClubOwner(ctx, cl.ID, cl.UserID, email)
	if err != nil {
		return err
	}
	audit.LogAdminAudit(c, adminUserID(c), audit.ActionAdminClubOwnerTransferred,
		"club_id", cl.ID, "previous_owner_id", cl.UserID, "new_owner_id", newOwner.ID)
	return c.JSON(adminTransferClubOwnerResponse{
		ClubID: updated.ID,
		Owner:  admin.UserRef{ID: newOwner.ID, Email: newOwner.Email},
	})
}

type adminTransferClubOwnerResponse struct {
	ClubID string        `json:"club_id"`
	Owner  admin.UserRef `json:"owner"`
}

type adminDeleteClubRequest struct {
	// ConfirmName must repeat the club's name.
	ConfirmName string `json:"confirm_name"`
}

// HandleAdminDeleteClub deletes a club with all its data (players, game days,
// transactions, … cascade) after the admin repeated the club's name. Nothing
// outside the club holds a balance of it, so the ledger stays consistent.
func (h *Handler) HandleAdminDeleteClub(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}
	var req adminDeleteClubRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}
	ctx, cancel := h.RequestContext()
	defer cancel()

	cl, err := h.adminTargetClub(ctx, c)
	if err != nil {
		return err
	}
	if strings.TrimSpace(req.ConfirmName) != strings.TrimSpace(cl.Name) {
		return fiber.NewError(fiber.StatusBadRequest, "Der eingegebene Klubname stimmt nicht überein")
	}
	if err := h.ClubRepo.Delete(ctx, cl.ID); err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, errMsgAdminClubNotFound)
		}
		slog.Error("admin delete club", "error", err, "club_id", cl.ID)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	audit.LogAdminAudit(c, adminUserID(c), audit.ActionAdminClubDeleted,
		"club_id", cl.ID, "club_name", cl.Name, "owner_id", cl.UserID)
	return c.SendStatus(fiber.StatusNoContent)
}
