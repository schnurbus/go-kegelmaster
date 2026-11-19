package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config aggregates runtime configuration for the backend HTTP server.
// Version is typically injected at build time via ldflags (e.g. -X main.Version=...).
type Config struct {
	Version                     string
	AppEnv                      string
	HTTPPort                    string
	BaseURL                     string
	DatabaseURL                 string
	RedisURL                    string
	JWTSecret                   string
	TokenTTLMin                 int
	RememberMeDays              int
	CORSOrigins                 string
	ResendAPIKey                string
	ResendFromEmail             string
	PasswordResetTokenExpiryMin int
	GeminiAPIKey                string
	GeminiModel                 string
	ChatRateLimitMax            int
	ChatRateLimitWindowMin      int
	AutoMigrate                 bool
	// RegistrationMode: wer ein Konto anlegen darf (REGISTRATION_MODE), siehe registration.go.
	RegistrationMode string
	// Legal pages: file path (read at request time) or inline HTML from env.
	// LEGAL_IMPRESSUM_PATH / LEGAL_DATENSCHUTZ_PATH, or IMPRESSUM_HTML / DATENSCHUTZ_HTML.
	LegalImpressumPath   string
	LegalDatenschutzPath string
	ImpressumHTML        string
	DatenschutzHTML      string
	// BuyMeACoffeeURL: wenn gesetzt, wird im Footer ein "Buy Me a Coffee"-Link angezeigt.
	BuyMeACoffeeURL string
	// LogLevel: "debug" loggt auch GET/HEAD-Requests; "info" (oder leer) nur Mutationen und Fehler.
	LogLevel string
	// ProxyHeader: Header mit der Client-IP hinter einem Reverse Proxy (z. B. "X-Forwarded-For").
	// Wird nur ausgewertet, wenn die Anfrage von einem Proxy aus TrustedProxies kommt.
	ProxyHeader string
	// TrustedProxies: kommagetrennte IPs/CIDRs vertrauenswürdiger Proxies; zusätzlich
	// die Schlüsselwörter "private", "loopback" und "linklocal".
	TrustedProxies string
	// WebAuthnRPID: Domain, an die Passkeys gebunden sind (Standard: Host aus PUBLIC_BASE_URL).
	WebAuthnRPID string
	// WebAuthnOrigins: kommagetrennte Origins, die Passkey-Zeremonien ausführen dürfen
	// (Standard: Origin aus PUBLIC_BASE_URL plus CORS_ALLOW_ORIGINS).
	WebAuthnOrigins string
	// InstanceAdminEmails: kommagetrennte E-Mails, die beim Start das Instanz-Admin-Flag
	// erhalten (nur bestehende Konten; es wird nie etwas entzogen).
	InstanceAdminEmails []string
	// AdminReauthMinutes: Minuten seit der letzten echten Anmeldung (Passwort/Passkey),
	// in denen der Admin-Bereich nutzbar ist (ADMIN_REAUTH_MINUTES), siehe admin_reauth.go.
	AdminReauthMinutes int
}

// Load builds a Config from process environment variables, applying defaults
// that make local development convenient.
func Load() Config {
	cfg := Config{
		Version:                     "", // set in main from ldflags
		AppEnv:                      getEnv("APP_ENV", "development"),
		HTTPPort:                    getEnv("BACKEND_PORT", "8080"),
		BaseURL:                     getEnv("PUBLIC_BASE_URL", "http://localhost:8080"),
		DatabaseURL:                 getEnv("DATABASE_URL", defaultPostgresURL()),
		RedisURL:                    getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:                   getEnv("JWT_SECRET", "dev-secret-change-me"),
		TokenTTLMin:                 getEnvInt("JWT_TTL_MINUTES", 60*24),
		RememberMeDays:              getEnvInt("REMEMBER_ME_DAYS", 30),
		CORSOrigins:                 getEnv("CORS_ALLOW_ORIGINS", "http://localhost:5173"),
		ResendAPIKey:                os.Getenv("RESEND_API_KEY"),
		ResendFromEmail:             getEnv("RESEND_FROM_EMAIL", "noreply@example.com"),
		PasswordResetTokenExpiryMin: getEnvInt("PASSWORD_RESET_TOKEN_EXPIRY_MIN", 60),
		GeminiAPIKey:                os.Getenv("GEMINI_API_KEY"),
		GeminiModel:                 getEnv("GEMINI_MODEL", "gemini-2.5-flash"),
		ChatRateLimitMax:            getEnvInt("CHAT_RATE_LIMIT_MAX", 30),
		ChatRateLimitWindowMin:      getEnvInt("CHAT_RATE_LIMIT_WINDOW_MINUTES", 1),
		AutoMigrate:                 getEnvBool("AUTO_MIGRATE", false),
		RegistrationMode:            strings.ToLower(strings.TrimSpace(getEnv("REGISTRATION_MODE", RegistrationModeOpen))),
		LegalImpressumPath:          os.Getenv("LEGAL_IMPRESSUM_PATH"),
		LegalDatenschutzPath:        os.Getenv("LEGAL_DATENSCHUTZ_PATH"),
		ImpressumHTML:               os.Getenv("IMPRESSUM_HTML"),
		DatenschutzHTML:             os.Getenv("DATENSCHUTZ_HTML"),
		BuyMeACoffeeURL:             strings.TrimSpace(os.Getenv("BUY_ME_A_COFFEE_URL")),
		LogLevel:                    getEnv("LOG_LEVEL", "info"),
		ProxyHeader:                 strings.TrimSpace(os.Getenv("PROXY_HEADER")),
		TrustedProxies:              os.Getenv("TRUSTED_PROXIES"),
		WebAuthnRPID:                strings.TrimSpace(os.Getenv("WEBAUTHN_RP_ID")),
		WebAuthnOrigins:             os.Getenv("WEBAUTHN_ORIGINS"),
		InstanceAdminEmails:         splitList(os.Getenv("INSTANCE_ADMIN_EMAILS")),
		AdminReauthMinutes:          parseAdminReauthMinutes(os.Getenv("ADMIN_REAUTH_MINUTES")),
	}

	return cfg
}

// PasskeyRelyingParty returns the WebAuthn relying party ID and allowed
// origins. Without explicit settings they are derived from PUBLIC_BASE_URL;
// CORS origins are added so the Vite dev server can run ceremonies too.
func (c Config) PasskeyRelyingParty() (string, []string, error) {
	base, err := url.Parse(strings.TrimSpace(c.BaseURL))
	if err != nil || base.Scheme == "" || base.Host == "" {
		return "", nil, fmt.Errorf("PUBLIC_BASE_URL %q is not an absolute URL", c.BaseURL)
	}
	rpID := c.WebAuthnRPID
	if rpID == "" {
		rpID = base.Hostname()
	}

	var origins []string
	if strings.TrimSpace(c.WebAuthnOrigins) != "" {
		origins = splitList(c.WebAuthnOrigins)
	} else {
		origins = append([]string{base.Scheme + "://" + base.Host}, splitList(c.CORSOrigins)...)
	}
	seen := make(map[string]bool, len(origins))
	out := make([]string, 0, len(origins))
	for _, o := range origins {
		o = strings.TrimSuffix(o, "/")
		// Wildcard CORS patterns (e.g. "https://*.example.org") are no WebAuthn origins.
		if o == "" || strings.Contains(o, "*") || seen[o] {
			continue
		}
		seen[o] = true
		out = append(out, o)
	}
	return rpID, out, nil
}

func splitList(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// DebugLogRequests returns true if read-only requests (GET/HEAD) should be logged.
func (c Config) DebugLogRequests() bool {
	return strings.ToLower(c.LogLevel) == "debug"
}

func getEnvBool(key string, fallback bool) bool {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "on", "yes":
		return true
	}
	return false
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	if v, err := strconv.Atoi(value); err == nil {
		return v
	}
	return fallback
}

func defaultPostgresURL() string {
	user := getEnv("POSTGRES_USER", "kegelmaster")
	pass := getEnv("POSTGRES_PASSWORD", "kegelmaster")
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	db := getEnv("POSTGRES_DB", "kegelmaster")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)
}
