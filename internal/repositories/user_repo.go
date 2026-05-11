package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"leaderboard-case-study/internal/models"
)

// UserRepository handles persistent user metadata in Postgres.
type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, id string) (*models.User, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository backed by Postgres.
func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

// CreateUser inserts a new user into Postgres.
func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, username, created_at)
		VALUES ($1, $2, NOW())
		RETURNING created_at`
	return r.db.QueryRow(ctx, query, user.ID, user.Username).
		Scan(&user.CreatedAt)
}

// GetUser fetches a user by ID from Postgres.
func (r *userRepository) GetUser(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, created_at FROM users WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).
		Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}
