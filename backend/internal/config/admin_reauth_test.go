package config

import (
	"testing"
	"time"
)

func TestAdminReauthMinutes(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		want    time.Duration
		wantErr bool
	}{
		{name: "unset defaults to 15", env: "", want: 15 * time.Minute},
		{name: "custom", env: " 30 ", want: 30 * time.Minute},
		{name: "maximum", env: "1440", want: 24 * time.Hour},
		{name: "zero", env: "0", wantErr: true},
		{name: "negative", env: "-5", wantErr: true},
		{name: "too large", env: "1441", wantErr: true},
		{name: "not a number", env: "15m", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ADMIN_REAUTH_MINUTES", tt.env)
			cfg := Load()
			err := cfg.ValidateAdminReauth()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.env)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := cfg.AdminReauthWindow(); got != tt.want {
				t.Errorf("window = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestAdminReauthWindowZeroConfig(t *testing.T) {
	if got := (Config{}).AdminReauthWindow(); got != DefaultAdminReauthMinutes*time.Minute {
		t.Errorf("window = %s, want default", got)
	}
}
