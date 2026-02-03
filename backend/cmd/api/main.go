package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/schnurbus/go-kegelmaster/backend/internal/auth"
	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/competition"
	"github.com/schnurbus/go-kegelmaster/backend/internal/config"
	"github.com/schnurbus/go-kegelmaster/backend/internal/database"
	"github.com/schnurbus/go-kegelmaster/backend/internal/email"
	"github.com/schnurbus/go-kegelmaster/backend/internal/gameday"
	"github.com/schnurbus/go-kegelmaster/backend/internal/invitation"
	"github.com/schnurbus/go-kegelmaster/backend/internal/passwordreset"
	"github.com/schnurbus/go-kegelmaster/backend/internal/penaltytype"
	"github.com/schnurbus/go-kegelmaster/backend/internal/permission"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
	"github.com/schnurbus/go-kegelmaster/backend/internal/server"
	"github.com/schnurbus/go-kegelmaster/backend/internal/transaction"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

func main() {
	cfg := config.Load()

	if cfg.AppEnv == "production" && (cfg.JWTSecret == "" || cfg.JWTSecret == "dev-secret-change-me") {
		slog.Error("JWT_SECRET must be set in production")
		os.Exit(1)
	}

	slog.Info("starting backend service", "env", cfg.AppEnv, "port", cfg.HTTPPort)

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		slog.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

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

	srv := server.New(cfg, server.Dependencies{
		UserRepo:        userRepo,
		ClubRepo:        clubRepo,
		RoleRepo:        roleRepo,
		PlayerRepo:      playerRepo,
		PenaltyTypeRepo: penaltyTypeRepo,
		CompetitionRepo: competitionRepo,
		GameDayRepo:     gameDayRepo,
		TransactionRepo: transactionRepo,
		AuthService:     authSvc,
		PermissionCheck: permissionCheck,
		InvitationRepo:    invitationRepo,
		EmailService:      emailSvc,
		PasswordResetRepo: passwordResetRepo,
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
}
