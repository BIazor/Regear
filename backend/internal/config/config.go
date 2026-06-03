package config

import (
	"os"
	"strings"
)

type Config struct {
	Port         string
	DatabasePath string
	CORSOrigins  []string
}
func Load() Config {
	databasePath := env("DATABASE_URL", "")
	if databasePath == "" {
		databasePath = env("DATABASE_PATH", "regears.db")
	}

	return Config{
		Port:         env("PORT", "8080"),
		DatabasePath: databasePath,
		CORSOrigins: split(env(
			"CORS_ORIGINS",
			"http://localhost:5173,http://127.0.0.1:5173,http://localhost:5174,http://127.0.0.1:5174",
		)),
	}
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func split(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
