package audit

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v3"
)

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

func TestLogAudit(t *testing.T) {
	cap := &captureHandler{}
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(cap))
	defer slog.SetDefault(oldLogger)

	app := fiber.New()
	app.Post("/test", func(c fiber.Ctx) error {
		LogAudit(c, "user-123", ActionClubCreated, "club_id", "club-456")
		return c.SendStatus(fiber.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodPost, "/test", nil)
	_, err := app.Test(req, fiber.TestConfig{})
	if err != nil {
		t.Fatal(err)
	}

	records := cap.getRecords()
	if len(records) != 1 {
		t.Fatalf("expected 1 audit log, got %d", len(records))
	}
	r := records[0]
	if r.Message != "audit" {
		t.Errorf("message: got %q", r.Message)
	}
	attrs := make(map[string]slog.Value)
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value
		return true
	})
	if attrs["event"].String() != "audit" {
		t.Errorf("event: got %q", attrs["event"].String())
	}
	if attrs["action"].String() != ActionClubCreated {
		t.Errorf("action: got %q", attrs["action"].String())
	}
	if attrs["user_id"].String() != "user-123" {
		t.Errorf("user_id: got %q", attrs["user_id"].String())
	}
	if attrs["club_id"].String() != "club-456" {
		t.Errorf("club_id: got %q", attrs["club_id"].String())
	}
}

func TestLogAudit_anonymousWhenUserIDEmpty(t *testing.T) {
	cap := &captureHandler{}
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(cap))
	defer slog.SetDefault(oldLogger)

	app := fiber.New()
	app.Post("/test", func(c fiber.Ctx) error {
		LogAudit(c, "", ActionUserRegistered, "target_user_id", "new-user-id")
		return c.SendStatus(fiber.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodPost, "/test", nil)
	_, err := app.Test(req, fiber.TestConfig{})
	if err != nil {
		t.Fatal(err)
	}

	records := cap.getRecords()
	if len(records) != 1 {
		t.Fatalf("expected 1 audit log, got %d", len(records))
	}
	attrs := make(map[string]string)
	records[0].Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.String()
		return true
	})
	if attrs["user_id"] != "anonymous" {
		t.Errorf("user_id (actor) when empty should be anonymous, got %q", attrs["user_id"])
	}
	if attrs["target_user_id"] != "new-user-id" {
		t.Errorf("target_user_id: got %q", attrs["target_user_id"])
	}
}
