package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/schnurbus/go-kegelmaster/backend/internal/db"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrClubNotFound = errors.New("club not found")
)

// Repository runs the aggregating admin queries (queries/admin.sql). It does
// not reuse the club-scoped repositories on purpose.
type Repository struct {
	queries *db.Queries
}

func NewRepository(dbConn *sql.DB) *Repository {
	return &Repository{queries: db.New(dbConn)}
}

// Stats returns the dashboard figures and anomalies.
func (r *Repository) Stats(ctx context.Context) (Stats, error) {
	row, err := r.queries.GetAdminStats(ctx)
	if err != nil {
		return Stats{}, err
	}
	spikeRows, err := r.queries.ListAdminRegistrationSpikes(ctx, RegistrationSpikeThreshold)
	if err != nil {
		return Stats{}, err
	}
	spikes := make([]RegistrationSpike, 0, len(spikeRows))
	for _, s := range spikeRows {
		spikes = append(spikes, RegistrationSpike{Date: s.Day.Format(time.DateOnly), Registrations: s.Registrations})
	}
	return Stats{
		Users: UserStats{
			Total:          row.UsersTotal,
			New7d:          row.UsersNew7d,
			New30d:         row.UsersNew30d,
			Active30d:      row.UsersActive30d,
			InstanceAdmins: row.UsersInstanceAdmins,
		},
		Clubs:           ClubStats{Total: row.ClubsTotal, Active90d: row.ClubsActive90d},
		Players:         PlayerStats{Total: row.PlayersTotal, Linked: row.PlayersLinked},
		OpenInvitations: row.InvitationsOpen,
		APITokens:       APITokenStats{Active: row.ApiTokensActive},
		Passkeys:        row.PasskeysTotal,
		Anomalies: Anomalies{
			OrphanedUsers:          row.UsersOrphaned,
			InactiveClubs:          row.ClubsInactive180d,
			APITokensWithoutExpiry: row.ApiTokensWithoutExpiry,
			APITokensUnused90d:     row.ApiTokensUnused90d,
			RegistrationSpikes:     spikes,
		},
	}, nil
}

// ListUsers returns one page of accounts. The query must be normalized
// (see ParseUserListQuery).
func (r *Repository) ListUsers(ctx context.Context, q UserListQuery) (Page[UserListItem], error) {
	search := escapeLike(q.Search)
	onlyAdmins := q.Filter == UserFilterAdmin
	onlyOrphaned := q.Filter == UserFilterOrphaned

	total, err := r.queries.CountAdminUsers(ctx, db.CountAdminUsersParams{
		Search:       search,
		OnlyAdmins:   onlyAdmins,
		OnlyOrphaned: onlyOrphaned,
	})
	if err != nil {
		return Page[UserListItem]{}, err
	}
	rows, err := r.queries.ListAdminUsers(ctx, db.ListAdminUsersParams{
		Search:       search,
		OnlyAdmins:   onlyAdmins,
		OnlyOrphaned: onlyOrphaned,
		Sort:         q.Sort + "_" + q.Order,
		PageOffset:   int32((q.Page - 1) * q.Limit),
		PageLimit:    int32(q.Limit),
	})
	if err != nil {
		return Page[UserListItem]{}, err
	}
	items := make([]UserListItem, 0, len(rows))
	for _, u := range rows {
		items = append(items, UserListItem{
			ID:              u.ID,
			Email:           u.Email,
			IsInstanceAdmin: u.IsInstanceAdmin,
			Orphaned:        u.Orphaned,
			CreatedAt:       u.CreatedAt,
			LastLoginAt:     nullTime(u.LastLoginAt),
			OwnedClubCount:  u.OwnedClubCount,
			PlayerClubCount: u.PlayerClubCount,
			PasskeyCount:    u.PasskeyCount,
			APITokenCount:   u.ApiTokenCount,
		})
	}
	return newPage(items, q.Page, q.Limit, total), nil
}

// GetUser returns an account with its clubs, player links, API tokens
// (without hash) and passkeys.
func (r *Repository) GetUser(ctx context.Context, id string) (UserDetail, error) {
	u, err := r.queries.GetAdminUser(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return UserDetail{}, ErrUserNotFound
	}
	if err != nil {
		return UserDetail{}, err
	}
	detail := UserDetail{
		ID:                u.ID,
		Email:             u.Email,
		IsInstanceAdmin:   u.IsInstanceAdmin,
		Orphaned:          u.Orphaned,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
		PasswordChangedAt: nullTime(u.PasswordChangedAt),
		LastLoginAt:       nullTime(u.LastLoginAt),
		OwnedClubs:        []OwnedClub{},
		Players:           []UserPlayerLink{},
		APITokens:         []APITokenInfo{},
		Passkeys:          []PasskeyInfo{},
	}

	clubs, err := r.queries.ListAdminUserOwnedClubs(ctx, id)
	if err != nil {
		return UserDetail{}, err
	}
	for _, c := range clubs {
		detail.OwnedClubs = append(detail.OwnedClubs, OwnedClub{ID: c.ID, Name: c.Name, CreatedAt: c.CreatedAt, PlayerCount: c.PlayerCount})
	}

	players, err := r.queries.ListAdminUserPlayers(ctx, &id)
	if err != nil {
		return UserDetail{}, err
	}
	for _, p := range players {
		detail.Players = append(detail.Players, UserPlayerLink{
			PlayerID:   p.ID,
			PlayerName: p.Name,
			Inactive:   p.Inactive,
			ClubID:     p.ClubID,
			ClubName:   p.ClubName,
			IsOwner:    p.IsOwner,
			RoleID:     p.RoleID,
			RoleName:   nullString(p.RoleName),
		})
	}

	tokens, err := r.queries.ListAdminUserAPITokens(ctx, id)
	if err != nil {
		return UserDetail{}, err
	}
	for _, t := range tokens {
		detail.APITokens = append(detail.APITokens, APITokenInfo{
			ID:          t.ID,
			Name:        t.Name,
			TokenPrefix: t.TokenPrefix,
			CreatedAt:   t.CreatedAt,
			ExpiresAt:   nullTime(t.ExpiresAt),
			LastUsedAt:  nullTime(t.LastUsedAt),
		})
	}

	passkeys, err := r.queries.ListAdminUserPasskeys(ctx, id)
	if err != nil {
		return UserDetail{}, err
	}
	for _, p := range passkeys {
		detail.Passkeys = append(detail.Passkeys, PasskeyInfo{ID: p.ID, Name: p.Name, CreatedAt: p.CreatedAt, LastUsedAt: nullTime(p.LastUsedAt)})
	}
	return detail, nil
}

// ListClubs returns one page of clubs. The query must be normalized
// (see ParseClubListQuery).
func (r *Repository) ListClubs(ctx context.Context, q ClubListQuery) (Page[ClubListItem], error) {
	search := escapeLike(q.Search)
	onlyInactive := q.Filter == ClubFilterInactive

	total, err := r.queries.CountAdminClubs(ctx, db.CountAdminClubsParams{Search: search, OnlyInactive: onlyInactive})
	if err != nil {
		return Page[ClubListItem]{}, err
	}
	rows, err := r.queries.ListAdminClubs(ctx, db.ListAdminClubsParams{
		Search:       search,
		OnlyInactive: onlyInactive,
		Sort:         q.Sort + "_" + q.Order,
		PageOffset:   int32((q.Page - 1) * q.Limit),
		PageLimit:    int32(q.Limit),
	})
	if err != nil {
		return Page[ClubListItem]{}, err
	}
	items := make([]ClubListItem, 0, len(rows))
	for _, c := range rows {
		lastActivity, err := anyTime(c.LastActivityAt)
		if err != nil {
			return Page[ClubListItem]{}, err
		}
		items = append(items, ClubListItem{
			ID:                c.ID,
			Name:              c.Name,
			CreatedAt:         c.CreatedAt,
			Owner:             UserRef{ID: c.OwnerID, Email: c.OwnerEmail},
			PlayerCount:       c.PlayerCount,
			LinkedPlayerCount: c.LinkedPlayerCount,
			GameDayCount:      c.GameDayCount,
			LastActivityAt:    lastActivity,
		})
	}
	return newPage(items, q.Page, q.Limit, total), nil
}

// GetClub returns a club with owner, players (and linked accounts), open
// invitations and roles. No balances, transactions or penalties.
func (r *Repository) GetClub(ctx context.Context, id string) (ClubDetail, error) {
	c, err := r.queries.GetAdminClub(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ClubDetail{}, ErrClubNotFound
	}
	if err != nil {
		return ClubDetail{}, err
	}
	lastActivity, err := anyTime(c.LastActivityAt)
	if err != nil {
		return ClubDetail{}, err
	}
	detail := ClubDetail{
		ID:                c.ID,
		Name:              c.Name,
		CreatedAt:         c.CreatedAt,
		Owner:             UserRef{ID: c.OwnerID, Email: c.OwnerEmail},
		PlayerCount:       c.PlayerCount,
		LinkedPlayerCount: c.LinkedPlayerCount,
		GameDayCount:      c.GameDayCount,
		TransactionCount:  c.TransactionCount,
		LastActivityAt:    lastActivity,
		Players:           []ClubPlayer{},
		OpenInvitations:   []OpenInvitation{},
		Roles:             []ClubRole{},
	}

	players, err := r.queries.ListAdminClubPlayers(ctx, id)
	if err != nil {
		return ClubDetail{}, err
	}
	for _, p := range players {
		cp := ClubPlayer{
			ID:        p.ID,
			Name:      p.Name,
			Inactive:  p.Inactive,
			CreatedAt: p.CreatedAt,
			RoleID:    p.RoleID,
			RoleName:  nullString(p.RoleName),
		}
		if p.UserID != nil {
			cp.User = &UserRef{ID: *p.UserID, Email: p.UserEmail.String}
		}
		detail.Players = append(detail.Players, cp)
	}

	invitations, err := r.queries.ListAdminClubOpenInvitations(ctx, id)
	if err != nil {
		return ClubDetail{}, err
	}
	for _, i := range invitations {
		detail.OpenInvitations = append(detail.OpenInvitations, OpenInvitation{
			ID:         i.ID,
			PlayerID:   i.PlayerID,
			PlayerName: i.PlayerName,
			Email:      i.Email,
			CreatedAt:  i.CreatedAt,
			ExpiresAt:  i.ExpiresAt,
		})
	}

	roles, err := r.queries.ListAdminClubRoles(ctx, id)
	if err != nil {
		return ClubDetail{}, err
	}
	for _, ro := range roles {
		detail.Roles = append(detail.Roles, ClubRole{ID: ro.ID, Name: ro.Name, PlayerCount: ro.PlayerCount})
	}
	return detail, nil
}

// Memberships returns every player link plus one row per club whose owner
// has no own player in it.
func (r *Repository) Memberships(ctx context.Context) ([]Membership, error) {
	rows, err := r.queries.ListAdminMemberships(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Membership, 0, len(rows))
	for _, m := range rows {
		out = append(out, Membership{
			UserEmail:  m.UserEmail,
			ClubName:   m.ClubName,
			IsOwner:    m.IsOwner,
			PlayerName: m.PlayerName,
			RoleName:   m.RoleName,
		})
	}
	return out, nil
}

// SetInstanceAdmin sets or clears the instance admin flag of the account with
// the given e-mail. It returns ErrUserNotFound when no such account exists.
func (r *Repository) SetInstanceAdmin(ctx context.Context, email string, isAdmin bool) error {
	_, err := r.queries.SetInstanceAdminByEmail(ctx, db.SetInstanceAdminByEmailParams{
		IsInstanceAdmin: isAdmin,
		Email:           strings.TrimSpace(email),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return ErrUserNotFound
	}
	return err
}

// ListInstanceAdmins returns every account with the instance admin flag.
func (r *Repository) ListInstanceAdmins(ctx context.Context) ([]InstanceAdmin, error) {
	rows, err := r.queries.ListInstanceAdmins(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]InstanceAdmin, 0, len(rows))
	for _, a := range rows {
		out = append(out, InstanceAdmin{ID: a.ID, Email: a.Email, CreatedAt: a.CreatedAt, LastLoginAt: nullTime(a.LastLoginAt)})
	}
	return out, nil
}

func newPage[T any](items []T, page, limit int, total int64) Page[T] {
	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	return Page[T]{Data: items, Page: page, Limit: limit, Total: total, TotalPages: totalPages}
}

// escapeLike escapes the ILIKE wildcards, so a search matches literally.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func nullTime(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}

func nullString(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	v := s.String
	return &v
}

// anyTime converts a computed timestamp column (sqlc cannot type GREATEST
// over sub-selects and generates interface{}) into a *time.Time.
func anyTime(v any) (*time.Time, error) {
	switch t := v.(type) {
	case nil:
		return nil, nil
	case time.Time:
		return &t, nil
	default:
		return nil, fmt.Errorf("unexpected timestamp type %T", v)
	}
}
