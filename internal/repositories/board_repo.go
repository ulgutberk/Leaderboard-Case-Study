package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"leaderboard-case-study/internal/models"
)

type BoardRepository interface {
	CreateBoard(ctx context.Context, board *models.Board) error
	GetBoard(ctx context.Context, id int) (*models.Board, error)
	GetBoardByName(ctx context.Context, name string) (*models.Board, error)
	ListBoards(ctx context.Context) ([]models.BoardSummary, error)
	ListScheduledBoards(ctx context.Context) ([]models.Board, error)
	UpdateLastResetAt(ctx context.Context, id int, lastResetAt time.Time) error
}

type boardRepository struct {
	db *pgxpool.Pool
}

func NewBoardRepository(db *pgxpool.Pool) BoardRepository {
	return &boardRepository{db: db}
}

func scheduleType(s *models.Schedule) *string {
	if s == nil {
		return nil
	}
	return &s.Type
}

func scheduleInterval(s *models.Schedule) *int {
	if s == nil {
		return nil
	}
	return s.IntervalSeconds
}

func buildSchedule(schedType *string, intervalSecs *int) *models.Schedule {
	if schedType == nil {
		return nil
	}
	return &models.Schedule{
		Type:            *schedType,
		IntervalSeconds: intervalSecs,
	}
}

func (r *boardRepository) CreateBoard(ctx context.Context, board *models.Board) error {
	query := `
		INSERT INTO boards (name, description, schedule_type, schedule_interval_seconds, last_reset_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at, last_reset_at`
	err := r.db.QueryRow(ctx, query,
		board.Name,
		board.Description,
		scheduleType(board.Schedule),
		scheduleInterval(board.Schedule),
	).Scan(&board.DbID, &board.CreatedAt, &board.LastResetAt)
	if err != nil {
		return err
	}
	board.BoardID = fmt.Sprintf("board_%d", board.DbID)
	return nil
}

func (r *boardRepository) GetBoard(ctx context.Context, id int) (*models.Board, error) {
	board := &models.Board{}
	var schedType *string
	var intervalSecs *int
	query := `
		SELECT id, name, description, schedule_type, schedule_interval_seconds, created_at, last_reset_at
		FROM boards WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).
		Scan(&board.DbID, &board.Name, &board.Description, &schedType, &intervalSecs, &board.CreatedAt, &board.LastResetAt)
	if err != nil {
		return nil, err
	}
	board.BoardID = fmt.Sprintf("board_%d", board.DbID)
	board.Schedule = buildSchedule(schedType, intervalSecs)
	return board, nil
}

func (r *boardRepository) ListBoards(ctx context.Context) ([]models.BoardSummary, error) {
	query := `SELECT id, name FROM boards ORDER BY id`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var boards []models.BoardSummary
	for rows.Next() {
		var b models.BoardSummary
		var dbID int
		if err := rows.Scan(&dbID, &b.Name); err != nil {
			return nil, err
		}
		b.BoardID = fmt.Sprintf("board_%d", dbID)
		boards = append(boards, b)
	}
	return boards, rows.Err()
}

func (r *boardRepository) GetBoardByName(ctx context.Context, name string) (*models.Board, error) {
	board := &models.Board{}
	var schedType *string
	var intervalSecs *int
	query := `
		SELECT id, name, description, schedule_type, schedule_interval_seconds, created_at, last_reset_at
		FROM boards WHERE name = $1`
	err := r.db.QueryRow(ctx, query, name).
		Scan(&board.DbID, &board.Name, &board.Description, &schedType, &intervalSecs, &board.CreatedAt, &board.LastResetAt)
	if err != nil {
		return nil, err
	}
	board.BoardID = fmt.Sprintf("board_%d", board.DbID)
	board.Schedule = buildSchedule(schedType, intervalSecs)
	return board, nil
}

func (r *boardRepository) ListScheduledBoards(ctx context.Context) ([]models.Board, error) {
	query := `
		SELECT id, name, description, schedule_type, schedule_interval_seconds, created_at, last_reset_at
		FROM boards
		WHERE schedule_type = 'interval' AND schedule_interval_seconds IS NOT NULL AND schedule_interval_seconds > 0
		ORDER BY id`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boards []models.Board
	for rows.Next() {
		var board models.Board
		var schedType *string
		var intervalSecs *int
		if err := rows.Scan(&board.DbID, &board.Name, &board.Description, &schedType, &intervalSecs, &board.CreatedAt, &board.LastResetAt); err != nil {
			return nil, err
		}
		board.BoardID = fmt.Sprintf("board_%d", board.DbID)
		board.Schedule = buildSchedule(schedType, intervalSecs)
		boards = append(boards, board)
	}

	return boards, rows.Err()
}

func (r *boardRepository) UpdateLastResetAt(ctx context.Context, id int, lastResetAt time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE boards SET last_reset_at = $2, updated_at = NOW() WHERE id = $1`, id, lastResetAt)
	return err
}
