package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"leaderboard-case-study/internal/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, id string) (*models.User, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // no-op if already committed, if any error occurs the transaction will be rolled back.

	query := `
		INSERT INTO users (id, username, created_at)
		VALUES ($1, $2, NOW())
		RETURNING created_at`
	if err := tx.QueryRow(ctx, query, user.ID, user.Username).Scan(&user.CreatedAt); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

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
