package invitation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
)

type Repository struct {
	queries    *db.Queries
	database   *sql.DB
	playerRepo *player.Repository
}

func NewRepository(dbConn *sql.DB, playerRepo *player.Repository) *Repository {
	return &Repository{
		queries:    db.New(dbConn),
		database:   dbConn,
		playerRepo: playerRepo,
	}
}

type CreateInvitationParams struct {
	PlayerID string
	Email    string
}

// Create creates a new player invitation with a generated token.
// The invitation expires after 7 days.
func (r *Repository) Create(ctx context.Context, params CreateInvitationParams) (PlayerInvitation, error) {
	slog.Debug("creating invitation", "player_id", params.PlayerID, "email", params.Email)

	token, err := GenerateToken()
	if err != nil {
		slog.Error("failed to generate token", "error", err)
		return PlayerInvitation{}, err
	}
	slog.Debug("token generated", "token_preview", token[:min(8, len(token))]+"...")

	now := time.Now().UTC()
	expiresAt := now.Add(7 * 24 * time.Hour) // 7 days

	invitationID := uuid.NewString()
	dbInvitation, err := r.queries.CreatePlayerInvitation(ctx, db.CreatePlayerInvitationParams{
		ID:        invitationID,
		PlayerID:  params.PlayerID,
		Email:     params.Email,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	})
	if err != nil {
		slog.Error("failed to create invitation in database", "error", err, "player_id", params.PlayerID, "email", params.Email)
		return PlayerInvitation{}, err
	}

	inv := dbInvitationToInvitation(dbInvitation)
	slog.Info("invitation created", "invitation_id", inv.ID, "player_id", params.PlayerID, "email", params.Email, "expires_at", inv.ExpiresAt)
	return inv, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetByToken retrieves an invitation by its token.
func (r *Repository) GetByToken(ctx context.Context, token string) (PlayerInvitation, error) {
	slog.Debug("getting invitation by token", "token_preview", token[:min(8, len(token))]+"...")

	dbInvitation, err := r.queries.GetPlayerInvitationByToken(ctx, token)
	if errors.Is(err, sql.ErrNoRows) {
		slog.Warn("invitation not found by token", "token_preview", token[:min(8, len(token))]+"...")
		return PlayerInvitation{}, ErrNotFound
	}
	if err != nil {
		slog.Error("failed to get invitation by token", "error", err, "token_preview", token[:min(8, len(token))]+"...")
		return PlayerInvitation{}, err
	}

	invitation := dbInvitationToInvitation(dbInvitation)
	slog.Debug("invitation found", "invitation_id", invitation.ID, "player_id", invitation.PlayerID, "email", invitation.Email)

	// Check if expired
	if time.Now().UTC().After(invitation.ExpiresAt) {
		slog.Warn("invitation expired", "invitation_id", invitation.ID, "expires_at", invitation.ExpiresAt)
		return PlayerInvitation{}, ErrExpired
	}

	// Check if already accepted
	if invitation.AcceptedAt != nil {
		slog.Warn("invitation already accepted", "invitation_id", invitation.ID, "accepted_at", invitation.AcceptedAt)
		return PlayerInvitation{}, ErrAlreadyAccepted
	}

	slog.Debug("invitation is valid", "invitation_id", invitation.ID)
	return invitation, nil
}

// Accept marks an invitation as accepted and links the player to the user.
// Everything runs in one database transaction with conditional updates, so an
// invitation can be used only once, a player that already has a user is never
// re-linked, and the player's balance is not touched.
func (r *Repository) Accept(ctx context.Context, invitationID, userID string) error {
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()
	qtx := r.queries.WithTx(tx)

	dbInvitation, err := qtx.GetPlayerInvitationByID(ctx, invitationID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("get invitation: %w", err)
	}
	inv := dbInvitationToInvitation(dbInvitation)

	playerRow, err := qtx.GetPlayerByIDForUpdate(ctx, inv.PlayerID)
	if err != nil {
		return fmt.Errorf("get player: %w", err)
	}
	if playerRow.UserID != nil {
		return ErrPlayerAlreadyLinked
	}
	uid := userID
	if _, err := qtx.GetPlayerByUserIDAndClubID(ctx, db.GetPlayerByUserIDAndClubIDParams{UserID: &uid, ClubID: playerRow.ClubID}); err == nil {
		return ErrUserAlreadyInClub
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check existing player: %w", err)
	}

	now := time.Now().UTC()
	n, err := qtx.MarkPlayerInvitationAccepted(ctx, db.MarkPlayerInvitationAcceptedParams{
		AcceptedAt: sql.NullTime{Time: now, Valid: true},
		ID:         invitationID,
	})
	if err != nil {
		return fmt.Errorf("mark invitation accepted: %w", err)
	}
	if n == 0 {
		if inv.AcceptedAt != nil {
			return ErrAlreadyAccepted
		}
		if now.After(inv.ExpiresAt) {
			return ErrExpired
		}
		// Accepted concurrently by another request.
		return ErrAlreadyAccepted
	}

	n, err = qtx.LinkPlayerToUser(ctx, db.LinkPlayerToUserParams{
		UserID:    &uid,
		UpdatedAt: now,
		ID:        inv.PlayerID,
	})
	if err != nil {
		return fmt.Errorf("link player: %w", err)
	}
	if n == 0 {
		return ErrPlayerAlreadyLinked
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	slog.Info("invitation accepted and player linked",
		"invitation_id", invitationID,
		"player_id", inv.PlayerID,
		"user_id", userID)
	return nil
}

// GetByPlayerID retrieves all invitations for a player.
func (r *Repository) GetByPlayerID(ctx context.Context, playerID string) ([]PlayerInvitation, error) {
	dbInvitations, err := r.queries.GetPlayerInvitationsByPlayerID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	invitations := make([]PlayerInvitation, len(dbInvitations))
	for i, dbInv := range dbInvitations {
		invitations[i] = dbInvitationToInvitation(dbInv)
	}

	return invitations, nil
}

// dbInvitationToInvitation converts a db.PlayerInvitation to an invitation.PlayerInvitation
func dbInvitationToInvitation(dbInv db.PlayerInvitation) PlayerInvitation {
	var acceptedAt *time.Time
	if dbInv.AcceptedAt.Valid {
		acceptedAt = &dbInv.AcceptedAt.Time
	}

	return PlayerInvitation{
		ID:         dbInv.ID,
		PlayerID:   dbInv.PlayerID,
		Email:      dbInv.Email,
		Token:      dbInv.Token,
		ExpiresAt:  dbInv.ExpiresAt,
		AcceptedAt: acceptedAt,
		CreatedAt:  dbInv.CreatedAt,
	}
}
