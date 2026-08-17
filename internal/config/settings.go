package config

import (
	"os"
	"strings"
	"time"
)

// Config describes the review room's process boundary.  Environment parsing
// lives here so handlers and adapters receive a fully validated snapshot.
type Config struct {
	Addr        string
	DatabaseURL string
	UploadDir   string
	Timeout     time.Duration
}

func Load() Config {
	return Config{
		Addr:        env("HTTP_ADDR", ":8080"),
		DatabaseURL: env("DATABASE_URL", "postgres://review:review@localhost:5432/review?sslmode=disable"),
		UploadDir:   env("UPLOAD_DIR", "./var/exchange"),
		Timeout:     envDuration("REQUEST_TIMEOUT", 5*time.Second),
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
