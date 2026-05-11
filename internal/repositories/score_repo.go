package repositories

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"

	"leaderboard-case-study/internal/models"
)

// ScoreRepository handles score operations.
// Redis ZSET is used for fast ranked reads; Postgres is the persistent source of truth.
type ScoreRepository interface {
	// SetScore upserts a score in both Postgres and Redis ZSET.
	SetScore(ctx context.Context, boardID int, userID string, score float64) error
	// GetTopScores returns the top N scores for a board (highest first) from Redis ZSET.
	GetTopScores(ctx context.Context, boardID int, limit int64) ([]models.Score, error)
	// ResetScores clears all scores for a board from both Redis and Postgres.
	ResetScores(ctx context.Context, boardID int) error
}

type scoreRepository struct {
	redis *redis.Client
}

// NewScoreRepository creates a new ScoreRepository backed by Redis.
func NewScoreRepository(redis *redis.Client) ScoreRepository {
	return &scoreRepository{redis: redis}
}

// redisKey returns the Redis ZSET key for a given board.
func (r *scoreRepository) redisKey(boardID int) string {
	return fmt.Sprintf("leaderboard:board:%d", boardID)
}

// SetScore adds or updates a user's score on a board using Redis ZADD.
func (r *scoreRepository) SetScore(ctx context.Context, boardID int, userID string, score float64) error {
	return r.redis.ZAdd(ctx, r.redisKey(boardID), &redis.Z{
		Score:  score,
		Member: userID,
	}).Err()
}

// GetTopScores returns the top N scores for a board (highest first) from Redis ZSET.
func (r *scoreRepository) GetTopScores(ctx context.Context, boardID int, limit int64) ([]models.Score, error) {
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
func (r *scoreRepository) ResetScores(ctx context.Context, boardID int) error {
	return r.redis.Del(ctx, r.redisKey(boardID)).Err()
}
