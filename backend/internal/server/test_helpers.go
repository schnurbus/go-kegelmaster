package server

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/auth"
	"github.com/schnurbus/go-kegelmaster/backend/internal/club"
	"github.com/schnurbus/go-kegelmaster/backend/internal/config"
	"github.com/schnurbus/go-kegelmaster/backend/internal/gameday"
	"github.com/schnurbus/go-kegelmaster/backend/internal/penaltytype"
	"github.com/schnurbus/go-kegelmaster/backend/internal/passwordreset"
	"github.com/schnurbus/go-kegelmaster/backend/internal/permission"
	"github.com/schnurbus/go-kegelmaster/backend/internal/player"
	"github.com/schnurbus/go-kegelmaster/backend/internal/role"
	"github.com/schnurbus/go-kegelmaster/backend/internal/transaction"
	"github.com/schnurbus/go-kegelmaster/backend/internal/user"
)

const testCSRFToken = "csrf-test-token"

// Cookie names for tests (matching handlers package)
const (
	csrfCookieName = "csrf_token"
	authCookieName = "auth_token"
)

func newTestServer(t *testing.T) (*Server, sqlmock.Sqlmock, *auth.Service) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	userRepo := user.NewRepository(db)
	clubRepo := club.NewRepository(db)
	roleRepo := role.NewRepository(db)
	playerRepo := player.NewRepository(db)
	penaltyTypeRepo := penaltytype.NewRepository(db)
	gameDayRepo := gameday.NewRepository(db)
	transactionRepo := transaction.NewRepository(db, playerRepo, clubRepo, gameDayRepo)
	authSvc := auth.NewService("test-secret", 60, 30)
	permissionCheck := permission.NewChecker(clubRepo, roleRepo, playerRepo)
	passwordResetRepo := passwordreset.NewRepository(db)

	srv := New(config.Config{
		AppEnv:      "test",
		HTTPPort:    "0",
		CORSOrigins: "http://localhost:5173",
	}, Dependencies{
		UserRepo:          userRepo,
		ClubRepo:          clubRepo,
		RoleRepo:          roleRepo,
		PlayerRepo:        playerRepo,
		PenaltyTypeRepo:   penaltyTypeRepo,
		GameDayRepo:       gameDayRepo,
		TransactionRepo:  transactionRepo,
		AuthService:       authSvc,
		PermissionCheck:   permissionCheck,
		PasswordResetRepo: passwordResetRepo,
	})

	return srv, mock, authSvc
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func mustJSONRequest(t *testing.T, method, path, body string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, path, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		req.Header.Set("X-CSRF-Token", testCSRFToken)
		req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: testCSRFToken})
	}
	return req
}

func doRequest(t *testing.T, srv *Server, req *http.Request) *http.Response {
	t.Helper()
	resp, err := srv.app.Test(req, fiber.TestConfig{})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}
