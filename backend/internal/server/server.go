package server

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"

	"github.com/schnurbus/go-kegelmaster/backend/internal/auth"
	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/config"
	"github.com/schnurbus/go-kegelmaster/backend/internal/gameday"
	"github.com/schnurbus/go-kegelmaster/backend/internal/handlers"
	"github.com/schnurbus/go-kegelmaster/backend/internal/penaltytype"
	"github.com/schnurbus/go-kegelmaster/backend/internal/permission"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

// Dependencies aggregates components the HTTP server relies on.
type Dependencies struct {
	UserRepo        *user.Repository
	ClubRepo        *club.Repository
	RoleRepo        *role.Repository
	PlayerRepo      *player.Repository
	PenaltyTypeRepo *penaltytype.Repository
	GameDayRepo     *gameday.Repository
	AuthService     *auth.Service
	PermissionCheck *permission.Checker
}

// Server wraps the Fiber application and its dependencies.
type Server struct {
	cfg      config.Config
	app      *fiber.App
	handlers *handlers.Handler
}

// New configures the HTTP server with default middleware and routes.
func New(cfg config.Config, deps Dependencies) *Server {
	if deps.UserRepo == nil {
		panic("user repository dependency is required")
	}
	if deps.ClubRepo == nil {
		panic("club repository dependency is required")
	}
	if deps.AuthService == nil {
		panic("auth service dependency is required")
	}
	if deps.RoleRepo == nil {
		panic("role repository dependency is required")
	}
	if deps.PlayerRepo == nil {
		panic("player repository dependency is required")
	}
	if deps.PermissionCheck == nil {
		panic("permission checker dependency is required")
	}
	if deps.GameDayRepo == nil {
		panic("game day repository dependency is required")
	}

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     splitAndTrim(cfg.CORSOrigins),
		AllowHeaders:     []string{"Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	}))

	handler := handlers.NewHandler(cfg, handlers.Dependencies{
		UserRepo:        deps.UserRepo,
		ClubRepo:        deps.ClubRepo,
		RoleRepo:        deps.RoleRepo,
		PlayerRepo:      deps.PlayerRepo,
		PenaltyTypeRepo: deps.PenaltyTypeRepo,
		GameDayRepo:     deps.GameDayRepo,
		AuthService:     deps.AuthService,
		PermissionCheck: deps.PermissionCheck,
	})

	server := &Server{
		cfg:      cfg,
		app:      app,
		handlers: handler,
	}

	server.registerRoutes()
	return server
}

func (s *Server) registerRoutes() {
	s.app.Get("/healthz", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": "0.1.0",
		})
	})

	api := s.app.Group("/api")
	api.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "backend placeholder",
		})
	})

	authGroup := api.Group("/auth")
	authGroup.Get("/csrf-token", s.handlers.HandleCSRFCookie)
	authGroup.Get("/me", s.handlers.HandleCurrentUser)
	authGroup.Post("/register", s.handlers.HandleRegister)
	authGroup.Post("/login", s.handlers.HandleLogin)
	authGroup.Post("/logout", s.handlers.HandleLogout)

	clubsGroup := api.Group("/clubs")
	clubsGroup.Get("/", s.handlers.HandleGetClubs)
	clubsGroup.Get("/:id", s.handlers.HandleGetClub)
	clubsGroup.Post("/", s.handlers.HandleCreateClub)
	clubsGroup.Put("/:id", s.handlers.HandleUpdateClub)
	clubsGroup.Delete("/:id", s.handlers.HandleDeleteClub)

	// Role endpoints
	clubsGroup.Get("/:clubId/roles", s.handlers.HandleGetRoles)
	clubsGroup.Get("/:clubId/roles/:id", s.handlers.HandleGetRole)
	clubsGroup.Post("/:clubId/roles", s.handlers.HandleCreateRole)
	clubsGroup.Put("/:clubId/roles/:id", s.handlers.HandleUpdateRole)
	clubsGroup.Delete("/:clubId/roles/:id", s.handlers.HandleDeleteRole)
	clubsGroup.Post("/:clubId/roles/:id/permissions", s.handlers.HandleAddPermission)
	clubsGroup.Delete("/:clubId/roles/:id/permissions", s.handlers.HandleRemovePermission)

	// Player endpoints
	clubsGroup.Get("/:clubId/players", s.handlers.HandleGetPlayers)
	clubsGroup.Get("/:clubId/players/:id", s.handlers.HandleGetPlayer)
	clubsGroup.Post("/:clubId/players", s.handlers.HandleCreatePlayer)
	clubsGroup.Put("/:clubId/players/:id", s.handlers.HandleUpdatePlayer)
	clubsGroup.Delete("/:clubId/players/:id", s.handlers.HandleDeletePlayer)

	// Penalty type endpoints
	clubsGroup.Get("/:clubId/penalty-types", s.handlers.HandleGetPenaltyTypes)
	clubsGroup.Get("/:clubId/penalty-types/:id", s.handlers.HandleGetPenaltyType)
	clubsGroup.Post("/:clubId/penalty-types", s.handlers.HandleCreatePenaltyType)
	clubsGroup.Put("/:clubId/penalty-types/:id", s.handlers.HandleUpdatePenaltyType)
	clubsGroup.Put("/:clubId/penalty-types/:id/display-order", s.handlers.HandleUpdatePenaltyTypeDisplayOrder)
	clubsGroup.Delete("/:clubId/penalty-types/:id", s.handlers.HandleDeletePenaltyType)

	// Game Day endpoints
	clubsGroup.Get("/:clubId/gamedays", s.handlers.HandleGetGameDays)
	clubsGroup.Get("/:clubId/gamedays/:id", s.handlers.HandleGetGameDay)
	clubsGroup.Post("/:clubId/gamedays", s.handlers.HandleCreateGameDay)
	clubsGroup.Put("/:clubId/gamedays/:id", s.handlers.HandleUpdateGameDay)
	clubsGroup.Delete("/:clubId/gamedays/:id", s.handlers.HandleDeleteGameDay)

	// Participant endpoints
	clubsGroup.Post("/:clubId/gamedays/:id/participants", s.handlers.HandleAddParticipant)
	clubsGroup.Delete("/:clubId/gamedays/:id/participants/:playerId", s.handlers.HandleRemoveParticipant)

	// Fee endpoints
	clubsGroup.Put("/:clubId/gamedays/:id/participants/:playerId/fees", s.handlers.HandleUpdateFees)
}

// Listen starts the HTTP server and blocks until it exits.
func (s *Server) Listen() error {
	addr := fmt.Sprintf(":%s", s.cfg.HTTPPort)
	slog.Info("fiber server listening", "addr", addr)
	return s.app.Listen(addr)
}

// Shutdown tries to gracefully stop the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("shutting down fiber server")
	return s.app.ShutdownWithContext(ctx)
}

func splitAndTrim(input string) []string {
	values := strings.Split(input, ",")
	result := make([]string, 0, len(values))
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return []string{"*"}
	}
	return result
}
