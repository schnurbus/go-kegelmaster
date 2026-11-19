package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/schnurbus/go-kegelmaster/backend/internal/admin"
	"github.com/schnurbus/go-kegelmaster/backend/internal/apitoken"
	"github.com/schnurbus/go-kegelmaster/backend/internal/auth"
	"github.com/schnurbus/go-kegelmaster/backend/internal/automigrate"
	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/competition"
	"github.com/schnurbus/go-kegelmaster/backend/internal/config"
	"github.com/schnurbus/go-kegelmaster/backend/internal/database"
	"github.com/schnurbus/go-kegelmaster/backend/internal/email"
	"github.com/schnurbus/go-kegelmaster/backend/internal/gameday"
	"github.com/schnurbus/go-kegelmaster/backend/internal/invitation"
	"github.com/schnurbus/go-kegelmaster/backend/internal/notification"
	"github.com/schnurbus/go-kegelmaster/backend/internal/passkey"
	"github.com/schnurbus/go-kegelmaster/backend/internal/passwordreset"
	"github.com/schnurbus/go-kegelmaster/backend/internal/penaltytype"
	"github.com/schnurbus/go-kegelmaster/backend/internal/permission"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
	"github.com/schnurbus/go-kegelmaster/backend/internal/server"
	"github.com/schnurbus/go-kegelmaster/backend/internal/statistics"
	"github.com/schnurbus/go-kegelmaster/backend/internal/transaction"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

// Version is set at build time via ldflags (e.g. -ldflags "-X main.Version=v1.2.3").
var Version string

func main() {
	// "kegelmaster admin ..." manages instance admins (also via kubectl exec).
	if len(os.Args) > 1 && os.Args[1] == "admin" {
		os.Exit(runAdminCLI(os.Args[2:]))
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg := config.Load()
	if Version != "" {
		cfg.Version = Version
	}
	if cfg.Version == "" {
		cfg.Version = "0.0.0-dev"
	}

	if cfg.AppEnv == "production" && (cfg.JWTSecret == "" || cfg.JWTSecret == "dev-secret-change-me") {
		slog.Error("JWT_SECRET must be set in production")
		os.Exit(1)
	}

	if err := cfg.ValidateRegistrationMode(); err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	slog.Info("starting backend service", "env", cfg.AppEnv, "port", cfg.HTTPPort, "registration_mode", cfg.EffectiveRegistrationMode())

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		slog.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("auto_migrate", "enabled", cfg.AutoMigrate)
	if err := automigrate.Run(context.Background(), cfg.DatabaseURL, cfg.AutoMigrate); err != nil {
		slog.Error("auto migration failed", "error", err)
		os.Exit(1)
	}

	if len(cfg.InstanceAdminEmails) > 0 {
		bootstrapCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		admin.BootstrapFromEnv(bootstrapCtx, admin.NewRepository(db), cfg.InstanceAdminEmails)
		cancel()
	}

	userRepo := user.NewRepository(db)
	clubRepo := club.NewRepository(db)
	roleRepo := role.NewRepository(db)
	playerRepo := player.NewRepository(db)
	penaltyTypeRepo := penaltytype.NewRepository(db)
	competitionRepo := competition.NewRepository(db)
	gameDayRepo := gameday.NewRepository(db)
	transactionRepo := transaction.NewRepository(db, playerRepo, clubRepo, gameDayRepo)
	authSvc := auth.NewService(cfg.JWTSecret, cfg.TokenTTLMin, cfg.RememberMeDays)
	permissionCheck := permission.NewChecker(clubRepo, roleRepo, playerRepo)
	invitationRepo := invitation.NewRepository(db, playerRepo)
	emailSvc := email.NewService(cfg.ResendAPIKey, cfg.ResendFromEmail)
	passwordResetRepo := passwordreset.NewRepository(db)
	apiTokenRepo := apitoken.NewRepository(db)
	statisticsRepo := statistics.NewRepository(db)
	gameDayNotifier := notification.NewNotifier(notification.NewRepository(db), emailSvc, cfg.BaseURL)

	// Passkeys are optional: a bad relying-party configuration disables them
	// (endpoints answer 503) instead of keeping the server from starting.
	var passkeySvc *passkey.Service
	if rpID, origins, err := cfg.PasskeyRelyingParty(); err != nil {
		slog.Warn("passkeys disabled", "error", err)
	} else if svc, err := passkey.NewService(db, passkey.Config{RPID: rpID, DisplayName: "Kegelmaster", Origins: origins}); err != nil {
		slog.Warn("passkeys disabled", "error", err)
	} else {
		passkeySvc = svc
		slog.Info("passkeys enabled", "rp_id", rpID, "origins", origins)
	}

	srv := server.New(cfg, server.Dependencies{
		UserRepo:          userRepo,
		ClubRepo:          clubRepo,
		RoleRepo:          roleRepo,
		PlayerRepo:        playerRepo,
		PenaltyTypeRepo:   penaltyTypeRepo,
		CompetitionRepo:   competitionRepo,
		GameDayRepo:       gameDayRepo,
		TransactionRepo:   transactionRepo,
		AuthService:       authSvc,
		PermissionCheck:   permissionCheck,
		InvitationRepo:    invitationRepo,
		EmailService:      emailSvc,
		PasswordResetRepo: passwordResetRepo,
		APITokenRepo:      apiTokenRepo,
		StatisticsRepo:    statisticsRepo,
		PasskeyService:    passkeySvc,
		GameDayNotifier:   gameDayNotifier,
		AdminRepo:         admin.NewRepository(db),
	})

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Listen(); err != nil {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		slog.Info("shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		if err != nil {
			slog.Error("server error", "error", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	} else {
		slog.Info("server stopped gracefully")
	}
	if !gameDayNotifier.Wait(ctx) {
		slog.Warn("pending game day notifications were not finished before shutdown")
	}
}

// runAdminCLI runs "kegelmaster admin [--database-url URL] grant|revoke|list".
// It never runs migrations; the server does that on start (AUTO_MIGRATE).
func runAdminCLI(args []string) int {
	// Log lines (audit) go to stderr, so stdout stays readable.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	fs := flag.NewFlagSet("admin", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { fmt.Fprint(os.Stderr, admin.CLIUsage) }
	databaseURL := fs.String("database-url", "", "PostgreSQL-URL (Standard: DATABASE_URL)")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if *databaseURL == "" {
		*databaseURL = config.Load().DatabaseURL
	}

	db, err := database.New(*databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Datenbankverbindung fehlgeschlagen: %v\n", err)
		return 1
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return admin.RunCLI(ctx, admin.NewRepository(db), fs.Args(), os.Stdout, os.Stderr)
}
