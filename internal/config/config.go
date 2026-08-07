// Package config loads runtime configuration from the environment.
package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	RedisHost         string
	RedisPort         string
	RedisUser         string
	RedisPassword     string
	OllamaBaseURL     string
	OllamaModel       string
	HTTPAddr          string
	MaxToolIterations int
}

// Load reads configuration from the environment (a .env file first, if present).
// The REDIS_* values are op:// references the vault package resolves from
// 1Password; they get no fallback default so a missing one fails fast rather
// than hiding behind a baked-in value. MAX_TOOL_ITERATIONS is a tuning knob, not
// a secret, so it keeps a default.
func Load() Config {
	_ = godotenv.Load()

	return Config{
		RedisHost:         os.Getenv("REDIS_HOST"),
		RedisPort:         os.Getenv("REDIS_PORT"),
		RedisUser:         os.Getenv("REDIS_USER"),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		OllamaBaseURL:     os.Getenv("OLLAMA_BASE_URL"),
		OllamaModel:       os.Getenv("OLLAMA_MODEL"),
		HTTPAddr:          os.Getenv("HTTP_ADDR"),
		MaxToolIterations: getInt("MAX_TOOL_ITERATIONS", 12),
	}
}

func getInt(key string, fallback int) int {
	n, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return n
}
