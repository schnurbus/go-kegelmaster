package config

import (
	"os"
	"testing"
)

func TestDebugLogRequests(t *testing.T) {
	t.Run("default is false", func(t *testing.T) {
		os.Unsetenv("LOG_LEVEL")
		cfg := Load()
		if cfg.DebugLogRequests() {
			t.Error("expected DebugLogRequests false when LOG_LEVEL unset")
		}
	})
	t.Run("info is false", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "info")
		defer os.Unsetenv("LOG_LEVEL")
		cfg := Load()
		if cfg.DebugLogRequests() {
			t.Error("expected DebugLogRequests false when LOG_LEVEL=info")
		}
	})
	t.Run("debug is true", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "debug")
		defer os.Unsetenv("LOG_LEVEL")
		cfg := Load()
		if !cfg.DebugLogRequests() {
			t.Error("expected DebugLogRequests true when LOG_LEVEL=debug")
		}
	})
	t.Run("DEBUG is case insensitive", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "DEBUG")
		defer os.Unsetenv("LOG_LEVEL")
		cfg := Load()
		if !cfg.DebugLogRequests() {
			t.Error("expected DebugLogRequests true when LOG_LEVEL=DEBUG")
		}
	})
}
