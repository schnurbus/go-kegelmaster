package config

import (
	"os"
	"strings"
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

func TestPasskeyRelyingParty(t *testing.T) {
	tests := []struct {
		name        string
		cfg         Config
		wantRPID    string
		wantOrigins []string
		wantErr     bool
	}{
		{
			name:        "derived from base URL plus CORS origins",
			cfg:         Config{BaseURL: "http://localhost:8080", CORSOrigins: "http://localhost:5173, http://localhost:8080/"},
			wantRPID:    "localhost",
			wantOrigins: []string{"http://localhost:8080", "http://localhost:5173"},
		},
		{
			name:        "production host",
			cfg:         Config{BaseURL: "https://kegel.example.org/", CORSOrigins: "https://*.example.org"},
			wantRPID:    "kegel.example.org",
			wantOrigins: []string{"https://kegel.example.org"},
		},
		{
			name:        "explicit settings win",
			cfg:         Config{BaseURL: "https://app.example.org", CORSOrigins: "http://localhost:5173", WebAuthnRPID: "example.org", WebAuthnOrigins: "https://app.example.org,https://www.example.org"},
			wantRPID:    "example.org",
			wantOrigins: []string{"https://app.example.org", "https://www.example.org"},
		},
		{
			name:    "relative base URL",
			cfg:     Config{BaseURL: "/kegelmaster"},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rpID, origins, err := tc.cfg.PasskeyRelyingParty()
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if rpID != tc.wantRPID {
				t.Fatalf("rpID = %q, want %q", rpID, tc.wantRPID)
			}
			if strings.Join(origins, " ") != strings.Join(tc.wantOrigins, " ") {
				t.Fatalf("origins = %v, want %v", origins, tc.wantOrigins)
			}
		})
	}
}
