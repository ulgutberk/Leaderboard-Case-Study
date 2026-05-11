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

	// Connect to Postgres (board/user metadata persistence)
	db, err := pgxpool.New(context.Background(), cfg.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
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
	scoreRepo := repositories.NewScoreRepository(redisClient)
	userRepo := repositories.NewUserRepository(db)

	boardService := services.NewBoardService(boardRepo)
	scoreService := services.NewScoreService(scoreRepo)
	userService := services.NewUserService(userRepo)

	boardHandler := handlers.NewBoardHandler(boardService)
	scoreHandler := handlers.NewScoreHandler(scoreService)
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

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
