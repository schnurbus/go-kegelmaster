package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Grenzen für ADMIN_REAUTH_MINUTES: Zeitfenster seit der letzten echten
// Anmeldung (Passwort oder Passkey), in dem der Admin-Bereich nutzbar ist.
const (
	DefaultAdminReauthMinutes = 15
	MaxAdminReauthMinutes     = 24 * 60
)

// parseAdminReauthMinutes reads ADMIN_REAUTH_MINUTES. Empty means the
// default; anything that is not a positive integer yields -1, so that
// ValidateAdminReauth rejects it instead of silently using the default.
func parseAdminReauthMinutes(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultAdminReauthMinutes
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 1 {
		return -1
	}
	return v
}

// AdminReauthWindow returns how long after the last real authentication the
// admin area stays usable. A zero value (e.g. a zero Config in tests) means
// the default.
func (c Config) AdminReauthWindow() time.Duration {
	minutes := c.AdminReauthMinutes
	if minutes == 0 {
		minutes = DefaultAdminReauthMinutes
	}
	return time.Duration(minutes) * time.Minute
}

// ValidateAdminReauth rejects invalid ADMIN_REAUTH_MINUTES values at startup.
func (c Config) ValidateAdminReauth() error {
	if c.AdminReauthMinutes < 0 || c.AdminReauthMinutes > MaxAdminReauthMinutes {
		return fmt.Errorf("ADMIN_REAUTH_MINUTES must be an integer between 1 and %d", MaxAdminReauthMinutes)
	}
	return nil
}
