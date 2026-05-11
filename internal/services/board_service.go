package services

import (
	"context"

	"leaderboard-case-study/internal/models"
	"leaderboard-case-study/internal/repositories"
)

// BoardService defines the business logic for board metadata operations.
type BoardService interface {
	CreateBoard(ctx context.Context, board *models.Board) error
	GetBoard(ctx context.Context, id int) (*models.Board, error)
	GetBoardByName(ctx context.Context, name string) (*models.Board, error)
	ListBoards(ctx context.Context) ([]models.BoardSummary, error)
}

// boardService is the concrete implementation of BoardService.
type boardService struct {
	repo repositories.BoardRepository
}

// NewBoardService creates a new BoardService backed by the given repository.
func NewBoardService(repo repositories.BoardRepository) BoardService {
	return &boardService{repo: repo}
}

func (s *boardService) CreateBoard(ctx context.Context, board *models.Board) error {
	return s.repo.CreateBoard(ctx, board)
}

func (s *boardService) GetBoard(ctx context.Context, id int) (*models.Board, error) {
	return s.repo.GetBoard(ctx, id)
}

func (s *boardService) GetBoardByName(ctx context.Context, name string) (*models.Board, error) {
	return s.repo.GetBoardByName(ctx, name)
}

func (s *boardService) ListBoards(ctx context.Context) ([]models.BoardSummary, error) {
	return s.repo.ListBoards(ctx)
}
