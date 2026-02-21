package automigrate

import (
	"context"
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/lib/pq"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/schnurbus/go-kegelmaster/backend/migrations"
)

// Run runs pending database migrations if enabled. When enabled is false, it logs
// migration_run with migrated=false and returns nil. When enabled is true, it
// runs migrate.Up() using embedded migrations and logs whether any migration was applied.
func Run(ctx context.Context, databaseURL string, enabled bool) error {
	if !enabled {
		slog.Info("migration_run", "migrated", false)
		return nil
	}

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("migration_run", "migrated", false)
			return nil
		}
		return err
	}

	slog.Info("migration_run", "migrated", true)
	return nil
}
