package config

import (
	"fmt"
	"os"
	"strconv"
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
	// Legal pages: file path (read at request time) or inline HTML from env.
	// LEGAL_IMPRESSUM_PATH / LEGAL_DATENSCHUTZ_PATH, or IMPRESSUM_HTML / DATENSCHUTZ_HTML.
	LegalImpressumPath   string
	LegalDatenschutzPath string
	ImpressumHTML        string
	DatenschutzHTML      string
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
		LegalImpressumPath:          os.Getenv("LEGAL_IMPRESSUM_PATH"),
		LegalDatenschutzPath:        os.Getenv("LEGAL_DATENSCHUTZ_PATH"),
		ImpressumHTML:               os.Getenv("IMPRESSUM_HTML"),
		DatenschutzHTML:             os.Getenv("DATENSCHUTZ_HTML"),
	}

	return cfg
}

func getEnvBool(key string, fallback bool) bool {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	switch value {
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
