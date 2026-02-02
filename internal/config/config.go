package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	TelegramBotToken  string
	DBPath            string
	LogLevel          string
	GLMAPIKey         string
	GLMModelName      string
	GLMBaseURL        string
	GLMTemperature    float64
	GLMFallbackToHF   bool
	HuggingFaceAPIKey string
	HuggingFaceModel  string
	EnableLLMFeatures bool
	EnableCuotas      bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file, but don't fail if it doesn't exist
	_ = godotenv.Load()

	cfg := &Config{
		TelegramBotToken:  getEnv("TELEGRAM_BOT_TOKEN", ""),
		DBPath:            getEnv("DB_PATH", "./data/bot.db"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		GLMAPIKey:         getEnv("GLM_API_KEY", ""),
		GLMModelName:      getEnv("GLM_MODEL_NAME", "glm-4"),
		GLMBaseURL:        getEnv("GLM_BASE_URL", "https://open.bigmodel.cn/api/paas/v4"),
		GLMTemperature:    getEnvFloat("GLM_TEMPERATURE", 0.3),
		GLMFallbackToHF:   getEnvBool("GLM_FALLBACK_TO_HF", true),
		HuggingFaceAPIKey: getEnv("HUGGINGFACE_API_KEY", ""),
		HuggingFaceModel:  getEnv("HUGGINGFACE_MODEL", "mistralai/Mistral-7B-Instruct-v0.2"),
		EnableLLMFeatures: getEnvBool("ENABLE_LLM_FEATURES", true),
		EnableCuotas:      getEnvBool("ENABLE_CUOTAS", true),
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

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

