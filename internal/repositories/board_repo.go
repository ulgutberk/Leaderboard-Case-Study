package repositories

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"

	"leaderboard-case-study/internal/models"
)

// BoardRepository defines all data access operations for boards and scores.
// Postgres is used for board metadata; Redis ZSET is used for score operations.
type BoardRepository interface {
	// --- Postgres: board metadata ---
	CreateBoard(ctx context.Context, board *models.Board) error
	GetBoard(ctx context.Context, id int) (*models.Board, error)

	// --- Redis ZSET: score operations ---
	SetScore(ctx context.Context, boardID int, userID string, score float64) error
	GetTopScores(ctx context.Context, boardID int, limit int64) ([]models.Score, error)
	ResetScores(ctx context.Context, boardID int) error
}

// boardRepository is the concrete implementation backed by Postgres + Redis.
type boardRepository struct {
	db    *pgxpool.Pool // Postgres connection pool
	redis *redis.Client // Redis client
}

// NewBoardRepository creates a new BoardRepository.
func NewBoardRepository(db *pgxpool.Pool, redis *redis.Client) BoardRepository {
	return &boardRepository{db: db, redis: redis}
}

// redisKey returns the Redis ZSET key for a given board.
func (r *boardRepository) redisKey(boardID int) string {
	return fmt.Sprintf("leaderboard:board:%d", boardID)
}

// CreateBoard inserts a new board record into Postgres.
func (r *boardRepository) CreateBoard(ctx context.Context, board *models.Board) error {
	query := `
		INSERT INTO boards (name, reset_cron, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, created_at, updated_at`
	return r.db.QueryRow(ctx, query, board.Name, board.ResetCron).
		Scan(&board.ID, &board.CreatedAt, &board.UpdatedAt)
}

// GetBoard fetches a board by ID from Postgres.
func (r *boardRepository) GetBoard(ctx context.Context, id int) (*models.Board, error) {
	board := &models.Board{}
	query := `SELECT id, name, reset_cron, created_at, updated_at FROM boards WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).
		Scan(&board.ID, &board.Name, &board.ResetCron, &board.CreatedAt, &board.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return board, nil
}

// SetScore adds or updates a user's score on a board using Redis ZADD.
func (r *boardRepository) SetScore(ctx context.Context, boardID int, userID string, score float64) error {
	return r.redis.ZAdd(ctx, r.redisKey(boardID), &redis.Z{
		Score:  score,
		Member: userID,
	}).Err()
}

// GetTopScores returns the top N scores for a board (highest first) from Redis ZSET.
func (r *boardRepository) GetTopScores(ctx context.Context, boardID int, limit int64) ([]models.Score, error) {
	results, err := r.redis.ZRevRangeWithScores(ctx, r.redisKey(boardID), 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	scores := make([]models.Score, len(results))
	for i, z := range results {
		scores[i] = models.Score{
			UserID:  z.Member.(string),
			BoardID: boardID,
			Value:   z.Score,
			Rank:    int64(i + 1),
		}
	}
	return scores, nil
}

// ResetScores deletes all scores for a board from Redis ZSET.
func (r *boardRepository) ResetScores(ctx context.Context, boardID int) error {
	return r.redis.Del(ctx, r.redisKey(boardID)).Err()
}
