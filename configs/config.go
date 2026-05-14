package configs

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerAddr         string
	PostgresURL        string
	PostgresReadURL    string
	RedisAddr          string
	ResetScanInterval  time.Duration
	RateLimitRequests  int
	RateLimitWindow    time.Duration
	RedisSentinelAddrs []string
	RedisMasterName    string
}

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
