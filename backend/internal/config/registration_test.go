package config

import "testing"

func TestRegistrationMode(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		want    string
		wantErr bool
	}{
		{name: "unset or empty defaults to open", env: "", want: RegistrationModeOpen},
		{name: "open", env: "open", want: RegistrationModeOpen},
		{name: "invite_only", env: "invite_only", want: RegistrationModeInviteOnly},
		{name: "closed with spaces and upper case", env: " CLOSED ", want: RegistrationModeClosed},
		{name: "invalid", env: "invite-only", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("REGISTRATION_MODE", tt.env)
			cfg := Load()
			err := cfg.ValidateRegistrationMode()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.env)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := cfg.EffectiveRegistrationMode(); got != tt.want {
				t.Errorf("mode = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEffectiveRegistrationModeZeroConfig(t *testing.T) {
	if got := (Config{}).EffectiveRegistrationMode(); got != RegistrationModeOpen {
		t.Errorf("zero config mode = %q, want open", got)
	}
}
