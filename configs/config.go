package configs

import "os"

// Config holds all application configuration loaded from environment variables.
type Config struct {
	ServerAddr  string // HTTP server address (e.g. ":8080")
	PostgresURL string // Postgres connection string
	RedisAddr   string // Redis address (e.g. "localhost:6379")
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		ServerAddr:  getEnv("SERVER_ADDR", ":8080"),
		PostgresURL: getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
