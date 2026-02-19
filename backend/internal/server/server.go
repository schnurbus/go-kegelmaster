package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"

	"github.com/schnurbus/go-kegelmaster/backend/internal/auth"
	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/competition"
	"github.com/schnurbus/go-kegelmaster/backend/internal/config"
	"github.com/schnurbus/go-kegelmaster/backend/internal/email"
	"github.com/schnurbus/go-kegelmaster/backend/internal/gameday"
	"github.com/schnurbus/go-kegelmaster/backend/internal/handlers"
	"github.com/schnurbus/go-kegelmaster/backend/internal/invitation"
	"github.com/schnurbus/go-kegelmaster/backend/internal/passwordreset"
	"github.com/schnurbus/go-kegelmaster/backend/internal/penaltytype"
	"github.com/schnurbus/go-kegelmaster/backend/internal/permission"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
	"github.com/schnurbus/go-kegelmaster/backend/internal/transaction"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
	"github.com/schnurbus/go-kegelmaster/backend/openapi"
)

// Dependencies aggregates components the HTTP server relies on.
type Dependencies struct {
	UserRepo        *user.Repository
	ClubRepo        *club.Repository
	RoleRepo        *role.Repository
	PlayerRepo      *player.Repository
	PenaltyTypeRepo *penaltytype.Repository
	CompetitionRepo *competition.Repository
	GameDayRepo     *gameday.Repository
	TransactionRepo *transaction.Repository
	AuthService     *auth.Service
	PermissionCheck *permission.Checker
	InvitationRepo    *invitation.Repository
	EmailService      *email.Service
	PasswordResetRepo *passwordreset.Repository
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

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"message": err.Error()})
		},
	})
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
		CompetitionRepo: deps.CompetitionRepo,
		GameDayRepo:     deps.GameDayRepo,
		TransactionRepo: deps.TransactionRepo,
		AuthService:     deps.AuthService,
		PermissionCheck: deps.PermissionCheck,
		InvitationRepo:    deps.InvitationRepo,
		EmailService:      deps.EmailService,
		PasswordResetRepo: deps.PasswordResetRepo,
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
	chatLimiter := limiter.New(limiter.Config{
		Max:        s.cfg.ChatRateLimitMax,
		Expiration: time.Duration(s.cfg.ChatRateLimitWindowMin) * time.Minute,
		KeyGenerator: func(c fiber.Ctx) string {
			token := strings.TrimSpace(c.Cookies("auth_token"))
			if token == "" {
				return c.IP()
			}
			claims, err := s.handlers.AuthSvc.ParseToken(token)
			if err != nil {
				return c.IP()
			}
			return claims.UserID
		},
	})
	api.Post("/chat", chatLimiter, s.handlers.HandleChat)

	authGroup := api.Group("/auth")
	authGroup.Get("/csrf-token", s.handlers.HandleCSRFCookie)
	authGroup.Get("/me", s.handlers.HandleCurrentUser)
	authGroup.Post("/register", s.handlers.HandleRegister)
	authGroup.Post("/login", s.handlers.HandleLogin)
	authGroup.Post("/logout", s.handlers.HandleLogout)
	authGroup.Post("/forgot-password", limiter.New(limiter.Config{
		Max:        5,
		Expiration: 15 * time.Minute,
	}), s.handlers.HandleForgotPassword)
	authGroup.Post("/reset-password", s.handlers.HandleResetPassword)

	clubsGroup := api.Group("/clubs")
	clubsGroup.Get("/", s.handlers.HandleGetClubs)
	clubsGroup.Get("/:id", s.handlers.HandleGetClub)
	clubsGroup.Post("/", s.handlers.HandleCreateClub)
	clubsGroup.Put("/:id", s.handlers.HandleUpdateClub)
	clubsGroup.Post("/:id/recalculate-balance", s.handlers.HandleRecalculateClubBalance)
	clubsGroup.Post("/:id/transfer-owner", s.handlers.HandleTransferClubOwner)
	clubsGroup.Delete("/:id", s.handlers.HandleDeleteClub)

	// My permissions for a club (must be before :id to avoid matching)
	clubsGroup.Get("/:clubId/permissions/me", s.handlers.HandleGetMyPermissions)

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
	clubsGroup.Get("/:clubId/players/me", s.handlers.HandleGetMyPlayer)
	clubsGroup.Get("/:clubId/players/me/penalty-history", s.handlers.HandleGetMyPenaltyHistory)
	clubsGroup.Get("/:clubId/players/me/competition-history", s.handlers.HandleGetMyCompetitionHistory)
	clubsGroup.Get("/:clubId/players/:id", s.handlers.HandleGetPlayer)
	clubsGroup.Post("/:clubId/players", s.handlers.HandleCreatePlayer)
	clubsGroup.Put("/:clubId/players/:id", s.handlers.HandleUpdatePlayer)
	clubsGroup.Post("/:clubId/players/:id/recalculate-balance", s.handlers.HandleRecalculatePlayerBalance)
	clubsGroup.Delete("/:clubId/players/:id", s.handlers.HandleDeletePlayer)
	clubsGroup.Post("/:clubId/players/:id/invite", s.handlers.HandleInvitePlayer)

	// Invitation endpoints
	api.Get("/invitations/:token", s.handlers.HandleGetInvitation)
	api.Post("/invitations/:token/accept", s.handlers.HandleAcceptInvitation)

	// Penalty type endpoints
	clubsGroup.Get("/:clubId/penalty-types", s.handlers.HandleGetPenaltyTypes)
	clubsGroup.Get("/:clubId/penalty-types/:id", s.handlers.HandleGetPenaltyType)
	clubsGroup.Post("/:clubId/penalty-types", s.handlers.HandleCreatePenaltyType)
	clubsGroup.Put("/:clubId/penalty-types/:id", s.handlers.HandleUpdatePenaltyType)
	clubsGroup.Put("/:clubId/penalty-types/:id/display-order", s.handlers.HandleUpdatePenaltyTypeDisplayOrder)
	clubsGroup.Delete("/:clubId/penalty-types/:id", s.handlers.HandleDeletePenaltyType)

	// Competition endpoints
	clubsGroup.Get("/:clubId/competitions", s.handlers.HandleGetCompetitions)
	clubsGroup.Get("/:clubId/competitions/:id", s.handlers.HandleGetCompetition)
	clubsGroup.Post("/:clubId/competitions", s.handlers.HandleCreateCompetition)
	clubsGroup.Put("/:clubId/competitions/:id", s.handlers.HandleUpdateCompetition)
	clubsGroup.Delete("/:clubId/competitions/:id", s.handlers.HandleDeleteCompetition)

	// Game Day endpoints
	clubsGroup.Get("/:clubId/gamedays", s.handlers.HandleGetGameDays)
	clubsGroup.Get("/:clubId/gamedays/summaries", s.handlers.HandleGetGameDaySummaries)
	clubsGroup.Get("/:clubId/gamedays/:id", s.handlers.HandleGetGameDay)
	clubsGroup.Post("/:clubId/gamedays", s.handlers.HandleCreateGameDay)
	clubsGroup.Put("/:clubId/gamedays/:id", s.handlers.HandleUpdateGameDay)
	clubsGroup.Delete("/:clubId/gamedays/:id", s.handlers.HandleDeleteGameDay)

	// Participant endpoints
	clubsGroup.Post("/:clubId/gamedays/:id/participants", s.handlers.HandleAddParticipant)
	clubsGroup.Delete("/:clubId/gamedays/:id/participants/:playerId", s.handlers.HandleRemoveParticipant)

	// Fee endpoints
	clubsGroup.Put("/:clubId/gamedays/:id/participants/:playerId/fees", s.handlers.HandleUpdateFees)

	// Competition value endpoints
	clubsGroup.Put("/:clubId/gamedays/:id/participants/:playerId/competition-values", s.handlers.HandleUpdateCompetitionValues)

	// Transaction endpoints
	clubsGroup.Get("/:clubId/transactions", s.handlers.HandleListTransactions)
	clubsGroup.Get("/:clubId/transactions/:id", s.handlers.HandleGetTransaction)
	clubsGroup.Post("/:clubId/transactions", s.handlers.HandleCreateTransaction)
	clubsGroup.Delete("/:clubId/transactions/:id", s.handlers.HandleDeleteTransaction)
	clubsGroup.Get("/:clubId/players/:playerId/transactions", s.handlers.HandleListPlayerTransactions)
	clubsGroup.Get("/:clubId/gamedays/:gamedayId/transactions", s.handlers.HandleListGameDayTransactions)
	clubsGroup.Get("/:clubId/gamedays/:gamedayId/transaction-summary", s.handlers.HandleGetGameDayTransactionSummary)

	// OpenAPI spec and Swagger UI (for technical users)
	api.Get("/openapi.yaml", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "application/x-yaml; charset=utf-8")
		return c.Send(openapi.OpenAPIYAML)
	})
	api.Get("/docs", serveSwaggerUI)
	registerStatic(s)
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

// serveSwaggerUI returns an HTML page that loads Swagger UI from CDN and the OpenAPI spec.
func serveSwaggerUI(c fiber.Ctx) error {
	const html = `<!DOCTYPE html>
<html lang="de">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Kegelmaster API – Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "/api/openapi.yaml",
        dom_id: "#swagger-ui",
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIBundle.SwaggerUIStandalonePreset
        ]
      });
    };
  </script>
</body>
</html>`
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return c.SendString(html)
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
