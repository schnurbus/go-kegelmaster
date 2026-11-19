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
	ActionUserPasswordChanged  = "user_password_changed"
	ActionUserSettingsUpdated  = "user_settings_updated"
	ActionAPITokenCreated      = "api_token_created"
	ActionAPITokenDeleted      = "api_token_deleted"
	ActionUserLoginPasskey     = "user_login_passkey"
	ActionUserLoginRejected    = "user_login_rejected"
	ActionPasskeyCreated       = "passkey_created"
	ActionPasskeyRenamed       = "passkey_renamed"
	ActionPasskeyDeleted       = "passkey_deleted"
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
	ActionPenaltyTypeCreated   = "penalty_type_created"
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
	ActionParticipantRemoved   = "gameday_participant_removed"
	ActionParticipantAttended  = "gameday_participant_attendance"
	ActionFeesUpdated          = "gameday_participant_fees_updated"
	ActionCompetitionValues    = "gameday_participant_competition_values_updated"
	ActionTransactionCreated   = "transaction_created"
	ActionTransactionDeleted   = "transaction_deleted"
	ActionTransactionUpdated   = "transaction_updated"

	// Instance admin (operator) actions; logged with LogAdminAudit.
	ActionAdminStatsViewed = "admin_stats_viewed"
	ActionAdminUsersListed = "admin_users_listed"
	ActionAdminUserViewed  = "admin_user_viewed"
	ActionAdminClubsListed = "admin_clubs_listed"
	ActionAdminClubViewed  = "admin_club_viewed"
	ActionAdminExport      = "admin_export"

	// Instance admin interventions (Phase 3).
	ActionAdminUserDisabled         = "admin_user_disabled"
	ActionAdminUserEnabled          = "admin_user_enabled"
	ActionAdminSessionsRevoked      = "admin_sessions_revoked"
	ActionAdminTokenRevoked         = "admin_token_revoked"
	ActionAdminClubOwnerTransferred = "admin_club_owner_transferred"
	ActionAdminUserDeleted          = "admin_user_deleted"
	ActionAdminClubDeleted          = "admin_club_deleted"
)

// LogAudit writes an audit log entry, including the client IP (c.IP() honours
// PROXY_HEADER/TRUSTED_PROXIES). userID may be empty; then "anonymous" is logged for unauthenticated actions (e.g. register).
// keyValues are alternating key-value pairs (e.g. "club_id", clubID, "player_id", playerID).
func LogAudit(c fiber.Ctx, userID, action string, keyValues ...any) {
	if userID == "" {
		userID = "anonymous"
	}
	attrs := make([]any, 0, 8+len(keyValues))
	attrs = append(attrs, "event", "audit", "action", action, "user_id", userID, "ip", c.IP())
	attrs = append(attrs, keyValues...)
	slog.InfoContext(c.Context(), "audit", attrs...)
}

// LogAdminAudit writes an audit entry for an instance admin access. Every
// entry carries admin=true, so operator accesses can be filtered in the log
// stack. Read accesses are audited as well.
func LogAdminAudit(c fiber.Ctx, adminUserID, action string, keyValues ...any) {
	LogAudit(c, adminUserID, action, append([]any{"admin", true}, keyValues...)...)
}
