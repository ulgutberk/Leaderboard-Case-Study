package services

import (
	"context"

	"leaderboard-case-study/internal/models"
	"leaderboard-case-study/internal/repositories"
)

// UserService defines the business logic for user operations.
type UserService interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, id string) (*models.User, error)
}

type userService struct {
	repo repositories.UserRepository
}

// NewUserService creates a new UserService backed by the given repository.
func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, user *models.User) error {
	return s.repo.CreateUser(ctx, user)
}

func (s *userService) GetUser(ctx context.Context, id string) (*models.User, error) {
	return s.repo.GetUser(ctx, id)
}
