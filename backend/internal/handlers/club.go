package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

type createClubRequest struct {
	Name               string `json:"name"`
	Balance            int    `json:"balance"`
	StartBalance      int    `json:"start_balance"`
	BaseFee           int    `json:"base_fee"`
	AutoTipEnabled    *bool  `json:"auto_tip_enabled,omitempty"`    // Optional, defaults to true
	CouplesModeEnabled *bool  `json:"couples_mode_enabled,omitempty"` // Optional, defaults to false
}

func (r createClubRequest) validate() error {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return errors.New("Name ist erforderlich")
	}
	if r.Balance < 0 {
		return errors.New("Balance darf nicht negativ sein")
	}
	if r.BaseFee < 0 {
		return errors.New("BaseFee darf nicht negativ sein")
	}
	return nil
}

func (h *Handler) HandleCreateClub(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	var req createClubRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if err := req.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Default auto_tip_enabled to true if not provided
	autoTipEnabled := true
	if req.AutoTipEnabled != nil {
		autoTipEnabled = *req.AutoTipEnabled
	}
	couplesModeEnabled := false
	if req.CouplesModeEnabled != nil {
		couplesModeEnabled = *req.CouplesModeEnabled
	}

	clubEntity, err := h.ClubRepo.Create(ctx, club.CreateClubParams{
		Name:               strings.TrimSpace(req.Name),
		Balance:            req.Balance,
		StartBalance:       req.StartBalance,
		BaseFee:            req.BaseFee,
		AutoTipEnabled:     autoTipEnabled,
		CouplesModeEnabled: couplesModeEnabled,
		UserID:             u.ID,
	})
	if err != nil {
		slog.Error("create club", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.Status(fiber.StatusCreated).JSON(ClubResponseFromEntity(clubEntity))
}

func (h *Handler) HandleGetClubs(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubs, err := h.ClubRepo.GetForUser(ctx, u.ID)
	if err != nil {
		slog.Error("get clubs", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(ClubsResponseFromEntities(clubs))
}

func (h *Handler) HandleGetClub(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("id")
	if clubID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID ist erforderlich")
	}

	clubEntity, err := h.ClubRepo.GetByID(ctx, clubID)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("get club", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Only owner or members (user has a player in this club) may access club details
	if clubEntity.UserID != u.ID {
		_, err := h.PlayerRepo.GetByUserIDAndClubID(ctx, u.ID, clubID)
		if err != nil {
			return fiber.NewError(fiber.StatusForbidden, "Keine Berechtigung, diesen Club einzusehen")
		}
	}

	return c.JSON(ClubResponseFromEntity(clubEntity))
}

type updateClubRequest struct {
	Name               string `json:"name"`
	Balance            int    `json:"balance"`
	StartBalance       int    `json:"start_balance"`
	BaseFee            int    `json:"base_fee"`
	AutoTipEnabled     bool   `json:"auto_tip_enabled"`
	CouplesModeEnabled bool   `json:"couples_mode_enabled"`
}

func (r updateClubRequest) validate() error {
	name := strings.TrimSpace(r.Name)
	if name == "" {
		return errors.New("Name ist erforderlich")
	}
	if r.Balance < 0 {
		return errors.New("Balance darf nicht negativ sein")
	}
	if r.BaseFee < 0 {
		return errors.New("BaseFee darf nicht negativ sein")
	}
	return nil
}

func (h *Handler) HandleUpdateClub(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("id")
	if clubID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID ist erforderlich")
	}

	// Prüfe, ob der Club existiert und der User der Owner ist
	existingClub, err := h.ClubRepo.GetByID(ctx, clubID)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("get club", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Owner-Check: Nur der Owner darf updaten
	if existingClub.UserID != u.ID {
		return fiber.NewError(fiber.StatusForbidden, "Nur der Owner darf diesen Club aktualisieren")
	}

	var req updateClubRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}

	if err := req.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	updatedClub, err := h.ClubRepo.Update(ctx, club.UpdateClubParams{
		ID:                 clubID,
		Name:               strings.TrimSpace(req.Name),
		Balance:            req.Balance,
		StartBalance:       req.StartBalance,
		BaseFee:            req.BaseFee,
		AutoTipEnabled:     req.AutoTipEnabled,
		CouplesModeEnabled: req.CouplesModeEnabled,
	})
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("update club", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(ClubResponseFromEntity(updatedClub))
}

type deleteClubRequest struct {
	Password string `json:"password"`
}

func (h *Handler) HandleDeleteClub(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("id")
	if clubID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID ist erforderlich")
	}

	// Prüfe, ob der Club existiert und der User der Owner ist
	existingClub, err := h.ClubRepo.GetByID(ctx, clubID)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("get club", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	// Owner-Check: Nur der Owner darf löschen
	if existingClub.UserID != u.ID {
		return fiber.NewError(fiber.StatusForbidden, "Nur der Owner darf diesen Club löschen")
	}

	var req deleteClubRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}
	pw := strings.TrimSpace(req.Password)
	if pw == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Passwort zur Bestätigung ist erforderlich")
	}
	if len(pw) > maxPasswordLength {
		return fiber.NewError(fiber.StatusBadRequest, "Passwort darf höchstens 128 Zeichen haben")
	}

	if err := h.AuthSvc.ComparePassword(u.PasswordHash, pw); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Ungültiges Passwort")
	}

	if err := h.ClubRepo.Delete(ctx, clubID); err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("delete club", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

type transferClubOwnerRequest struct {
	NewOwnerEmail string `json:"new_owner_email"`
	Password      string `json:"password"`
}

func (r transferClubOwnerRequest) validate() error {
	email := strings.TrimSpace(strings.ToLower(r.NewOwnerEmail))
	if email == "" || !strings.Contains(email, "@") {
		return errors.New("Gültige E-Mail-Adresse des neuen Eigentümers ist erforderlich")
	}
	pw := strings.TrimSpace(r.Password)
	if pw == "" {
		return errors.New("Passwort zur Bestätigung ist erforderlich")
	}
	if len(pw) > maxPasswordLength {
		return errors.New("Passwort darf höchstens 128 Zeichen haben")
	}
	return nil
}

func (h *Handler) HandleTransferClubOwner(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("id")
	if clubID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID ist erforderlich")
	}

	existingClub, err := h.ClubRepo.GetByID(ctx, clubID)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("get club", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if existingClub.UserID != u.ID {
		return fiber.NewError(fiber.StatusForbidden, "Nur der Owner darf den Eigentümer wechseln")
	}

	var req transferClubOwnerRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	}
	if err := req.validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if err := h.AuthSvc.ComparePassword(u.PasswordHash, strings.TrimSpace(req.Password)); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Ungültiges Passwort")
	}

	newOwnerEmail := strings.ToLower(strings.TrimSpace(req.NewOwnerEmail))
	newOwner, err := h.UserRepo.GetByEmail(ctx, newOwnerEmail)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return fiber.NewError(fiber.StatusBadRequest, "User mit dieser E-Mail existiert nicht")
		}
		slog.Error("get user by email", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if newOwner.ID == u.ID {
		return fiber.NewError(fiber.StatusBadRequest, "Der neue Eigentümer muss ein anderer User sein")
	}

	updatedClub, err := h.ClubRepo.UpdateOwner(ctx, clubID, newOwner.ID)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("update club owner", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(ClubResponseFromEntity(updatedClub))
}

// HandleRecalculateClubBalance recalculates club balance from start_balance + SUM(amount) of all transactions. Only club owner may call.
func (h *Handler) HandleRecalculateClubBalance(c fiber.Ctx) error {
	if err := h.ValidateCSRF(c); err != nil {
		return err
	}

	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("id")
	if clubID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID ist erforderlich")
	}

	existingClub, err := h.ClubRepo.GetByID(ctx, clubID)
	if err != nil {
		if errors.Is(err, club.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Club nicht gefunden")
		}
		slog.Error("get club", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	if existingClub.UserID != u.ID {
		return fiber.NewError(fiber.StatusForbidden, "Nur der Owner darf die Club-Balance neu berechnen")
	}

	if err := h.TransactionRepo.RecalculateClubBalance(ctx, clubID); err != nil {
		slog.Error("recalculate club balance", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	updatedClub, err := h.ClubRepo.GetByID(ctx, clubID)
	if err != nil {
		slog.Error("get club after recalc", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(ClubResponseFromEntity(updatedClub))
}
