package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

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
)

func main() {
	clubName := flag.String("club", "", "Club-Name (exakter Match, Pflicht)")
	filePath := flag.String("file", "", "Pfad zur CSV-Datei (Pflicht)")
	databaseURL := flag.String("database-url", "", "PostgreSQL-Connection-String (Fallback: DATABASE_URL)")
	dryRun := flag.Bool("dry-run", false, "Keine DB-Änderungen; nur geplante Aktionen ausgeben")
	flag.Parse()

	if *clubName == "" || *filePath == "" {
		flag.Usage()
		os.Exit(2)
	}

	dbURL := *databaseURL
	if dbURL == "" {
		cfg := config.Load()
		dbURL = cfg.DatabaseURL
	}
	if dbURL == "" {
		slog.Error("DATABASE_URL oder --database-url erforderlich")
		os.Exit(1)
	}

	db, err := database.New(dbURL)
	if err != nil {
		slog.Error("Datenbank verbinden", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	clubRepo := club.NewRepository(db)
	playerRepo := player.NewRepository(db)
	penaltyTypeRepo := penaltytype.NewRepository(db)
	competitionRepo := competition.NewRepository(db)
	gameDayRepo := gameday.NewRepository(db)
	transactionRepo := transaction.NewRepository(db, playerRepo, clubRepo, gameDayRepo)
	roleRepo := role.NewRepository(db)

	deps := &gamedayimport.Dependencies{
		ClubRepo:         clubRepo,
		PlayerRepo:       playerRepo,
		PenaltyTypeRepo:  penaltyTypeRepo,
		CompetitionRepo:  competitionRepo,
		GameDayRepo:      gameDayRepo,
		TransactionRepo:  transactionRepo,
		RoleRepo:         roleRepo,
	}

	parsed, err := gamedayimport.ParseFile(*filePath)
	if err != nil {
		slog.Error("CSV parsen", "file", *filePath, "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	resolved, playerByName, err := gamedayimport.Resolve(ctx, parsed, *clubName, deps)
	if err != nil {
		switch {
		case err == gamedayimport.ErrClubNotFound:
			slog.Error("Club nicht gefunden", "club", *clubName)
		case err == gamedayimport.ErrClubAmbiguous:
			slog.Error("Club-Name mehrdeutig", "club", *clubName)
		case err == gamedayimport.ErrUnknownPlayer:
			slog.Error("Unbekannter Spieler in CSV", "error", err)
		case err == gamedayimport.ErrUnknownColumn:
			slog.Error("Unbekannte Spalte in CSV", "error", err)
		case err == gamedayimport.ErrAmbiguousColumn:
			slog.Error("Spaltenname mehrdeutig (Strafentyp und Wettbewerb)", "error", err)
		default:
			slog.Error("Auflösung fehlgeschlagen", "error", err)
		}
		os.Exit(1)
	}

	if err := gamedayimport.Apply(ctx, resolved, playerByName, deps, *dryRun); err != nil {
		slog.Error("Import anwenden", "error", err)
		os.Exit(1)
	}

	slog.Info("Import abgeschlossen", "club", resolved.ClubName, "dry-run", *dryRun)
}
