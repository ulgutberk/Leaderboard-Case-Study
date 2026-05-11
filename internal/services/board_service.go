package services

import (
	"context"

	"leaderboard-case-study/internal/models"
	"leaderboard-case-study/internal/repositories"
)

// BoardService defines the business logic for leaderboard operations.
type BoardService interface {
	CreateBoard(ctx context.Context, board *models.Board) error
	GetBoard(ctx context.Context, id int) (*models.Board, error)
	SetScore(ctx context.Context, boardID int, userID string, score float64) error
	GetTopScores(ctx context.Context, boardID int, limit int64) ([]models.Score, error)
	ResetScores(ctx context.Context, boardID int) error
}

// boardService is the concrete implementation of BoardService.
type boardService struct {
	repo repositories.BoardRepository
}

// NewBoardService creates a new BoardService with the given repository.
func NewBoardService(repo repositories.BoardRepository) BoardService {
	return &boardService{repo: repo}
}

func (s *boardService) CreateBoard(ctx context.Context, board *models.Board) error {
	return s.repo.CreateBoard(ctx, board)
}

func (s *boardService) GetBoard(ctx context.Context, id int) (*models.Board, error) {
	return s.repo.GetBoard(ctx, id)
}

func (s *boardService) SetScore(ctx context.Context, boardID int, userID string, score float64) error {
	return s.repo.SetScore(ctx, boardID, userID, score)
}

func (s *boardService) GetTopScores(ctx context.Context, boardID int, limit int64) ([]models.Score, error) {
	return s.repo.GetTopScores(ctx, boardID, limit)
}

func (s *boardService) ResetScores(ctx context.Context, boardID int) error {
	return s.repo.ResetScores(ctx, boardID)
}
