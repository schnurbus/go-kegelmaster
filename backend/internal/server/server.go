package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"

	"github.com/schnurbus/go-kegelmaster/backend/internal/admin"
	"github.com/schnurbus/go-kegelmaster/backend/internal/apitoken"
	"github.com/schnurbus/go-kegelmaster/backend/internal/auth"
	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/competition"
	"github.com/schnurbus/go-kegelmaster/backend/internal/config"
	"github.com/schnurbus/go-kegelmaster/backend/internal/email"
	"github.com/schnurbus/go-kegelmaster/backend/internal/gameday"
	"github.com/schnurbus/go-kegelmaster/backend/internal/handlers"
	"github.com/schnurbus/go-kegelmaster/backend/internal/invitation"
	"github.com/schnurbus/go-kegelmaster/backend/internal/mcpserver"
	"github.com/schnurbus/go-kegelmaster/backend/internal/middleware"
	"github.com/schnurbus/go-kegelmaster/backend/internal/passkey"
	"github.com/schnurbus/go-kegelmaster/backend/internal/passwordreset"
	"github.com/schnurbus/go-kegelmaster/backend/internal/penaltytype"
	"github.com/schnurbus/go-kegelmaster/backend/internal/permission"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
	"github.com/schnurbus/go-kegelmaster/backend/internal/statistics"
	"github.com/schnurbus/go-kegelmaster/backend/internal/transaction"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
	"github.com/schnurbus/go-kegelmaster/backend/openapi"
)

// Dependencies aggregates components the HTTP server relies on.
type Dependencies struct {
	UserRepo          *user.Repository
	ClubRepo          *club.Repository
	RoleRepo          *role.Repository
	PlayerRepo        *player.Repository
	PenaltyTypeRepo   *penaltytype.Repository
	CompetitionRepo   *competition.Repository
	GameDayRepo       *gameday.Repository
	TransactionRepo   *transaction.Repository
	AuthService       *auth.Service
	PermissionCheck   *permission.Checker
	InvitationRepo    *invitation.Repository
	EmailService      *email.Service
	PasswordResetRepo *passwordreset.Repository
	APITokenRepo      *apitoken.Repository
	StatisticsRepo    *statistics.Repository
	PasskeyService    *passkey.Service
	GameDayNotifier   handlers.GameDayNotifier
	// AdminRepo backs the read-only operator area /api/admin; nil disables it (404).
	AdminRepo *admin.Repository
}

// Server wraps the Fiber application and its dependencies.
type Server struct {
	cfg      config.Config
	app      *fiber.App
	handlers *handlers.Handler
	mcpAPI   *mcpserver.InProcessAPI
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

	app := fiber.New(newFiberConfig(cfg))
	// A panic in one handler must not take the whole process down.
	app.Use(recover.New())
	app.Use(securityHeaders)
	app.Use(cors.New(cors.Config{
		AllowOrigins:     splitAndTrim(cfg.CORSOrigins),
		AllowHeaders:     []string{"Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	}))

	userIDExtractor := func(c fiber.Ctx) string {
		// Set by the handler once it authenticated the user (covers API tokens).
		if id, ok := c.Locals(handlers.LocalsUserID).(string); ok && id != "" {
			return id
		}
		token := strings.TrimSpace(c.Cookies("auth_token"))
		if token == "" {
			return ""
		}
		claims, err := deps.AuthService.ParseToken(token)
		if err != nil {
			return ""
		}
		return claims.UserID
	}
	app.Use(middleware.RequestLog(middleware.RequestLogFromConfig(cfg, userIDExtractor)))

	handler := handlers.NewHandler(cfg, handlers.Dependencies{
		UserRepo:          deps.UserRepo,
		ClubRepo:          deps.ClubRepo,
		RoleRepo:          deps.RoleRepo,
		PlayerRepo:        deps.PlayerRepo,
		PenaltyTypeRepo:   deps.PenaltyTypeRepo,
		CompetitionRepo:   deps.CompetitionRepo,
		GameDayRepo:       deps.GameDayRepo,
		TransactionRepo:   deps.TransactionRepo,
		AuthService:       deps.AuthService,
		PermissionCheck:   deps.PermissionCheck,
		InvitationRepo:    deps.InvitationRepo,
		EmailService:      deps.EmailService,
		PasswordResetRepo: deps.PasswordResetRepo,
		APITokenRepo:      deps.APITokenRepo,
		StatisticsRepo:    deps.StatisticsRepo,
		PasskeyService:    deps.PasskeyService,
		GameDayNotifier:   deps.GameDayNotifier,
		AdminRepo:         deps.AdminRepo,
	})

	server := &Server{
		cfg:      cfg,
		app:      app,
		handlers: handler,
		mcpAPI:   mcpserver.NewInProcessAPI(),
	}

	server.registerRoutes()
	// MCP tools call the REST API in-process; the handler is only complete
	// once every route is registered.
	server.mcpAPI.SetHandler(app.Handler())
	return server
}

func (s *Server) registerRoutes() {
	s.app.Get("/healthz", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": s.cfg.Version,
		})
	})

	// MCP server (Streamable HTTP, stateless) for AI agents; authenticated with
	// personal API tokens. See internal/mcpserver.
	s.app.All("/mcp", adaptor.HTTPHandler(mcpserver.New(s.mcpAPI, s.cfg.Version)))

	api := s.app.Group("/api")
	api.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "backend placeholder",
		})
	})
	// Public config for frontend (version, optional buy-me-a-coffee URL); no auth
	api.Get("/config/public", func(c fiber.Ctx) error {
		m := fiber.Map{
			"version":      s.cfg.Version,
			"chat_enabled": strings.TrimSpace(s.cfg.GeminiAPIKey) != "",
		}
		if s.cfg.BuyMeACoffeeURL != "" {
			m["buy_me_a_coffee_url"] = s.cfg.BuyMeACoffeeURL
		}
		return c.JSON(m)
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

	// Legal pages (public, no auth): content from file path or env
	api.Get("/legal/impressum", s.handleLegalPage("impressum", s.cfg.LegalImpressumPath, s.cfg.ImpressumHTML, "Impressum"))
	api.Get("/legal/datenschutz", s.handleLegalPage("datenschutz", s.cfg.LegalDatenschutzPath, s.cfg.DatenschutzHTML, "Datenschutz"))

	authGroup := api.Group("/auth")
	authGroup.Get("/csrf-token", s.handlers.HandleCSRFCookie)
	authGroup.Get("/me", s.handlers.HandleCurrentUser)
	authGroup.Patch("/me", s.handlers.HandleUpdateCurrentUser)
	authGroup.Post("/register", s.handlers.HandleRegister)
	// Brute-force protection: per client IP (see PROXY_HEADER behind a reverse proxy).
	authGroup.Post("/login", limiter.New(limiter.Config{
		Max:        10,
		Expiration: 5 * time.Minute,
	}), s.handlers.HandleLogin)
	authGroup.Post("/logout", s.handlers.HandleLogout)
	// Rate limited like login: the current password is verified here.
	authGroup.Post("/change-password", limiter.New(limiter.Config{
		Max:        10,
		Expiration: 5 * time.Minute,
	}), s.handlers.HandleChangePassword)
	authGroup.Get("/tokens", s.handlers.HandleListAPITokens)
	authGroup.Post("/tokens", s.handlers.HandleCreateAPIToken)
	authGroup.Delete("/tokens/:id", s.handlers.HandleDeleteAPIToken)
	// Passkeys: login endpoints are public and rate limited like /login;
	// register/begin verifies the current password and is limited as well.
	passkeyLimiter := limiter.New(limiter.Config{
		Max:        20,
		Expiration: 5 * time.Minute,
	})
	authGroup.Post("/passkeys/login/begin", passkeyLimiter, s.handlers.HandleBeginPasskeyLogin)
	authGroup.Post("/passkeys/login/finish", passkeyLimiter, s.handlers.HandleFinishPasskeyLogin)
	authGroup.Post("/passkeys/register/begin", limiter.New(limiter.Config{
		Max:        10,
		Expiration: 5 * time.Minute,
	}), s.handlers.HandleBeginPasskeyRegistration)
	authGroup.Post("/passkeys/register/finish", s.handlers.HandleFinishPasskeyRegistration)
	authGroup.Get("/passkeys", s.handlers.HandleListPasskeys)
	authGroup.Patch("/passkeys/:id", s.handlers.HandleRenamePasskey)
	authGroup.Delete("/passkeys/:id", s.handlers.HandleDeletePasskey)
	authGroup.Post("/forgot-password", limiter.New(limiter.Config{
		Max:        5,
		Expiration: 15 * time.Minute,
	}), s.handlers.HandleForgotPassword)
	authGroup.Post("/reset-password", limiter.New(limiter.Config{
		Max:        10,
		Expiration: 15 * time.Minute,
	}), s.handlers.HandleResetPassword)

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
	clubsGroup.Get("/:clubId/players/me/statistics", s.handlers.HandleGetMyStatistics)
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
	clubsGroup.Patch("/:clubId/gamedays/:id/participants/:playerId", s.handlers.HandleSetParticipantAttendance)

	// Fee endpoints
	clubsGroup.Put("/:clubId/gamedays/:id/participants/:playerId/fees", s.handlers.HandleUpdateFees)

	// Competition value endpoints
	clubsGroup.Put("/:clubId/gamedays/:id/participants/:playerId/competition-values", s.handlers.HandleUpdateCompetitionValues)

	// Transaction endpoints
	clubsGroup.Get("/:clubId/transactions", s.handlers.HandleListTransactions)
	clubsGroup.Get("/:clubId/transactions/:id", s.handlers.HandleGetTransaction)
	clubsGroup.Post("/:clubId/transactions", s.handlers.HandleCreateTransaction)
	clubsGroup.Put("/:clubId/transactions/:id", s.handlers.HandleUpdateTransaction)
	clubsGroup.Delete("/:clubId/transactions/:id", s.handlers.HandleDeleteTransaction)
	clubsGroup.Get("/:clubId/players/:playerId/transactions", s.handlers.HandleListPlayerTransactions)
	clubsGroup.Get("/:clubId/gamedays/:gamedayId/transactions", s.handlers.HandleListGameDayTransactions)
	clubsGroup.Get("/:clubId/gamedays/:gamedayId/transaction-summary", s.handlers.HandleGetGameDayTransactionSummary)

	// Instance admin (operator) area, read-only. Cookie sessions of instance
	// admins only; everyone else (incl. API tokens and MCP) gets 404.
	adminGroup := api.Group("/admin", s.handlers.RequireInstanceAdmin)
	adminGroup.Get("/stats", s.handlers.HandleAdminStats)
	adminGroup.Get("/users", s.handlers.HandleAdminListUsers)
	adminGroup.Get("/users/:id", s.handlers.HandleAdminGetUser)
	adminGroup.Get("/clubs", s.handlers.HandleAdminListClubs)
	adminGroup.Get("/clubs/:id", s.handlers.HandleAdminGetClub)
	adminGroup.Get("/export/memberships.csv", s.handlers.HandleAdminExportMemberships)

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

// handleLegalPage returns a handler that serves legal page content from file path or inline HTML.
// If neither is set, returns a short placeholder so the app does not break.
func (s *Server) handleLegalPage(name, filePath, inlineHTML, title string) func(fiber.Ctx) error {
	return func(c fiber.Ctx) error {
		var body string
		if inlineHTML != "" {
			body = inlineHTML
		} else if filePath != "" && filepath.IsAbs(filepath.Clean(filePath)) {
			abs := filepath.Clean(filePath)
			b, err := os.ReadFile(abs)
			if err != nil {
				slog.Warn("legal page file read failed", "name", name, "path", abs, "error", err)
				body = "<p>Inhalt derzeit nicht verfügbar.</p>"
			} else {
				body = string(b)
			}
		} else {
			body = "<p>" + title + " wird konfiguriert.</p>"
		}
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.SendString(body)
	}
}

// newFiberConfig builds the Fiber config: error handler and reverse-proxy trust.
func newFiberConfig(cfg config.Config) fiber.Config {
	fc := fiber.Config{ErrorHandler: errorHandler}
	if cfg.ProxyHeader != "" {
		// Without this, c.IP() is the proxy address behind a reverse proxy and
		// every rate limit is shared by all clients.
		fc.ProxyHeader = cfg.ProxyHeader
		fc.TrustProxy = true
		for _, p := range splitList(cfg.TrustedProxies) {
			switch strings.ToLower(p) {
			case "private":
				fc.TrustProxyConfig.Private = true
			case "loopback":
				fc.TrustProxyConfig.Loopback = true
			case "linklocal":
				fc.TrustProxyConfig.LinkLocal = true
			default:
				fc.TrustProxyConfig.Proxies = append(fc.TrustProxyConfig.Proxies, p)
			}
		}
	}
	return fc
}

// errorHandler returns the message of fiber errors to the client. Any other
// error is logged and answered with a generic message, so database or driver
// details never reach the client.
func errorHandler(c fiber.Ctx, err error) error {
	var e *fiber.Error
	if errors.As(err, &e) {
		return c.Status(e.Code).JSON(fiber.Map{"message": e.Message})
	}
	slog.Error("unhandled error", "method", c.Method(), "path", c.Path(), "error", err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Interner Fehler"})
}

// securityHeaders sets conservative browser security headers on every response.
func securityHeaders(c fiber.Ctx) error {
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("X-Frame-Options", "SAMEORIGIN")
	c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	return c.Next()
}

// splitList splits a comma-separated list and drops empty entries.
func splitList(input string) []string {
	values := strings.Split(input, ",")
	result := make([]string, 0, len(values))
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
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
