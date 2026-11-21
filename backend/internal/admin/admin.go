// Package admin provides the read-only operator overview of an instance
// (instance admins only): users, clubs, owners and player↔user links.
//
// It deliberately exposes no business data of the clubs (balances,
// transactions, penalties of individual players), at most aggregated counts.
// Access control lives in handlers.RequireInstanceAdmin and is independent of
// the club PermissionChecker.
package admin

import "time"

// Thresholds shared by the queries in queries/admin.sql (keep in sync).
const (
	// OrphanedAfterDays: an account without club and player link counts as
	// orphaned once it is older than this.
	OrphanedAfterDays = 30
	// ActiveClubDays: a club is active with a game day or transaction in this window.
	ActiveClubDays = 90
	// InactiveClubDays: a club without activity for this long is flagged.
	InactiveClubDays = 180
	// UnusedAPITokenDays: an active token unused for this long is flagged.
	UnusedAPITokenDays = 90
	// RegistrationSpikeThreshold: more registrations than this on one day (UTC)
	// within the last 30 days are flagged.
	RegistrationSpikeThreshold = 10
)

// Stats are the key figures of the dashboard.
type Stats struct {
	Users           UserStats     `json:"users"`
	Clubs           ClubStats     `json:"clubs"`
	Players         PlayerStats   `json:"players"`
	OpenInvitations int64         `json:"open_invitations"`
	APITokens       APITokenStats `json:"api_tokens"`
	Passkeys        int64         `json:"passkeys"`
	Anomalies       Anomalies     `json:"anomalies"`
}

type UserStats struct {
	Total          int64 `json:"total"`
	New7d          int64 `json:"new_7d"`
	New30d         int64 `json:"new_30d"`
	Active30d      int64 `json:"active_30d"`
	InstanceAdmins int64 `json:"instance_admins"`
	Disabled       int64 `json:"disabled"`
}

type ClubStats struct {
	Total     int64 `json:"total"`
	Active90d int64 `json:"active_90d"`
}

type PlayerStats struct {
	Total  int64 `json:"total"`
	Linked int64 `json:"linked"`
}

type APITokenStats struct {
	// Active: not expired (including tokens without expiry).
	Active int64 `json:"active"`
}

// Anomalies are hints for the operator; each is a count except the spikes.
type Anomalies struct {
	OrphanedUsers          int64               `json:"orphaned_users"`
	InactiveClubs          int64               `json:"inactive_clubs"`
	APITokensWithoutExpiry int64               `json:"api_tokens_without_expiry"`
	APITokensUnused90d     int64               `json:"api_tokens_unused_90d"`
	RegistrationSpikes     []RegistrationSpike `json:"registration_spikes"`
}

type RegistrationSpike struct {
	// Date is the UTC day (YYYY-MM-DD).
	Date          string `json:"date"`
	Registrations int64  `json:"registrations"`
}

// Page is one page of a list.
type Page[T any] struct {
	Data       []T   `json:"data"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// UserListItem is one row of the user list.
type UserListItem struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	IsInstanceAdmin bool       `json:"is_instance_admin"`
	Orphaned        bool       `json:"orphaned"`
	CreatedAt       time.Time  `json:"created_at"`
	LastLoginAt     *time.Time `json:"last_login_at"`
	// DisabledAt is set while the account is disabled.
	DisabledAt      *time.Time `json:"disabled_at"`
	OwnedClubCount  int64      `json:"owned_club_count"`
	PlayerClubCount int64      `json:"player_club_count"`
	PasskeyCount    int64      `json:"passkey_count"`
	APITokenCount   int64      `json:"api_token_count"`
}

// UserDetail is the account with its clubs, player links, tokens and passkeys.
type UserDetail struct {
	ID                string     `json:"id"`
	Email             string     `json:"email"`
	IsInstanceAdmin   bool       `json:"is_instance_admin"`
	Orphaned          bool       `json:"orphaned"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	PasswordChangedAt *time.Time `json:"password_changed_at"`
	LastLoginAt       *time.Time `json:"last_login_at"`
	DisabledAt        *time.Time `json:"disabled_at"`
	// SessionsValidAfter: browser sessions issued before this were revoked.
	SessionsValidAfter *time.Time       `json:"sessions_valid_after"`
	OwnedClubs         []OwnedClub      `json:"owned_clubs"`
	Players            []UserPlayerLink `json:"players"`
	APITokens          []APITokenInfo   `json:"api_tokens"`
	Passkeys           []PasskeyInfo    `json:"passkeys"`
}

type OwnedClub struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	PlayerCount int64     `json:"player_count"`
}

// UserPlayerLink is a player linked to the user, seen from the user.
type UserPlayerLink struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	Inactive   bool   `json:"inactive"`
	ClubID     string `json:"club_id"`
	ClubName   string `json:"club_name"`
	// IsOwner: the user also owns this club.
	IsOwner  bool    `json:"is_owner"`
	RoleID   string  `json:"role_id"`
	RoleName *string `json:"role_name"`
}

// APITokenInfo never contains the token hash.
type APITokenInfo struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	TokenPrefix string     `json:"token_prefix"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
}

type PasskeyInfo struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}

// UserRef identifies an account by ID and e-mail.
type UserRef struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// ClubListItem is one row of the club list.
type ClubListItem struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	CreatedAt         time.Time  `json:"created_at"`
	Owner             UserRef    `json:"owner"`
	PlayerCount       int64      `json:"player_count"`
	LinkedPlayerCount int64      `json:"linked_player_count"`
	GameDayCount      int64      `json:"game_day_count"`
	LastActivityAt    *time.Time `json:"last_activity_at"`
}

// ClubDetail is a club with players, open invitations and roles, but without
// any business data (balances, transactions, penalties).
type ClubDetail struct {
	ID                string           `json:"id"`
	Name              string           `json:"name"`
	CreatedAt         time.Time        `json:"created_at"`
	Owner             UserRef          `json:"owner"`
	PlayerCount       int64            `json:"player_count"`
	LinkedPlayerCount int64            `json:"linked_player_count"`
	GameDayCount      int64            `json:"game_day_count"`
	TransactionCount  int64            `json:"transaction_count"`
	LastActivityAt    *time.Time       `json:"last_activity_at"`
	Players           []ClubPlayer     `json:"players"`
	OpenInvitations   []OpenInvitation `json:"open_invitations"`
	Roles             []ClubRole       `json:"roles"`
}

// ClubPlayer is a player of the club with its linked account (if any).
type ClubPlayer struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Inactive  bool      `json:"inactive"`
	CreatedAt time.Time `json:"created_at"`
	// User is null when the player is not linked to an account.
	User     *UserRef `json:"user"`
	RoleID   string   `json:"role_id"`
	RoleName *string  `json:"role_name"`
}

type OpenInvitation struct {
	ID         string    `json:"id"`
	PlayerID   string    `json:"player_id"`
	PlayerName string    `json:"player_name"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type ClubRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PlayerCount int64  `json:"player_count"`
}

// Membership is one row of the memberships CSV export.
type Membership struct {
	UserEmail  string
	ClubName   string
	IsOwner    bool
	PlayerName string
	RoleName   string
}

// InstanceAdmin is an account with the instance admin flag (CLI listing).
type InstanceAdmin struct {
	ID          string
	Email       string
	CreatedAt   time.Time
	LastLoginAt *time.Time
}
