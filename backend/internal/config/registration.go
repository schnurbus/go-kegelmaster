package config

import "fmt"

// Werte für REGISTRATION_MODE.
const (
	// RegistrationModeOpen: jeder darf ein Konto anlegen (Standard, rückwärtskompatibel).
	RegistrationModeOpen = "open"
	// RegistrationModeInviteOnly: Konto nur mit gültiger Einladung (player_invitations)
	// für genau die eingeladene E-Mail-Adresse.
	RegistrationModeInviteOnly = "invite_only"
	// RegistrationModeClosed: keine neuen Konten.
	RegistrationModeClosed = "closed"
)

// EffectiveRegistrationMode returns the configured registration mode; an
// empty value (e.g. a zero Config in tests) means "open".
func (c Config) EffectiveRegistrationMode() string {
	if c.RegistrationMode == "" {
		return RegistrationModeOpen
	}
	return c.RegistrationMode
}

// ValidateRegistrationMode rejects unknown REGISTRATION_MODE values, so a
// typo never silently opens (or closes) registration.
func (c Config) ValidateRegistrationMode() error {
	switch c.EffectiveRegistrationMode() {
	case RegistrationModeOpen, RegistrationModeInviteOnly, RegistrationModeClosed:
		return nil
	}
	return fmt.Errorf("REGISTRATION_MODE %q is invalid (allowed: %s, %s, %s)",
		c.RegistrationMode, RegistrationModeOpen, RegistrationModeInviteOnly, RegistrationModeClosed)
}
