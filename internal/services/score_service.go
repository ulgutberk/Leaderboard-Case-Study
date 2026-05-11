package services

import (
    "context"

    "leaderboard-case-study/internal/models"
    "leaderboard-case-study/internal/repositories"
)

type ScoreService interface {
    SetScore(ctx context.Context, boardID int, userID string, score float64) error
    GetTopScores(ctx context.Context, boardID int, limit int64) ([]models.ScoreEntry, error)
    GetSurroundings(ctx context.Context, boardID int, userID string, n int64) (*models.SurroundingsResponse, error)
    ResetScores(ctx context.Context, boardID int) error
}

type scoreService struct {
    repo repositories.ScoreRepository
}

func NewScoreService(repo repositories.ScoreRepository) ScoreService {
    return &scoreService{repo: repo}
}

func (s *scoreService) SetScore(ctx context.Context, boardID int, userID string, score float64) error {
    return s.repo.SetScore(ctx, boardID, userID, score)
}

func (s *scoreService) GetTopScores(ctx context.Context, boardID int, limit int64) ([]models.ScoreEntry, error) {
    return s.repo.GetTopScores(ctx, boardID, limit)
}

func (s *scoreService) GetSurroundings(ctx context.Context, boardID int, userID string, n int64) (*models.SurroundingsResponse, error) {
    return s.repo.GetSurroundings(ctx, boardID, userID, n)
}

func (s *scoreService) ResetScores(ctx context.Context, boardID int) error {
    return s.repo.ResetScores(ctx, boardID)
}