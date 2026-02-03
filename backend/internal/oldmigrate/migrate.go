package oldmigrate

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// Maps hold old bigint ID -> new UUID for FK resolution.
type Maps struct {
	UserID                map[int64]uuid.UUID
	ClubID                 map[int64]uuid.UUID
	RoleID                 map[int64]uuid.UUID
	PlayerID               map[int64]uuid.UUID
	PenaltyTypeID          map[int64]uuid.UUID
	CompetitionID          map[int64]uuid.UUID
	GameDayID              map[int64]uuid.UUID
	GameDayParticipantID    map[string]uuid.UUID // key: "matchdayID:playerID"
	GameDayFeeID           map[int64]uuid.UUID  // old fee_entry_id -> new game_day_fee id
}

func newMaps() *Maps {
	return &Maps{
		UserID:             make(map[int64]uuid.UUID),
		ClubID:             make(map[int64]uuid.UUID),
		RoleID:             make(map[int64]uuid.UUID),
		PlayerID:           make(map[int64]uuid.UUID),
		PenaltyTypeID:      make(map[int64]uuid.UUID),
		CompetitionID:      make(map[int64]uuid.UUID),
		GameDayID:          make(map[int64]uuid.UUID),
		GameDayParticipantID: make(map[string]uuid.UUID),
		GameDayFeeID:       make(map[int64]uuid.UUID),
	}
}

func participantKey(matchdayID, playerID int64) string {
	return fmt.Sprintf("%d:%d", matchdayID, playerID)
}

// Run runs the full migration from source to target. If dryRun is true, no writes are performed.
func Run(ctx context.Context, source, target *sql.DB, dryRun bool) error {
	maps := newMaps()

	if dryRun {
		slog.Info("Dry-Run: nur Lesen von der Quell-DB, keine Schreibzugriffe")
	}

	tx := (*sql.Tx)(nil)
	if !dryRun {
		var err error
		tx, err = target.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("target begin tx: %w", err)
		}
		defer func() {
			if tx != nil {
				_ = tx.Rollback()
			}
		}()
	}

	exec := target.ExecContext
	if tx != nil {
		exec = tx.ExecContext
	}
	var queryRow func(ctx context.Context, query string, args ...any) *sql.Row
	if tx != nil {
		queryRow = tx.QueryRowContext
	}

	// 1. Users
	if err := migrateUsers(ctx, source, exec, queryRow, maps, dryRun); err != nil {
		return err
	}
	// 2. Clubs (+ club_settings for auto_tip)
	if err := migrateClubs(ctx, source, exec, maps, dryRun); err != nil {
		return err
	}
	// 3. Roles
	if err := migrateRoles(ctx, source, exec, maps, dryRun); err != nil {
		return err
	}
	// 4. Role permissions (optional mapping)
	if err := migrateRolePermissions(ctx, source, exec, maps, dryRun); err != nil {
		return err
	}
	// 5. Players
	if err := migratePlayers(ctx, source, exec, maps, dryRun); err != nil {
		return err
	}
	// 6. Penalty types (fee_types)
	if err := migratePenaltyTypes(ctx, source, exec, maps, dryRun); err != nil {
		return err
	}
	// 7. Competitions (competition_types)
	if err := migrateCompetitions(ctx, source, exec, maps, dryRun); err != nil {
		return err
	}
	// 8. Game days (matchdays)
	if err := migrateGameDays(ctx, source, exec, maps, dryRun); err != nil {
		return err
	}
	// 9. Game day participants (matchday_player)
	if err := migrateGameDayParticipants(ctx, source, exec, maps, dryRun); err != nil {
		return err
	}
	// 10. Game day fees (fee_entries + fee_type_versions + fee_types)
	if err := migrateGameDayFees(ctx, source, exec, maps, dryRun); err != nil {
		return err
	}
	// 11. Game day competition values (competition_entries)
	if err := migrateGameDayCompetitionValues(ctx, source, exec, maps, dryRun); err != nil {
		return err
	}
	// 12. Transactions
	if err := migrateTransactions(ctx, source, exec, maps, dryRun); err != nil {
		return err
	}
	// 13. Player invitations
	if err := migratePlayerInvitations(ctx, source, exec, maps, dryRun); err != nil {
		return err
	}

	if !dryRun && tx != nil {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit: %w", err)
		}
		tx = nil
	}
	slog.Info("Migration abgeschlossen", "dry_run", dryRun)
	return nil
}

type execContext func(ctx context.Context, query string, args ...any) (sql.Result, error)

func migrateUsers(ctx context.Context, source *sql.DB, exec execContext, queryRow func(context.Context, string, ...any) *sql.Row, maps *Maps, dryRun bool) error {
	rows, err := source.QueryContext(ctx, `SELECT id, email, password, created_at, updated_at FROM users`)
	if err != nil {
		return fmt.Errorf("read users: %w", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var id int64
		var email, password string
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &email, &password, &createdAt, &updatedAt); err != nil {
			return fmt.Errorf("scan user: %w", err)
		}
		if !dryRun && queryRow != nil {
			var existingID uuid.UUID
			err := queryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&existingID)
			if err == nil {
				maps.UserID[id] = existingID
				count++
				continue
			}
			if err != sql.ErrNoRows {
				return fmt.Errorf("lookup user by email %q: %w", email, err)
			}
		}
		newID := ID("users", id)
		maps.UserID[id] = newID
		if dryRun {
			count++
			continue
		}
		_, err := exec(ctx, `INSERT INTO users (id, email, password_hash, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO NOTHING`,
			newID, email, password, createdAt, updatedAt)
		if err != nil {
			return fmt.Errorf("insert user %d: %w", id, err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("users rows: %w", err)
	}
	slog.Info("Users migriert", "count", count)
	return nil
}

func migrateClubs(ctx context.Context, source *sql.DB, exec execContext, maps *Maps, dryRun bool) error {
	// Load auto_tip from club_settings: club_id -> value (1 = enabled)
	autoTip := make(map[int64]bool)
	rs, err := source.QueryContext(ctx, `SELECT club_id, value FROM club_settings WHERE name = 'auto_tip'`)
	if err != nil {
		return fmt.Errorf("read club_settings: %w", err)
	}
	for rs.Next() {
		var cid int64
		var v int
		if err := rs.Scan(&cid, &v); err != nil {
			rs.Close()
			return fmt.Errorf("scan club_setting: %w", err)
		}
		autoTip[cid] = v == 1
	}
	rs.Close()
	if err := rs.Err(); err != nil {
		return fmt.Errorf("club_settings rows: %w", err)
	}

	rows, err := source.QueryContext(ctx, `SELECT id, user_id, name, balance, base_fee, created_at, updated_at FROM clubs`)
	if err != nil {
		return fmt.Errorf("read clubs: %w", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var id, userID int64
		var name string
		var balance, baseFee int32
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &userID, &name, &balance, &baseFee, &createdAt, &updatedAt); err != nil {
			return fmt.Errorf("scan club: %w", err)
		}
		newUserID, ok := maps.UserID[userID]
		if !ok {
			slog.Warn("Club übersprungen: user_id nicht migriert", "club_id", id, "user_id", userID)
			continue
		}
		newID := ID("clubs", id)
		maps.ClubID[id] = newID
		at := true
		if v, exists := autoTip[id]; exists {
			at = v
		}
		if dryRun {
			count++
			continue
		}
		_, err := exec(ctx, `INSERT INTO clubs (id, user_id, name, balance, base_fee, auto_tip_enabled, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (id) DO NOTHING`,
			newID, newUserID, name, balance, baseFee, at, createdAt, updatedAt)
		if err != nil {
			return fmt.Errorf("insert club %d: %w", id, err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("clubs rows: %w", err)
	}
	slog.Info("Clubs migriert", "count", count)
	return nil
}

func migrateRoles(ctx context.Context, source *sql.DB, exec execContext, maps *Maps, dryRun bool) error {
	rows, err := source.QueryContext(ctx, `SELECT id, club_id, name, is_base_fee_active, created_at, updated_at FROM roles`)
	if err != nil {
		return fmt.Errorf("read roles: %w", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var id int64
		var clubID sql.NullInt64
		var name string
		var paysBaseFee bool
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &clubID, &name, &paysBaseFee, &createdAt, &updatedAt); err != nil {
			return fmt.Errorf("scan role: %w", err)
		}
		if !clubID.Valid {
			slog.Warn("Rolle übersprungen: club_id NULL", "role_id", id)
			continue
		}
		newClubID, ok := maps.ClubID[clubID.Int64]
		if !ok {
			slog.Warn("Rolle übersprungen: club nicht migriert", "role_id", id, "club_id", clubID.Int64)
			continue
		}
		newID := ID("roles", id)
		maps.RoleID[id] = newID
		if dryRun {
			count++
			continue
		}
		_, err := exec(ctx, `INSERT INTO roles (id, club_id, name, pays_base_fee, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO NOTHING`,
			newID, newClubID, name, paysBaseFee, createdAt, updatedAt)
		if err != nil {
			return fmt.Errorf("insert role %d: %w", id, err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("roles rows: %w", err)
	}
	slog.Info("Roles migriert", "count", count)
	return nil
}

// permissionNameToEntityPermission maps old permission names to (entity_type, permission_type).
// Unknown names are skipped when migrating role_permissions.
var permissionNameToEntityPermission = map[string][2]string{
	"view roles":   {"roles", "view"},
	"create roles": {"roles", "create"},
	"update roles": {"roles", "update"},
	"delete roles": {"roles", "delete"},
	"list roles":   {"roles", "list"},
	"view players": {"players", "view"},
	"create players": {"players", "create"},
	"update players": {"players", "update"},
	"delete players": {"players", "delete"},
	"list players": {"players", "list"},
	"view game_days": {"game_days", "view"},
	"create game_days": {"game_days", "create"},
	"update game_days": {"game_days", "update"},
	"delete game_days": {"game_days", "delete"},
	"list game_days": {"game_days", "list"},
	"view penalty_types": {"penalty_types", "view"},
	"create penalty_types": {"penalty_types", "create"},
	"update penalty_types": {"penalty_types", "update"},
	"delete penalty_types": {"penalty_types", "delete"},
	"list penalty_types": {"penalty_types", "list"},
	"view competitions": {"competitions", "view"},
	"create competitions": {"competitions", "create"},
	"update competitions": {"competitions", "update"},
	"delete competitions": {"competitions", "delete"},
	"list competitions": {"competitions", "list"},
	"view transactions": {"transactions", "view"},
	"create transactions": {"transactions", "create"},
	"update transactions": {"transactions", "update"},
	"delete transactions": {"transactions", "delete"},
	"list transactions": {"transactions", "list"},
}

func migrateRolePermissions(ctx context.Context, source *sql.DB, exec execContext, maps *Maps, dryRun bool) error {
	rows, err := source.QueryContext(ctx, `SELECT rp.role_id, p.name
		FROM role_has_permissions rp
		JOIN permissions p ON p.id = rp.permission_id`)
	if err != nil {
		return fmt.Errorf("read role_has_permissions: %w", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var roleID int64
		var permName string
		if err := rows.Scan(&roleID, &permName); err != nil {
			return fmt.Errorf("scan role_has_permissions: %w", err)
		}
		newRoleID, ok := maps.RoleID[roleID]
		if !ok {
			continue
		}
		ep, ok := permissionNameToEntityPermission[permName]
		if !ok {
			slog.Debug("Permission ohne Mapping übersprungen", "name", permName)
			continue
		}
		permID := uuid.NewSHA1(namespace, []byte(fmt.Sprintf("role_perm:%d:%s", roleID, permName)))
		if dryRun {
			count++
			continue
		}
		_, err := exec(ctx, `INSERT INTO role_permissions (id, role_id, entity_type, permission_type, created_at)
			VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO NOTHING`,
			permID, newRoleID, ep[0], ep[1], time.Now())
		if err != nil {
			return fmt.Errorf("insert role_permission: %w", err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("role_permissions rows: %w", err)
	}
	slog.Info("Role permissions migriert", "count", count)
	return nil
}

func migratePlayers(ctx context.Context, source *sql.DB, exec execContext, maps *Maps, dryRun bool) error {
	rows, err := source.QueryContext(ctx, `SELECT id, club_id, user_id, role_id, name, balance, initial_balance, sex, created_at, updated_at FROM players`)
	if err != nil {
		return fmt.Errorf("read players: %w", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var id, clubID, roleID int64
		var userID sql.NullInt64
		var name string
		var balance, initialBalance int32
		var sex int
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &clubID, &userID, &roleID, &name, &balance, &initialBalance, &sex, &createdAt, &updatedAt); err != nil {
			return fmt.Errorf("scan player: %w", err)
		}
		newClubID, ok := maps.ClubID[clubID]
		if !ok {
			slog.Warn("Player übersprungen: club nicht migriert", "player_id", id)
			continue
		}
		var newUserID *uuid.UUID
		if userID.Valid {
			if u, ok := maps.UserID[userID.Int64]; ok {
				newUserID = &u
			}
		}
		newRoleID, ok := maps.RoleID[roleID]
		if !ok {
			slog.Warn("Player übersprungen: role nicht migriert", "player_id", id, "role_id", roleID)
			continue
		}
		var gender sql.NullString
		switch sex {
		case 1:
			gender = sql.NullString{String: "male", Valid: true}
		case 2:
			gender = sql.NullString{String: "female", Valid: true}
		}
		newID := ID("players", id)
		maps.PlayerID[id] = newID
		if dryRun {
			count++
			continue
		}
		_, err := exec(ctx, `INSERT INTO players (id, club_id, user_id, role_id, name, balance, start_balance, gender, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT (id) DO NOTHING`,
			newID, newClubID, nullUUID(newUserID), newRoleID, name, balance, initialBalance, nullString(gender), createdAt, updatedAt)
		if err != nil {
			return fmt.Errorf("insert player %d: %w", id, err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("players rows: %w", err)
	}
	slog.Info("Players migriert", "count", count)
	return nil
}

func nullUUID(u *uuid.UUID) interface{} {
	if u == nil {
		return nil
	}
	return u.String()
}

func nullString(n sql.NullString) interface{} {
	if !n.Valid {
		return nil
	}
	return n.String
}

func nullTimeToTime(n sql.NullTime) time.Time {
	if !n.Valid {
		return time.Time{}
	}
	return n.Time
}

func migratePenaltyTypes(ctx context.Context, source *sql.DB, exec execContext, maps *Maps, dryRun bool) error {
	rows, err := source.QueryContext(ctx, `SELECT id, club_id, name, description, amount, position, created_at, updated_at FROM fee_types`)
	if err != nil {
		return fmt.Errorf("read fee_types: %w", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var id, clubID int64
		var name string
		var desc sql.NullString
		var amount, position int32
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &clubID, &name, &desc, &amount, &position, &createdAt, &updatedAt); err != nil {
			return fmt.Errorf("scan fee_type: %w", err)
		}
		newClubID, ok := maps.ClubID[clubID]
		if !ok {
			continue
		}
		newID := ID("penalty_types", id)
		maps.PenaltyTypeID[id] = newID
		if dryRun {
			count++
			continue
		}
		_, err := exec(ctx, `INSERT INTO penalty_types (id, club_id, name, description, price, display_order, allows_decimal_quantity, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, false, $7, $8) ON CONFLICT (id) DO NOTHING`,
			newID, newClubID, name, nullString(desc), amount, position, createdAt, updatedAt)
		if err != nil {
			return fmt.Errorf("insert penalty_type %d: %w", id, err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("fee_types rows: %w", err)
	}
	slog.Info("Penalty types migriert", "count", count)
	return nil
}

// scoring_type: 0→winner, 1→loser, 2→both
func migrateCompetitions(ctx context.Context, source *sql.DB, exec execContext, maps *Maps, dryRun bool) error {
	rows, err := source.QueryContext(ctx, `SELECT id, club_id, name, type, is_sex_specific, position, created_at, updated_at FROM competition_types`)
	if err != nil {
		return fmt.Errorf("read competition_types: %w", err)
	}
	defer rows.Close()

	scoringMap := map[int16]string{0: "winner", 1: "loser", 2: "both"}
	var count int
	for rows.Next() {
		var id, clubID int64
		var name string
		var typ int16
		var isSexSpecific bool
		var position int32
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &clubID, &name, &typ, &isSexSpecific, &position, &createdAt, &updatedAt); err != nil {
			return fmt.Errorf("scan competition_type: %w", err)
		}
		newClubID, ok := maps.ClubID[clubID]
		if !ok {
			continue
		}
		scoringType := "winner"
		if s, ok := scoringMap[typ]; ok {
			scoringType = s
		}
		newID := ID("competitions", id)
		maps.CompetitionID[id] = newID
		if dryRun {
			count++
			continue
		}
		_, err := exec(ctx, `INSERT INTO competitions (id, club_id, name, scoring_type, is_gender_specific, display_order, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (id) DO NOTHING`,
			newID, newClubID, name, scoringType, isSexSpecific, position, createdAt, updatedAt)
		if err != nil {
			return fmt.Errorf("insert competition %d: %w", id, err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("competition_types rows: %w", err)
	}
	slog.Info("Competitions migriert", "count", count)
	return nil
}

func migrateGameDays(ctx context.Context, source *sql.DB, exec execContext, maps *Maps, dryRun bool) error {
	rows, err := source.QueryContext(ctx, `SELECT id, club_id, date, notes, created_at, updated_at FROM matchdays`)
	if err != nil {
		return fmt.Errorf("read matchdays: %w", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var id, clubID int64
		var date time.Time
		var notes sql.NullString
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &clubID, &date, &notes, &createdAt, &updatedAt); err != nil {
			return fmt.Errorf("scan matchday: %w", err)
		}
		newClubID, ok := maps.ClubID[clubID]
		if !ok {
			continue
		}
		newID := ID("game_days", id)
		maps.GameDayID[id] = newID
		if dryRun {
			count++
			continue
		}
		_, err := exec(ctx, `INSERT INTO game_days (id, club_id, date, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO NOTHING`,
			newID, newClubID, date, nullString(notes), createdAt, updatedAt)
		if err != nil {
			return fmt.Errorf("insert game_day %d: %w", id, err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("matchdays rows: %w", err)
	}
	slog.Info("Game days migriert", "count", count)
	return nil
}

func migrateGameDayParticipants(ctx context.Context, source *sql.DB, exec execContext, maps *Maps, dryRun bool) error {
	rows, err := source.QueryContext(ctx, `SELECT matchday_id, player_id, created_at FROM matchday_player`)
	if err != nil {
		return fmt.Errorf("read matchday_player: %w", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var matchdayID, playerID int64
		var createdAt time.Time
		if err := rows.Scan(&matchdayID, &playerID, &createdAt); err != nil {
			return fmt.Errorf("scan matchday_player: %w", err)
		}
		newGameDayID, ok1 := maps.GameDayID[matchdayID]
		newPlayerID, ok2 := maps.PlayerID[playerID]
		if !ok1 || !ok2 {
			continue
		}
		key := participantKey(matchdayID, playerID)
		newID := IDFromComposite("gdp", matchdayID, playerID)
		maps.GameDayParticipantID[key] = newID
		if dryRun {
			count++
			continue
		}
		_, err := exec(ctx, `INSERT INTO game_day_participants (id, game_day_id, player_id, created_at)
			VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING`,
			newID, newGameDayID, newPlayerID, createdAt)
		if err != nil {
			return fmt.Errorf("insert game_day_participant: %w", err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("matchday_player rows: %w", err)
	}
	slog.Info("Game day participants migriert", "count", count)
	return nil
}

func migrateGameDayFees(ctx context.Context, source *sql.DB, exec execContext, maps *Maps, dryRun bool) error {
	// fee_entries: id, matchday_id, player_id, fee_type_version_id, amount (double), created_at, updated_at
	// fee_type_versions: id, fee_type_id, name, description, amount, created_at, updated_at
	// We need: game_day_participant_id (from matchday+player), penalty_type_id (from fee_type via fee_type_version), snapshot fields, count, quantity_scale
	query := `
		SELECT fe.id, fe.matchday_id, fe.player_id, fe.fee_type_version_id, fe.amount, fe.created_at, fe.updated_at,
			ftv.fee_type_id, ftv.name AS ftv_name, ftv.description AS ftv_desc, ftv.amount AS ftv_amount
		FROM fee_entries fe
		JOIN fee_type_versions ftv ON ftv.id = fe.fee_type_version_id
	`
	rows, err := source.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("read fee_entries: %w", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var feID, matchdayID, playerID, feeTypeVersionID, feeTypeID int64
		var amount float64
		var feCreatedAt, feUpdatedAt time.Time
		var ftvName string
		var ftvDescNull sql.NullString
		var ftvAmount int32
		if err := rows.Scan(&feID, &matchdayID, &playerID, &feeTypeVersionID, &amount, &feCreatedAt, &feUpdatedAt,
			&feeTypeID, &ftvName, &ftvDescNull, &ftvAmount); err != nil {
			return fmt.Errorf("scan fee_entry: %w", err)
		}
		ftvDesc := ""
		if ftvDescNull.Valid {
			ftvDesc = ftvDescNull.String
		}
		participantKey := participantKey(matchdayID, playerID)
		participantID, ok := maps.GameDayParticipantID[participantKey]
		if !ok {
			slog.Warn("Game day fee übersprungen: participant nicht gefunden", "fee_entry_id", feID)
			continue
		}
		penaltyTypeID, ok := maps.PenaltyTypeID[feeTypeID]
		if !ok {
			slog.Warn("Game day fee übersprungen: penalty_type nicht migriert", "fee_entry_id", feID, "fee_type_id", feeTypeID)
			continue
		}
		// count: integer part; quantity_scale: 1 for integer, 100 for 2 decimals if amount has fractional part
		countVal := int32(amount)
		scale := int32(1)
		if amount != float64(countVal) {
			scale = 100
			countVal = int32(amount * 100)
		}
		if countVal < 1 {
			countVal = 1
		}
		newID := ID("game_day_fees", feID)
		maps.GameDayFeeID[feID] = newID
		if dryRun {
			count++
			continue
		}
		_, err := exec(ctx, `INSERT INTO game_day_fees (id, game_day_participant_id, penalty_type_id, penalty_type_name, penalty_type_description, penalty_type_price, count, quantity_scale, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT (id) DO NOTHING`,
			newID, participantID, penaltyTypeID, ftvName, ftvDesc, ftvAmount, countVal, scale, feCreatedAt, feUpdatedAt)
		if err != nil {
			return fmt.Errorf("insert game_day_fee %d: %w", feID, err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("fee_entries rows: %w", err)
	}
	slog.Info("Game day fees migriert", "count", count)
	return nil
}

func migrateGameDayCompetitionValues(ctx context.Context, source *sql.DB, exec execContext, maps *Maps, dryRun bool) error {
	rows, err := source.QueryContext(ctx, `SELECT id, matchday_id, player_id, competition_type_id, amount, created_at, updated_at FROM competition_entries`)
	if err != nil {
		return fmt.Errorf("read competition_entries: %w", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var id, matchdayID, playerID, competitionTypeID int64
		var amount int32
		var createdAtNull, updatedAtNull sql.NullTime
		if err := rows.Scan(&id, &matchdayID, &playerID, &competitionTypeID, &amount, &createdAtNull, &updatedAtNull); err != nil {
			return fmt.Errorf("scan competition_entry: %w", err)
		}
		createdAt, updatedAt := nullTimeToTime(createdAtNull), nullTimeToTime(updatedAtNull)
		key := participantKey(matchdayID, playerID)
		participantID, ok1 := maps.GameDayParticipantID[key]
		competitionID, ok2 := maps.CompetitionID[competitionTypeID]
		if !ok1 || !ok2 {
			continue
		}
		newID := ID("game_day_competition_values", id)
		if dryRun {
			count++
			continue
		}
		_, err := exec(ctx, `INSERT INTO game_day_competition_values (id, game_day_participant_id, competition_id, value, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO NOTHING`,
			newID, participantID, competitionID, amount, createdAt, updatedAt)
		if err != nil {
			return fmt.Errorf("insert game_day_competition_value %d: %w", id, err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("competition_entries rows: %w", err)
	}
	slog.Info("Game day competition values migriert", "count", count)
	return nil
}

// transaction type: 1=base_fee, 2=fee, 3=deposit, 4=tip, 5=expense
var transactionTypeMap = map[int]string{
	1: "base_fee", 2: "fee", 3: "deposit", 4: "tip", 5: "expense",
}

func migrateTransactions(ctx context.Context, source *sql.DB, exec execContext, maps *Maps, dryRun bool) error {
	rows, err := source.QueryContext(ctx, `SELECT id, club_id, player_id, matchday_id, fee_entry_id, type, amount, date, notes, created_at, updated_at FROM transactions`)
	if err != nil {
		return fmt.Errorf("read transactions: %w", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var id, clubID int64
		var playerID, matchdayID, feeEntryID sql.NullInt64
		var typ, amount int32
		var date time.Time
		var notes sql.NullString
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &clubID, &playerID, &matchdayID, &feeEntryID, &typ, &amount, &date, &notes, &createdAt, &updatedAt); err != nil {
			return fmt.Errorf("scan transaction: %w", err)
		}
		newClubID, ok := maps.ClubID[clubID]
		if !ok {
			continue
		}
		txType, ok := transactionTypeMap[int(typ)]
		if !ok {
			txType = "deposit"
		}
		var newPlayerID, newGameDayID, newGameDayFeeID interface{}
		if playerID.Valid {
			if u, ok := maps.PlayerID[playerID.Int64]; ok {
				newPlayerID = u.String()
			}
		}
		if matchdayID.Valid {
			if u, ok := maps.GameDayID[matchdayID.Int64]; ok {
				newGameDayID = u.String()
			}
		}
		if feeEntryID.Valid {
			if u, ok := maps.GameDayFeeID[feeEntryID.Int64]; ok {
				newGameDayFeeID = u.String()
			}
		}
		newID := ID("transactions", id)
		// Balance snapshots: plan says NULL; new schema has club_balance_before/after NOT NULL -> use 0
		clubBalBefore, clubBalAfter := int32(0), int32(0)
		if dryRun {
			count++
			continue
		}
		_, err := exec(ctx, `INSERT INTO transactions (id, club_id, player_id, transaction_type, amount, description, game_day_fee_id, game_day_id, player_balance_before, player_balance_after, club_balance_before, club_balance_after, transaction_date, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULL, NULL, $9, $10, $11, $12, $13) ON CONFLICT (id) DO NOTHING`,
			newID, newClubID, newPlayerID, txType, amount, nullString(notes), newGameDayFeeID, newGameDayID, clubBalBefore, clubBalAfter, date, createdAt, updatedAt)
		if err != nil {
			return fmt.Errorf("insert transaction %d: %w", id, err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("transactions rows: %w", err)
	}
	slog.Info("Transactions migriert", "count", count)
	return nil
}

func migratePlayerInvitations(ctx context.Context, source *sql.DB, exec execContext, maps *Maps, dryRun bool) error {
	rows, err := source.QueryContext(ctx, `SELECT id, player_id, email, token, expires_at, created_at, updated_at FROM player_invitations`)
	if err != nil {
		return fmt.Errorf("read player_invitations: %w", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var id, playerID int64
		var email, token string
		var expiresAt, createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &playerID, &email, &token, &expiresAt, &createdAt, &updatedAt); err != nil {
			return fmt.Errorf("scan player_invitation: %w", err)
		}
		newPlayerID, ok := maps.PlayerID[playerID]
		if !ok {
			continue
		}
		newID := ID("player_invitations", id)
		if dryRun {
			count++
			continue
		}
		_, err := exec(ctx, `INSERT INTO player_invitations (id, player_id, email, token, expires_at, accepted_at, created_at)
			VALUES ($1, $2, $3, $4, $5, NULL, $6) ON CONFLICT (id) DO NOTHING`,
			newID, newPlayerID, email, token, expiresAt, createdAt)
		if err != nil {
			return fmt.Errorf("insert player_invitation %d: %w", id, err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("player_invitations rows: %w", err)
	}
	slog.Info("Player invitations migriert", "count", count)
	return nil
}
