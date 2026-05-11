package services

import (
	"context"

	"leaderboard-case-study/internal/models"
	"leaderboard-case-study/internal/repositories"
)

// ScoreService defines the business logic for score operations.
type ScoreService interface {
	SetScore(ctx context.Context, boardID int, userID string, score float64) error
	GetTopScores(ctx context.Context, boardID int, limit int64) ([]models.Score, error)
	ResetScores(ctx context.Context, boardID int) error
}

type scoreService struct {
	repo repositories.ScoreRepository
}

// NewScoreService creates a new ScoreService backed by the given repository.
func NewScoreService(repo repositories.ScoreRepository) ScoreService {
	return &scoreService{repo: repo}
}

func (s *scoreService) SetScore(ctx context.Context, boardID int, userID string, score float64) error {
	return s.repo.SetScore(ctx, boardID, userID, score)
}

func (s *scoreService) GetTopScores(ctx context.Context, boardID int, limit int64) ([]models.Score, error) {
	return s.repo.GetTopScores(ctx, boardID, limit)
}

func (s *scoreService) ResetScores(ctx context.Context, boardID int) error {
	return s.repo.ResetScores(ctx, boardID)
}
