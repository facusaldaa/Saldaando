package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	TelegramBotToken string
	DBPath          string
	LogLevel        string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file, but don't fail if it doesn't exist
	_ = godotenv.Load()

	cfg := &Config{
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		DBPath:          getEnv("DB_PATH", "./data/bot.db"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
	}

	if cfg.TelegramBotToken == "" {
		return nil, ErrMissingBotToken
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

