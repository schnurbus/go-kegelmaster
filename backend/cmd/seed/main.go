// Command seed fills a local development database with example data: two users,
// a club with roles, players, penalty types, competitions, published and draft
// game days and a few deposits. All bookings go through the regular repositories,
// so player and club balances stay consistent with the ledger.
//
// The command is meant for local test instances only. It refuses to run when the
// demo user already exists, so it is safe to call repeatedly.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/schnurbus/go-kegelmaster/backend/internal/automigrate"
	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/competition"
	"github.com/schnurbus/go-kegelmaster/backend/internal/config"
	"github.com/schnurbus/go-kegelmaster/backend/internal/database"
	"github.com/schnurbus/go-kegelmaster/backend/internal/gameday"
	"github.com/schnurbus/go-kegelmaster/backend/internal/gamedayimport"
	"github.com/schnurbus/go-kegelmaster/backend/internal/penaltytype"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
	"github.com/schnurbus/go-kegelmaster/backend/internal/transaction"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

const (
	clubName       = "KC Alle Neune"
	ownerEmail     = "demo@kegelmaster.local"
	memberEmail    = "mitglied@kegelmaster.local"
	seedPassword   = "kegelmaster"
	baseFeeCents   = 500
	gameDayCount   = 8
	gameDayEveryNd = 14
)

type repos struct {
	users        *user.Repository
	clubs        *club.Repository
	roles        *role.Repository
	players      *player.Repository
	penaltyTypes *penaltytype.Repository
	competitions *competition.Repository
	gameDays     *gameday.Repository
	transactions *transaction.Repository
}

func main() {
	databaseURL := flag.String("database-url", "", "PostgreSQL-Connection-String (Fallback: DATABASE_URL)")
	migrate := flag.Bool("migrate", true, "Ausstehende Migrationen vor dem Seeden ausführen")
	flag.Parse()

	dbURL := *databaseURL
	if dbURL == "" {
		dbURL = config.Load().DatabaseURL
	}

	ctx := context.Background()
	if err := automigrate.Run(ctx, dbURL, *migrate); err != nil {
		slog.Error("Migrationen", "error", err)
		os.Exit(1)
	}

	conn, err := database.New(dbURL)
	if err != nil {
		slog.Error("Datenbank verbinden", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	playerRepo := player.NewRepository(conn)
	clubRepo := club.NewRepository(conn)
	gameDayRepo := gameday.NewRepository(conn)
	r := repos{
		users:        user.NewRepository(conn),
		clubs:        clubRepo,
		roles:        role.NewRepository(conn),
		players:      playerRepo,
		penaltyTypes: penaltytype.NewRepository(conn),
		competitions: competition.NewRepository(conn),
		gameDays:     gameDayRepo,
		transactions: transaction.NewRepository(conn, playerRepo, clubRepo, gameDayRepo),
	}

	if _, err := r.users.GetByEmail(ctx, ownerEmail); err == nil {
		slog.Info("Beispiel-Daten bereits vorhanden, nichts zu tun", "user", ownerEmail)
		return
	} else if !errors.Is(err, user.ErrNotFound) {
		slog.Error("Demo-User prüfen", "error", err)
		os.Exit(1)
	}

	if err := seed(ctx, r); err != nil {
		slog.Error("Seed fehlgeschlagen", "error", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Beispiel-Daten angelegt:")
	fmt.Printf("  Club:        %s\n", clubName)
	fmt.Printf("  Owner:       %s / %s\n", ownerEmail, seedPassword)
	fmt.Printf("  Mitglied:    %s / %s (eingeschränkte Rechte)\n", memberEmail, seedPassword)
	fmt.Println()
}

func seed(ctx context.Context, r repos) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(seedPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Passwort hashen: %w", err)
	}
	owner, err := r.users.Create(ctx, user.CreateUserParams{Email: ownerEmail, PasswordHash: string(hash)})
	if err != nil {
		return fmt.Errorf("Owner anlegen: %w", err)
	}
	member, err := r.users.Create(ctx, user.CreateUserParams{Email: memberEmail, PasswordHash: string(hash)})
	if err != nil {
		return fmt.Errorf("Mitglied-User anlegen: %w", err)
	}

	clubEntity, err := r.clubs.Create(ctx, club.CreateClubParams{
		Name:           clubName,
		BaseFee:        baseFeeCents,
		AutoTipEnabled: false,
		UserID:         owner.ID,
	})
	if err != nil {
		return fmt.Errorf("Club anlegen: %w", err)
	}
	slog.Info("Club angelegt", "club", clubEntity.Name, "id", clubEntity.ID)

	// Roles: Vorstand may do everything, Mitglied may only read, Gast pays no base fee.
	vorstand, err := createRole(ctx, r, clubEntity.ID, "Vorstand", true, role.ValidPermissionTypes())
	if err != nil {
		return err
	}
	mitglied, err := createRole(ctx, r, clubEntity.ID, "Mitglied", true,
		[]role.PermissionType{role.PermissionTypeList, role.PermissionTypeView})
	if err != nil {
		return err
	}
	gast, err := createRole(ctx, r, clubEntity.ID, "Gast", false,
		[]role.PermissionType{role.PermissionTypeList, role.PermissionTypeView})
	if err != nil {
		return err
	}

	type seedPlayer struct {
		name   string
		gender string
		roleID string
		userID *string
	}
	seedPlayers := []seedPlayer{
		{"Klaus", "male", vorstand.ID, &owner.ID},
		{"Petra", "female", vorstand.ID, nil},
		{"Jürgen", "male", mitglied.ID, &member.ID},
		{"Sabine", "female", mitglied.ID, nil},
		{"Dieter", "male", mitglied.ID, nil},
		{"Monika", "female", mitglied.ID, nil},
		{"Uwe", "male", mitglied.ID, nil},
		{"Gabi", "female", gast.ID, nil},
	}
	playerNames := make([]string, 0, len(seedPlayers))
	playerIDs := make(map[string]string, len(seedPlayers))
	for _, sp := range seedPlayers {
		roleID, gender := sp.roleID, sp.gender
		p, err := r.players.Create(ctx, player.CreatePlayerParams{
			ClubID: clubEntity.ID,
			UserID: sp.userID,
			RoleID: roleID,
			Name:   sp.name,
			Gender: &gender,
		})
		if err != nil {
			return fmt.Errorf("Spieler %s anlegen: %w", sp.name, err)
		}
		playerNames = append(playerNames, sp.name)
		playerIDs[sp.name] = p.ID
	}

	penalties := []penaltytype.CreatePenaltyTypeParams{
		{Name: "Pudel", Description: "Kugel in der Rinne", Price: 20},
		{Name: "Kranz", Description: "Kranz geworfen (Runde schmeißen)", Price: 100},
		{Name: "Handy", Description: "Handy klingelt", Price: 50},
		{Name: "Zu spät", Description: "Nach Spielbeginn erschienen", Price: 200},
		{Name: "Bier", Description: "Getränke-Umlage pro Liter", Price: 300, AllowsDecimalQuantity: true},
	}
	for _, p := range penalties {
		p.ClubID = clubEntity.ID
		if _, err := r.penaltyTypes.Create(ctx, p); err != nil {
			return fmt.Errorf("Strafentyp %s anlegen: %w", p.Name, err)
		}
	}

	competitions := []competition.CreateCompetitionParams{
		{Name: "Pokal", ScoringType: "winner"},
		{Name: "Pumpe", ScoringType: "loser"},
		{Name: "Königin/König", ScoringType: "winner", IsGenderSpecific: true},
	}
	for _, c := range competitions {
		c.ClubID = clubEntity.ID
		if _, err := r.competitions.Create(ctx, c); err != nil {
			return fmt.Errorf("Wettbewerb %s anlegen: %w", c.Name, err)
		}
	}

	// Game days: every two weeks back from today; the most recent one stays a draft.
	today := time.Now().UTC().Truncate(24 * time.Hour)
	dates := make([]time.Time, 0, gameDayCount)
	for i := gameDayCount - 1; i >= 0; i-- {
		dates = append(dates, today.AddDate(0, 0, -i*gameDayEveryNd))
	}

	parsed := buildGameDayData(dates, playerNames)
	deps := &gamedayimport.Dependencies{
		ClubRepo:        r.clubs,
		PlayerRepo:      r.players,
		PenaltyTypeRepo: r.penaltyTypes,
		CompetitionRepo: r.competitions,
		GameDayRepo:     r.gameDays,
		TransactionRepo: r.transactions,
		RoleRepo:        r.roles,
	}
	resolved, playerByName, err := gamedayimport.Resolve(ctx, parsed, clubName, deps)
	if err != nil {
		return fmt.Errorf("Spieltage auflösen: %w", err)
	}
	// The import creates draft game days with fees but without bookings.
	if err := gamedayimport.Apply(ctx, resolved, playerByName, deps, false); err != nil {
		return fmt.Errorf("Spieltage anlegen: %w", err)
	}

	gameDays, err := r.gameDays.GetByClubID(ctx, clubEntity.ID)
	if err != nil {
		return fmt.Errorf("Spieltage laden: %w", err)
	}
	latest := dates[len(dates)-1]
	for _, gd := range gameDays {
		if gd.Date.UTC().Truncate(24 * time.Hour).Equal(latest) {
			continue
		}
		// Publishing books base fees and penalties exactly like the UI does.
		if _, err := r.transactions.SaveGameDayWithCharges(ctx, transaction.SaveGameDayWithChargesParams{
			ID:         gd.ID,
			ClubID:     clubEntity.ID,
			Date:       gd.Date,
			Notes:      gd.Notes,
			IsDraft:    false,
			ChargeFees: true,
		}); err != nil {
			return fmt.Errorf("Spieltag %s veröffentlichen: %w", gd.Date.Format("02.01.2006"), err)
		}
	}

	// Deposits and a club expense so the ledger has every booking type.
	deposits := []struct {
		name   string
		amount int
	}{
		{"Klaus", 5000}, {"Petra", 5000}, {"Jürgen", 3000}, {"Sabine", 5000},
		{"Dieter", 4000}, {"Monika", 2000},
	}
	for _, d := range deposits {
		pid := playerIDs[d.name]
		if _, err := r.transactions.Create(ctx, transaction.CreateTransactionParams{
			ClubID:          clubEntity.ID,
			PlayerID:        &pid,
			TransactionType: transaction.TransactionTypeDeposit,
			Amount:          d.amount,
			Description:     "Einzahlung bar",
			TransactionDate: dates[len(dates)/2],
		}); err != nil {
			return fmt.Errorf("Einzahlung %s: %w", d.name, err)
		}
	}
	if _, err := r.transactions.Create(ctx, transaction.CreateTransactionParams{
		ClubID:          clubEntity.ID,
		TransactionType: transaction.TransactionTypeExpense,
		Amount:          -4500,
		Description:     "Bahnmiete",
		TransactionDate: dates[len(dates)-2],
	}); err != nil {
		return fmt.Errorf("Ausgabe anlegen: %w", err)
	}

	return nil
}

func createRole(ctx context.Context, r repos, clubID, name string, paysBaseFee bool, perms []role.PermissionType) (role.Role, error) {
	created, err := r.roles.Create(ctx, role.CreateRoleParams{ClubID: clubID, Name: name, PaysBaseFee: paysBaseFee})
	if err != nil {
		return role.Role{}, fmt.Errorf("Rolle %s anlegen: %w", name, err)
	}
	for _, entity := range role.ValidEntityTypes() {
		for _, perm := range perms {
			if _, err := r.roles.AddPermission(ctx, role.AddPermissionParams{
				RoleID: created.ID, EntityType: entity, PermissionType: perm,
			}); err != nil {
				return role.Role{}, fmt.Errorf("Recht %s/%s für %s: %w", entity, perm, name, err)
			}
		}
	}
	return created, nil
}

// buildGameDayData creates deterministic pseudo-random penalties and competition
// results per game day and player, in the same shape the CSV import expects.
func buildGameDayData(dates []time.Time, players []string) *gamedayimport.ParsedCSV {
	header := []string{"Datum", "Name", "Pudel", "Kranz", "Handy", "Zu spät", "Bier", "Pokal", "Pumpe", "Königin/König"}
	parsed := &gamedayimport.ParsedCSV{Header: header}
	seed := uint32(42)
	next := func(n int) int { // small LCG, stable across runs
		seed = seed*1664525 + 1013904223
		return int(seed>>16) % n
	}
	for di, d := range dates {
		for pi, name := range players {
			// Not everybody shows up every time.
			if (di+pi)%5 == 4 {
				continue
			}
			parsed.Rows = append(parsed.Rows, gamedayimport.CSVRow{
				Date:       d,
				PlayerName: name,
				Values: map[string]float64{
					"Pudel":         float64(next(6)),
					"Kranz":         float64(next(10) / 8),
					"Handy":         float64(next(10) / 9),
					"Zu spät":       float64(next(10) / 9),
					"Bier":          float64(next(5)) * 0.5,
					"Pokal":         float64(40 + next(30)),
					"Pumpe":         float64(next(8)),
					"Königin/König": float64(20 + next(20)),
				},
			})
		}
	}
	return parsed
}
