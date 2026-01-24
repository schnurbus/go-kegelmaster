package handlers

import (
	"github.com/schnurbus/go-kegelmaster/backend/internal/auth"
	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/config"
	"github.com/schnurbus/go-kegelmaster/backend/internal/gameday"
	"github.com/schnurbus/go-kegelmaster/backend/internal/penaltytype"
	"github.com/schnurbus/go-kegelmaster/backend/internal/permission"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

// Handler contains all dependencies needed for HTTP handlers.
type Handler struct {
	Config          config.Config
	UserRepo        *user.Repository
	ClubRepo        *club.Repository
	RoleRepo        *role.Repository
	PlayerRepo      *player.Repository
	PenaltyTypeRepo *penaltytype.Repository
	GameDayRepo     *gameday.Repository
	AuthSvc         *auth.Service
	PermissionCheck *permission.Checker
}

// NewHandler creates a new handler instance with the given dependencies.
func NewHandler(cfg config.Config, deps Dependencies) *Handler {
	return &Handler{
		Config:          cfg,
		UserRepo:        deps.UserRepo,
		ClubRepo:        deps.ClubRepo,
		RoleRepo:        deps.RoleRepo,
		PlayerRepo:      deps.PlayerRepo,
		PenaltyTypeRepo: deps.PenaltyTypeRepo,
		GameDayRepo:     deps.GameDayRepo,
		AuthSvc:         deps.AuthService,
		PermissionCheck: deps.PermissionCheck,
	}
}

// Dependencies aggregates components the HTTP handlers rely on.
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
