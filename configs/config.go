package configs

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	ServerAddr         string        // HTTP server address (e.g. ":8080")
	PostgresURL        string        // Postgres connection string
	PostgresReadURL    string        // replica — read-only queries
	RedisAddr          string        // Redis address (e.g. "localhost:6379")
	ResetScanInterval  time.Duration // interval between scheduled reset scans
	RateLimitRequests  int           // requests per window
	RateLimitWindow    time.Duration // window size
	RedisSentinelAddrs []string      // Redis Sentinel addresses
	RedisMasterName    string        // Redis Master name
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		ServerAddr:         getEnv("SERVER_ADDR", ":8080"),
		PostgresURL:        getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable"),
		PostgresReadURL:    getEnv("POSTGRES_READ_URL", "postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		ResetScanInterval:  getEnvDuration("RESET_SCAN_INTERVAL", time.Minute),
		RateLimitRequests:  getEnvInt("RATE_LIMIT_REQUESTS", 100), // Rate limit defaults to 100 requests per window
		RateLimitWindow:    getEnvDuration("RATE_LIMIT_WINDOW", time.Minute),
		RedisSentinelAddrs: getEnvList("REDIS_SENTINEL_ADDR"),
		RedisMasterName:    getEnv("REDIS_MASTER_NAME", "mymaster"),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvList(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}

	return values
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
func getEnvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
