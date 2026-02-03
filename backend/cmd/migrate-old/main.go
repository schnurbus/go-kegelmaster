package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/schnurbus/go-kegelmaster/backend/internal/database"
	"github.com/schnurbus/go-kegelmaster/backend/internal/oldmigrate"
)

func main() {
	sourceURL := flag.String("source", "", "PostgreSQL-URL der alten Datenbank (Pflicht)")
	targetURL := flag.String("target", "", "PostgreSQL-URL der neuen Datenbank (Pflicht)")
	dryRun := flag.Bool("dry-run", false, "Nur lesen, keine Schreibzugriffe auf die Ziel-DB")
	flag.Parse()

	if *sourceURL == "" || *targetURL == "" {
		flag.Usage()
		os.Exit(2)
	}

	source, err := database.New(*sourceURL)
	if err != nil {
		slog.Error("Quell-DB verbinden", "error", err)
		os.Exit(1)
	}
	defer source.Close()

	target, err := database.New(*targetURL)
	if err != nil {
		slog.Error("Ziel-DB verbinden", "error", err)
		os.Exit(1)
	}
	defer target.Close()

	ctx := context.Background()
	if err := oldmigrate.Run(ctx, source, target, *dryRun); err != nil {
		slog.Error("Migration fehlgeschlagen", "error", err)
		os.Exit(1)
	}
}
