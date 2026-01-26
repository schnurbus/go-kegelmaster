package invitation

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
)

type Repository struct {
	queries      *db.Queries
	playerRepo   *player.Repository
}

func NewRepository(dbConn *sql.DB, playerRepo *player.Repository) *Repository {
	return &Repository{
		queries:    db.New(dbConn),
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
func (r *Repository) Accept(ctx context.Context, invitationID, userID string) error {
	slog.Info("accepting invitation", "invitation_id", invitationID, "user_id", userID)
	
	// Get invitation to verify it exists and get player ID
	dbInvitation, err := r.queries.GetPlayerInvitationByID(ctx, invitationID)
	if errors.Is(err, sql.ErrNoRows) {
		slog.Warn("invitation not found by ID", "invitation_id", invitationID)
		return ErrNotFound
	}
	if err != nil {
		slog.Error("failed to get invitation by ID", "error", err, "invitation_id", invitationID)
		return err
	}

	invitation := dbInvitationToInvitation(dbInvitation)
	slog.Debug("invitation retrieved", "invitation_id", invitation.ID, "player_id", invitation.PlayerID, "email", invitation.Email)

	// Check if expired
	if time.Now().UTC().After(invitation.ExpiresAt) {
		slog.Warn("invitation expired", "invitation_id", invitation.ID, "expires_at", invitation.ExpiresAt)
		return ErrExpired
	}

	// Check if already accepted
	if invitation.AcceptedAt != nil {
		slog.Warn("invitation already accepted", "invitation_id", invitation.ID, "accepted_at", invitation.AcceptedAt)
		return ErrAlreadyAccepted
	}

	// Mark invitation as accepted
	now := time.Now().UTC()
	slog.Debug("marking invitation as accepted", "invitation_id", invitationID, "accepted_at", now)
	err = r.queries.AcceptPlayerInvitation(ctx, db.AcceptPlayerInvitationParams{
		AcceptedAt: sql.NullTime{Time: now, Valid: true},
		ID:         invitationID,
	})
	if err != nil {
		slog.Error("failed to mark invitation as accepted", "error", err, "invitation_id", invitationID)
		return err
	}
	slog.Debug("invitation marked as accepted", "invitation_id", invitationID)

	// Update player to link with user
	slog.Debug("updating player to link with user", "player_id", invitation.PlayerID, "user_id", userID)
	existingPlayer, err := r.playerRepo.GetByID(ctx, invitation.PlayerID)
	if err != nil {
		slog.Error("failed to get player for linking", "error", err, "player_id", invitation.PlayerID)
		return err
	}

	_, err = r.playerRepo.Update(ctx, player.UpdatePlayerParams{
		ID:           existingPlayer.ID,
		Name:         existingPlayer.Name,
		Balance:      existingPlayer.Balance,
		StartBalance: existingPlayer.StartBalance,
		UserID:       &userID,
		RoleID:       existingPlayer.RoleID,
	})
	if err != nil {
		slog.Error("failed to update player with user", "error", err, "player_id", invitation.PlayerID, "user_id", userID)
		return err
	}
	
	slog.Info("invitation accepted and player linked", 
		"invitation_id", invitationID, 
		"player_id", invitation.PlayerID, 
		"user_id", userID,
		"email", invitation.Email)
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
