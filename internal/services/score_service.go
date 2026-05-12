package services

import (
	"context"
	"time"

	"leaderboard-case-study/internal/models"
	"leaderboard-case-study/internal/repositories"
)

type ScoreService interface {
	SetScore(ctx context.Context, boardID int, userID string, score float64) error
	GetTopScores(ctx context.Context, boardID int, limit int64) ([]models.ScoreEntry, error)
	GetSurroundings(ctx context.Context, boardID int, userID string, n int64) (*models.SurroundingsResponse, error)
	ResetScores(ctx context.Context, boardID int) error
	ResetDueBoards(ctx context.Context) error
}

type scoreService struct {
	repo      repositories.ScoreRepository
	boardRepo repositories.BoardRepository
	now       func() time.Time
}

func NewScoreService(repo repositories.ScoreRepository, boardRepo repositories.BoardRepository) ScoreService {
	return &scoreService{repo: repo, boardRepo: boardRepo, now: time.Now}
}

func (s *scoreService) SetScore(ctx context.Context, boardID int, userID string, score float64) error {
	if err := s.ensureCurrentPeriod(ctx, boardID); err != nil {
		return err
	}
	return s.repo.SetScore(ctx, boardID, userID, score)
}

func (s *scoreService) GetTopScores(ctx context.Context, boardID int, limit int64) ([]models.ScoreEntry, error) {
	if err := s.ensureCurrentPeriod(ctx, boardID); err != nil {
		return nil, err
	}
	return s.repo.GetTopScores(ctx, boardID, limit)
}

func (s *scoreService) GetSurroundings(ctx context.Context, boardID int, userID string, n int64) (*models.SurroundingsResponse, error) {
	if err := s.ensureCurrentPeriod(ctx, boardID); err != nil {
		return nil, err
	}
	return s.repo.GetSurroundings(ctx, boardID, userID, n)
}

func (s *scoreService) ResetScores(ctx context.Context, boardID int) error {
	return s.repo.ResetScores(ctx, boardID)
}

func (s *scoreService) ResetDueBoards(ctx context.Context) error {
	boards, err := s.boardRepo.ListScheduledBoards(ctx)
	if err != nil {
		return err
	}

	now := s.now()
	for i := range boards {
		if err := s.resetBoardIfDue(ctx, &boards[i], now); err != nil {
			return err
		}
	}

	return nil
}

func (s *scoreService) ensureCurrentPeriod(ctx context.Context, boardID int) error {
	board, err := s.boardRepo.GetBoard(ctx, boardID)
	if err != nil {
		return err
	}

	return s.resetBoardIfDue(ctx, board, s.now())
}

func (s *scoreService) resetBoardIfDue(ctx context.Context, board *models.Board, now time.Time) error {
	currentPeriodStart, ok := activePeriodStart(board, now)
	if !ok || !board.LastResetAt.Before(currentPeriodStart) {
		return nil
	}

	if err := s.repo.ResetScores(ctx, board.DbID); err != nil {
		return err
	}

	board.LastResetAt = currentPeriodStart
	return s.boardRepo.UpdateLastResetAt(ctx, board.DbID, currentPeriodStart)
}

func activePeriodStart(board *models.Board, now time.Time) (time.Time, bool) {
	if board.Schedule == nil || board.Schedule.Type != "interval" || board.Schedule.IntervalSeconds == nil || *board.Schedule.IntervalSeconds <= 0 {
		return time.Time{}, false
	}

	interval := time.Duration(*board.Schedule.IntervalSeconds) * time.Second
	if now.Before(board.CreatedAt) {
		return board.CreatedAt, true
	}

	elapsed := now.Sub(board.CreatedAt)
	periods := int64(elapsed / interval)
	return board.CreatedAt.Add(time.Duration(periods) * interval), true
}
