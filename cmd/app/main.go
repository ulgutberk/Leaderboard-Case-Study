package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

// App holds the main dependencies for the service
// (DB, Redis, Router, etc.)
type App struct {
	Router *mux.Router
	DB     *pgxpool.Pool
	Redis  *redis.Client
}

func main() {
	// Load config from env or use defaults
	pgURL := os.Getenv("POSTGRES_URL")
	if pgURL == "" {
		pgURL = "postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable"
	}
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	// Connect to Postgres
	dbpool, err := pgxpool.New(context.Background(), pgURL)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer dbpool.Close()

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	app := &App{
		Router: mux.NewRouter(),
		DB:     dbpool,
		Redis:  redisClient,
	}

	// Health endpoint (for test)
	app.Router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Açıklama: Tüm client istekleri önce App (Leaderboard Service) üzerinden geçer.
	// App, ilgili işlemi Postgres (kalıcı metadata) veya Redis ZSET (hızlı skor işlemleri) ile yapar.
	// Şu an sadece temel bağlantılar ve bir test endpointi var.

	fmt.Println("Leaderboard Service running on :8080")
	log.Fatal(http.ListenAndServe(":8080", app.Router))
}
