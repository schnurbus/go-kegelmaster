package audit

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
)

// Action names for audit log events.
const (
	ActionUserRegistered       = "user_registered"
	ActionUserLogin            = "user_login"
	ActionUserPasswordReset    = "user_password_reset"
	ActionClubCreated          = "club_created"
	ActionClubUpdated          = "club_updated"
	ActionClubDeleted          = "club_deleted"
	ActionClubOwnerTransferred = "club_owner_transferred"
	ActionClubBalanceRecalc    = "club_balance_recalculated"
	ActionRoleCreated          = "role_created"
	ActionRoleUpdated          = "role_updated"
	ActionRoleDeleted          = "role_deleted"
	ActionPermissionAdded      = "permission_added"
	ActionPermissionRemoved    = "permission_removed"
	ActionPlayerCreated        = "player_created"
	ActionPlayerUpdated        = "player_updated"
	ActionPlayerDeleted        = "player_deleted"
	ActionPlayerBalanceRecalc  = "player_balance_recalculated"
	ActionPlayerInvited        = "player_invited"
	ActionInvitationAccepted   = "invitation_accepted"
	ActionPenaltyTypeCreated    = "penalty_type_created"
	ActionPenaltyTypeUpdated   = "penalty_type_updated"
	ActionPenaltyTypeDeleted   = "penalty_type_deleted"
	ActionPenaltyTypeOrder     = "penalty_type_display_order_updated"
	ActionCompetitionCreated   = "competition_created"
	ActionCompetitionUpdated   = "competition_updated"
	ActionCompetitionDeleted   = "competition_deleted"
	ActionGameDayCreated       = "gameday_created"
	ActionGameDayUpdated       = "gameday_updated"
	ActionGameDayDeleted       = "gameday_deleted"
	ActionParticipantAdded     = "gameday_participant_added"
	ActionParticipantRemoved  = "gameday_participant_removed"
	ActionFeesUpdated          = "gameday_participant_fees_updated"
	ActionCompetitionValues    = "gameday_participant_competition_values_updated"
	ActionTransactionCreated   = "transaction_created"
	ActionTransactionDeleted   = "transaction_deleted"
)

// LogAudit writes an audit log entry. userID may be empty; then "anonymous" is logged for unauthenticated actions (e.g. register).
// keyValues are alternating key-value pairs (e.g. "club_id", clubID, "player_id", playerID).
func LogAudit(c fiber.Ctx, userID, action string, keyValues ...any) {
	if userID == "" {
		userID = "anonymous"
	}
	attrs := make([]any, 0, 4+len(keyValues))
	attrs = append(attrs, "event", "audit", "action", action, "user_id", userID)
	attrs = append(attrs, keyValues...)
	slog.InfoContext(c.Context(), "audit", attrs...)
}
