package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// captureHandler is a slog.Handler that records log records for testing.
type captureHandler struct {
	mu      sync.Mutex
	records []slog.Record
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = append(h.records, r.Clone())
	return nil
}

func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(_ string) slog.Handler      { return h }

func (h *captureHandler) getRecords() []slog.Record {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]slog.Record, len(h.records))
	copy(out, h.records)
	return out
}

func (h *captureHandler) reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = nil
}

func TestRequestLog_GET_skippedWhenNotDebug(t *testing.T) {
	cap := &captureHandler{}
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(cap))
	defer slog.SetDefault(oldLogger)

	app := fiber.New()
	app.Use(RequestLog(RequestLogConfig{DebugLogRequests: false}))
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	_, err := app.Test(req, fiber.TestConfig{})
	if err != nil {
		t.Fatal(err)
	}

	records := cap.getRecords()
	if len(records) != 0 {
		t.Errorf("expected no log when GET and not debug, got %d records", len(records))
	}
}

func TestRequestLog_GET_loggedWhenDebug(t *testing.T) {
	cap := &captureHandler{}
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(cap))
	defer slog.SetDefault(oldLogger)

	app := fiber.New()
	app.Use(RequestLog(RequestLogConfig{DebugLogRequests: true}))
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	_, err := app.Test(req, fiber.TestConfig{})
	if err != nil {
		t.Fatal(err)
	}

	records := cap.getRecords()
	if len(records) != 1 {
		t.Fatalf("expected 1 log when GET and debug, got %d", len(records))
	}
	r := records[0]
	if r.Message != "request" {
		t.Errorf("message: got %q", r.Message)
	}
	var mutation bool
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "mutation" {
			mutation = a.Value.Bool()
		}
		return true
	})
	if mutation {
		t.Error("GET request should have mutation=false")
	}
}

func TestRequestLog_POST_alwaysLoggedWithMutationTrue(t *testing.T) {
	cap := &captureHandler{}
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(cap))
	defer slog.SetDefault(oldLogger)

	app := fiber.New()
	app.Use(RequestLog(RequestLogConfig{DebugLogRequests: false}))
	app.Post("/test", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusCreated)
	})

	req, _ := http.NewRequest(http.MethodPost, "/test", nil)
	_, err := app.Test(req, fiber.TestConfig{})
	if err != nil {
		t.Fatal(err)
	}

	records := cap.getRecords()
	if len(records) != 1 {
		t.Fatalf("expected 1 log for POST, got %d", len(records))
	}
	r := records[0]
	var method string
	var mutation bool
	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "method":
			method = a.Value.String()
		case "mutation":
			mutation = a.Value.Bool()
		}
		return true
	})
	if method != "POST" {
		t.Errorf("method: got %q", method)
	}
	if !mutation {
		t.Error("POST request should have mutation=true")
	}
}

func TestRequestLog_userIDExtractor(t *testing.T) {
	cap := &captureHandler{}
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(cap))
	defer slog.SetDefault(oldLogger)

	extractor := func(c fiber.Ctx) string {
		return c.Get("X-Test-User-ID")
	}
	app := fiber.New()
	app.Use(RequestLog(RequestLogConfig{DebugLogRequests: true, UserIDExtractor: extractor}))
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Test-User-ID", "user-123")
	_, err := app.Test(req, fiber.TestConfig{})
	if err != nil {
		t.Fatal(err)
	}

	records := cap.getRecords()
	if len(records) != 1 {
		t.Fatalf("expected 1 log, got %d", len(records))
	}
	var userID string
	records[0].Attrs(func(a slog.Attr) bool {
		if a.Key == "user_id" {
			userID = a.Value.String()
		}
		return true
	})
	if userID != "user-123" {
		t.Errorf("user_id: got %q", userID)
	}
}

// Handlers report failures by returning an error; those requests must be
// logged with the error's status even though the app's error handler sets
// the response status only later.
func TestRequestLog_ReturnedErrorIsLogged(t *testing.T) {
	cap := &captureHandler{}
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(cap))
	defer slog.SetDefault(oldLogger)

	app := fiber.New()
	app.Use(RequestLog(RequestLogConfig{DebugLogRequests: false}))
	app.Get("/forbidden", func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusForbidden, "nope")
	})

	req, _ := http.NewRequest(http.MethodGet, "/forbidden", nil)
	resp, err := app.Test(req, fiber.TestConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}

	records := cap.getRecords()
	if len(records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(records))
	}
	var status int64
	records[0].Attrs(func(a slog.Attr) bool {
		if a.Key == "status" {
			status = a.Value.Int64()
		}
		return true
	})
	if status != fiber.StatusForbidden {
		t.Errorf("expected logged status 403, got %d", status)
	}
	if records[0].Level != slog.LevelWarn {
		t.Errorf("expected warn level, got %v", records[0].Level)
	}
}
