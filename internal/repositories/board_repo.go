package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"leaderboard-case-study/internal/models"
)

// BoardRepository handles persistent board metadata in Postgres.
type BoardRepository interface {
	CreateBoard(ctx context.Context, board *models.Board) error
	GetBoard(ctx context.Context, id int) (*models.Board, error)
	GetBoardByName(ctx context.Context, name string) (*models.Board, error)
	ListBoards(ctx context.Context) ([]models.BoardSummary, error)
}

type boardRepository struct {
	db *pgxpool.Pool
}

// NewBoardRepository creates a new BoardRepository backed by Postgres.
func NewBoardRepository(db *pgxpool.Pool) BoardRepository {
	return &boardRepository{db: db}
}

// scheduleType extracts the schedule type string (nil-safe).
func scheduleType(s *models.Schedule) *string {
	if s == nil {
		return nil
	}
	return &s.Type
}

// scheduleInterval extracts the interval seconds (nil-safe).
func scheduleInterval(s *models.Schedule) *int {
	if s == nil {
		return nil
	}
	return s.IntervalSeconds
}

// buildSchedule reconstructs a Schedule from nullable DB columns.
func buildSchedule(schedType *string, intervalSecs *int) *models.Schedule {
	if schedType == nil {
		return nil
	}
	return &models.Schedule{
		Type:            *schedType,
		IntervalSeconds: intervalSecs,
	}
}

// CreateBoard inserts a new board record into Postgres.
func (r *boardRepository) CreateBoard(ctx context.Context, board *models.Board) error {
	query := `
		INSERT INTO boards (name, description, schedule_type, schedule_interval_seconds)
		VALUES ($1, $2, $3, $4)
		RETURNING id`
	err := r.db.QueryRow(ctx, query,
		board.Name,
		board.Description,
		scheduleType(board.Schedule),
		scheduleInterval(board.Schedule),
	).Scan(&board.DbID)
	if err != nil {
		return err
	}
	// Expose the DB integer ID as "board_{id}" string
	board.BoardID = fmt.Sprintf("board_%d", board.DbID)
	return nil
}

// GetBoard fetches a board by its integer primary key.
func (r *boardRepository) GetBoard(ctx context.Context, id int) (*models.Board, error) {
	board := &models.Board{}
	var schedType *string
	var intervalSecs *int
	query := `
		SELECT id, name, description, schedule_type, schedule_interval_seconds, created_at
		FROM boards WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).
		Scan(&board.DbID, &board.Name, &board.Description, &schedType, &intervalSecs, &board.CreatedAt)
	if err != nil {
		return nil, err
	}
	board.BoardID = fmt.Sprintf("board_%d", board.DbID)
	board.Schedule = buildSchedule(schedType, intervalSecs)
	return board, nil
}

// ListBoards returns all boards with basic metadata (id + name).
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

// GetBoardByName fetches a board by its unique name.
func (r *boardRepository) GetBoardByName(ctx context.Context, name string) (*models.Board, error) {
	board := &models.Board{}
	var schedType *string
	var intervalSecs *int
	query := `
		SELECT id, name, description, schedule_type, schedule_interval_seconds, created_at
		FROM boards WHERE name = $1`
	err := r.db.QueryRow(ctx, query, name).
		Scan(&board.DbID, &board.Name, &board.Description, &schedType, &intervalSecs, &board.CreatedAt)
	if err != nil {
		return nil, err
	}
	board.BoardID = fmt.Sprintf("board_%d", board.DbID)
	board.Schedule = buildSchedule(schedType, intervalSecs)
	return board, nil
}
