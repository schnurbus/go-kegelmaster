package handlers

import (
	"github.com/schnurbus/go-kegelmaster/backend/internal/auth"
	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/config"
	"github.com/schnurbus/go-kegelmaster/backend/internal/email"
	"github.com/schnurbus/go-kegelmaster/backend/internal/gameday"
	"github.com/schnurbus/go-kegelmaster/backend/internal/invitation"
	"github.com/schnurbus/go-kegelmaster/backend/internal/penaltytype"
	"github.com/schnurbus/go-kegelmaster/backend/internal/permission"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
	"github.com/schnurbus/go-kegelmaster/backend/internal/transaction"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

// Handler contains all dependencies needed for HTTP handlers.
type Handler struct {
	Config            config.Config
	UserRepo          *user.Repository
	ClubRepo          *club.Repository
	RoleRepo          *role.Repository
	PlayerRepo        *player.Repository
	PenaltyTypeRepo   *penaltytype.Repository
	GameDayRepo       *gameday.Repository
	TransactionRepo   *transaction.Repository
	AuthSvc           *auth.Service
	PermissionChecker *permission.Checker
	InvitationRepo    *invitation.Repository
	EmailSvc          *email.Service
}

// NewHandler creates a new handler instance with the given dependencies.
func NewHandler(cfg config.Config, deps Dependencies) *Handler {
	return &Handler{
		Config:            cfg,
		UserRepo:          deps.UserRepo,
		ClubRepo:          deps.ClubRepo,
		RoleRepo:          deps.RoleRepo,
		PlayerRepo:        deps.PlayerRepo,
		PenaltyTypeRepo:   deps.PenaltyTypeRepo,
		GameDayRepo:       deps.GameDayRepo,
		TransactionRepo:   deps.TransactionRepo,
		AuthSvc:           deps.AuthService,
		PermissionChecker: deps.PermissionCheck,
		InvitationRepo:    deps.InvitationRepo,
		EmailSvc:          deps.EmailService,
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
	TransactionRepo *transaction.Repository
	AuthService     *auth.Service
	PermissionCheck *permission.Checker
	InvitationRepo  *invitation.Repository
	EmailService    *email.Service
}
