package config

import "os"

type Config struct {
	HTTPAddr    string
	DatabaseURL string
}

func MustLoad() Config {
	cfg := Config{
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@db:5432/org_structure?sslmode=disable"),
	}

	if cfg.DatabaseURL == "" {
		panic("DATABASE_URL is required")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
