package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	Host            string
	AdminToken      string
	SessionSecret   string
	DataDir         string
	MaxUploadSize   int64
	DBPath          string
}

func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		Port:          getEnv("PORT", "8080"),
		Host:          getEnv("HOST", "0.0.0.0"),
		AdminToken:    getEnv("ADMIN_TOKEN", ""),
		SessionSecret: getEnv("SESSION_SECRET", ""),
		DataDir:       getEnv("DATA_DIR", "./data"),
		DBPath:        getEnv("DB_PATH", "./data/feedback.db"),
	}

	// Parse max upload size
	maxUploadStr := getEnv("MAX_UPLOAD_SIZE", "52428800")
	maxUpload, err := strconv.ParseInt(maxUploadStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_UPLOAD_SIZE: %w", err)
	}
	cfg.MaxUploadSize = maxUpload

	// Validate required fields
	if cfg.AdminToken == "" {
		return nil, fmt.Errorf("ADMIN_TOKEN is required")
	}
	if cfg.SessionSecret == "" {
		return nil, fmt.Errorf("SESSION_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
