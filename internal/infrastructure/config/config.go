// Package config loads runtime configuration from the environment with sane
// defaults that match the docker-compose topology.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all process configuration.
type Config struct {
	HTTPAddr          string
	PostgresDSN       string
	RedisAddr         string
	RedisPassword     string
	NatsURL           string
	BaseURL           string
	RelayPollInterval time.Duration
	RelayBatchSize    int
}

// Load reads configuration from environment variables, falling back to defaults.
func Load() *Config {
	return &Config{
		HTTPAddr:          getEnv("HTTP_ADDR", ":8080"),
		PostgresDSN:       getEnv("POSTGRES_DSN", "postgres://admin:admin@postgres:5432/shortener?sslmode=disable"),
		RedisAddr:         getEnv("REDIS_ADDR", "redis:6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		NatsURL:           getEnv("NATS_URL", "nats://nats:4222"),
		BaseURL:           getEnv("BASE_URL", "http://localhost:8080"),
		RelayPollInterval: getEnvDuration("RELAY_POLL_INTERVAL", time.Second),
		RelayBatchSize:    getEnvInt("RELAY_BATCH_SIZE", 100),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
