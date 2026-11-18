package handlers

import (
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
	"github.com/schnurbus/go-kegelmaster/backend/internal/statistics"
)

// HandleGetMyStatistics returns comprehensive statistics for the current user's player in a club.
// Query params from/to (YYYY-MM-DD) are optional: without from all published game days are
// considered, to defaults to today. Only published (non-draft) game days are evaluated.
func (h *Handler) HandleGetMyStatistics(c fiber.Ctx) error {
	ctx, cancel := h.RequestContext()
	defer cancel()

	u, err := h.UserFromCookie(ctx, c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Nicht angemeldet")
	}

	clubID := c.Params("clubId")
	if clubID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Club-ID ist erforderlich")
	}

	var from *time.Time
	if v := c.Query("from"); v != "" {
		t, err := time.Parse(statistics.DateLayout, v)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Ungültiges Startdatum (erwartet YYYY-MM-DD)")
		}
		from = &t
	}
	now := time.Now().UTC()
	to := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if v := c.Query("to"); v != "" {
		t, err := time.Parse(statistics.DateLayout, v)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Ungültiges Enddatum (erwartet YYYY-MM-DD)")
		}
		to = t
	}
	if from != nil && from.After(to) {
		return fiber.NewError(fiber.StatusBadRequest, "Startdatum muss vor dem Enddatum liegen")
	}

	playerEntity, err := h.PlayerRepo.GetByUserIDAndClubID(ctx, u.ID, clubID)
	if err != nil {
		if errors.Is(err, player.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Kein Spieler in diesem Club")
		}
		slog.Error("get my player for statistics", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}
	if playerEntity.ClubID != clubID {
		return fiber.NewError(fiber.StatusNotFound, "Kein Spieler in diesem Club")
	}

	gender := ""
	if playerEntity.Gender != nil {
		gender = *playerEntity.Gender
	}
	stats, err := h.StatisticsRepo.ForPlayer(ctx, statistics.PlayerParams{
		ClubID:       clubID,
		PlayerID:     playerEntity.ID,
		PlayerName:   playerEntity.Name,
		PlayerGender: gender,
		StartBalance: playerEntity.StartBalance,
		From:         from,
		To:           to,
	})
	if err != nil {
		slog.Error("compute player statistics", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Interner Fehler")
	}

	return c.JSON(stats)
}
