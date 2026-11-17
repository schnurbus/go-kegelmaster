package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/schnurbus/go-kegelmaster/backend/internal/config"
)

func TestErrorHandler_HidesInternalErrors(t *testing.T) {
	app := fiber.New(newFiberConfig(config.Config{}))
	app.Get("/internal", func(c fiber.Ctx) error {
		return errors.New(`pq: relation "secret_table" does not exist`)
	})
	app.Get("/client", func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadRequest, "Ungültige Anfrage")
	})

	check := func(path string, wantStatus int, wantMessage string) {
		t.Helper()
		req, _ := http.NewRequest(http.MethodGet, path, nil)
		resp, err := app.Test(req, fiber.TestConfig{})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != wantStatus {
			t.Fatalf("%s: expected %d, got %d", path, wantStatus, resp.StatusCode)
		}
		raw, _ := io.ReadAll(resp.Body)
		var body map[string]string
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("%s: decode: %v (%s)", path, err, raw)
		}
		if body["message"] != wantMessage {
			t.Errorf("%s: expected message %q, got %q", path, wantMessage, body["message"])
		}
	}
	check("/internal", fiber.StatusInternalServerError, "Interner Fehler")
	check("/client", fiber.StatusBadRequest, "Ungültige Anfrage")
}

func TestNewFiberConfig_TrustedProxies(t *testing.T) {
	fc := newFiberConfig(config.Config{ProxyHeader: "X-Forwarded-For", TrustedProxies: "private, 203.0.113.7 ,loopback"})
	if !fc.TrustProxy || fc.ProxyHeader != "X-Forwarded-For" {
		t.Fatalf("expected proxy trust with header, got %+v", fc)
	}
	if !fc.TrustProxyConfig.Private || !fc.TrustProxyConfig.Loopback || fc.TrustProxyConfig.LinkLocal {
		t.Errorf("unexpected range flags: %+v", fc.TrustProxyConfig)
	}
	if len(fc.TrustProxyConfig.Proxies) != 1 || fc.TrustProxyConfig.Proxies[0] != "203.0.113.7" {
		t.Errorf("unexpected proxies: %v", fc.TrustProxyConfig.Proxies)
	}

	if plain := newFiberConfig(config.Config{TrustedProxies: "private"}); plain.TrustProxy {
		t.Error("without PROXY_HEADER the proxy must not be trusted")
	}
}
