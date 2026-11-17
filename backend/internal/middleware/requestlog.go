package middleware

import (
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/config"
)

const requestStartKey = "request_start"

// RequestLogConfig configures the request logging middleware.
type RequestLogConfig struct {
	DebugLogRequests bool
	UserIDExtractor  func(fiber.Ctx) string
}

// RequestLog returns a Fiber middleware that logs HTTP requests.
// GET/HEAD are only logged when cfg.DebugLogRequests is true.
// POST/PUT/PATCH/DELETE are always logged with mutation=true.
func RequestLog(cfg RequestLogConfig) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		c.Locals(requestStartKey, start)

		err := c.Next()

		// Handlers report failures by returning an error; the app's error
		// handler sets the status only after this middleware returns. Derive
		// the status here so failed requests are logged as well.
		status := c.Response().StatusCode()
		if err != nil {
			status = fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				status = fe.Code
			}
		}
		method := c.Method()
		path := c.Path()
		ip := c.IP()
		latency := time.Since(start).Milliseconds()

		isReadOnly := method == "GET" || method == "HEAD"
		if isReadOnly && !cfg.DebugLogRequests && status < 400 {
			return err
		}

		mutation := !isReadOnly
		attrs := []any{
			"method", method,
			"path", path,
			"status", status,
			"latency_ms", latency,
			"ip", ip,
			"mutation", mutation,
		}
		if cfg.UserIDExtractor != nil {
			if userID := cfg.UserIDExtractor(c); userID != "" {
				attrs = append(attrs, "user_id", userID)
			}
		}

		var level slog.Level
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		default:
			level = slog.LevelInfo
		}
		slog.Log(c.Context(), level, "request", attrs...)

		return err
	}
}

// RequestLogFromConfig builds RequestLogConfig from app config and an optional UserID extractor.
func RequestLogFromConfig(cfg config.Config, userIDExtractor func(fiber.Ctx) string) RequestLogConfig {
	return RequestLogConfig{
		DebugLogRequests: cfg.DebugLogRequests(),
		UserIDExtractor:  userIDExtractor,
	}
}
