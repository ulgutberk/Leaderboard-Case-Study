// @title           Leaderboard API
// @version         1.0
// @description     REST API for managing leaderboards, users and scores. Requests hit the Leaderboard Service which splits writes/reads between Postgres (boards & users) and Redis ZSET (scores).
// @host            localhost:8080
// @BasePath        /
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"

	httpSwagger "github.com/swaggo/http-swagger"

	"leaderboard-case-study/configs"
	_ "leaderboard-case-study/docs"
	"leaderboard-case-study/internal/handlers"
	"leaderboard-case-study/internal/repositories"
	"leaderboard-case-study/internal/services"
)

func main() {
	// Load configuration from environment variables
	cfg := configs.Load()

	// Connect to Postgres — retry until ready (container may be up before Postgres accepts connections).
	db := mustConnectPostgres(cfg.PostgresURL)
	defer db.Close()

	// Connect to Redis (fast score operations via ZSET)
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Wire up layers:
	//   Repository → Service → Handler
	//
	//   Client → Handler → Service
	//                        ├── BoardRepository  (Postgres)
	//                        ├── ScoreRepository  (Redis ZSET)
	//                        └── UserRepository   (Postgres)

	boardRepo := repositories.NewBoardRepository(db)
	scoreRepo := repositories.NewScoreRepository(db, redisClient)
	userRepo := repositories.NewUserRepository(db)

	boardService := services.NewBoardService(boardRepo)
	scoreService := services.NewScoreService(scoreRepo)
	userService := services.NewUserService(userRepo)

	// Rebuild Redis cache from Postgres on startup (crash recovery / cold start).
	if err := scoreRepo.WarmCache(context.Background()); err != nil {
		log.Printf("warn: Redis cache warm failed (scores may be missing until next SetScore): %v", err)
	} else {
		log.Println("Redis score cache warmed from Postgres")
	}

	boardHandler := handlers.NewBoardHandler(boardService)
	scoreHandler := handlers.NewScoreHandler(scoreService, boardService)
	userHandler := handlers.NewUserHandler(userService)

	// Register routes
	router := mux.NewRouter()
	router.HandleFunc("/health", healthHandler).Methods(http.MethodGet)
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	boardHandler.RegisterRoutes(router)
	scoreHandler.RegisterRoutes(router)
	userHandler.RegisterRoutes(router)

	fmt.Printf("Leaderboard Service running on %s\n", cfg.ServerAddr)
	log.Fatal(http.ListenAndServe(cfg.ServerAddr, router))
}

// mustConnectPostgres retries connecting to Postgres until it succeeds or times out.
// pgxpool.New is lazy (no real connection yet), so we Ping explicitly.
func mustConnectPostgres(url string) *pgxpool.Pool {
	const maxAttempts = 20
	const delay = 3 * time.Second

	for i := 1; i <= maxAttempts; i++ {
		db, err := pgxpool.New(context.Background(), url)
		if err == nil {
			if pingErr := db.Ping(context.Background()); pingErr == nil {
				log.Printf("Postgres connected (attempt %d)", i)
				return db
			} else {
				db.Close()
			}
		}
		log.Printf("Postgres not ready, retrying in %s (attempt %d/%d)...", delay, i, maxAttempts)
		time.Sleep(delay)
	}
	log.Fatalf("Failed to connect to Postgres after %d attempts", maxAttempts)
	return nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
