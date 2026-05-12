package services_test

import (
"context"
"errors"
"testing"
"time"

"leaderboard-case-study/internal/models"
"leaderboard-case-study/internal/services"
)

func TestUserService_CreateUser_Success(t *testing.T) {
called := false
repo := &mockUserRepo{
createFn: func(_ context.Context, u *models.User) error {
called = true
u.CreatedAt = time.Now()
return nil
},
}
svc := services.NewUserService(repo)
user := &models.User{ID: "u1", Username: "alice"}
if err := svc.CreateUser(context.Background(), user); err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !called {
t.Error("expected repo.CreateUser to be called")
}
if user.CreatedAt.IsZero() {
t.Error("expected CreatedAt to be populated")
}
}

func TestUserService_CreateUser_Error(t *testing.T) {
want := errors.New("constraint violation")
repo := &mockUserRepo{
createFn: func(_ context.Context, _ *models.User) error { return want },
}
svc := services.NewUserService(repo)
err := svc.CreateUser(context.Background(), &models.User{ID: "u2", Username: "bob"})
if !errors.Is(err, want) {
t.Errorf("expected %v, got %v", want, err)
}
}

func TestUserService_GetUser_Success(t *testing.T) {
repo := &mockUserRepo{
getFn: func(_ context.Context, id string) (*models.User, error) {
return &models.User{ID: id, Username: "charlie"}, nil
},
}
svc := services.NewUserService(repo)
got, err := svc.GetUser(context.Background(), "u3")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if got.ID != "u3" || got.Username != "charlie" {
t.Errorf("unexpected user: %+v", got)
}
}

func TestUserService_GetUser_NotFound(t *testing.T) {
want := errors.New("no rows in result set")
repo := &mockUserRepo{
getFn: func(_ context.Context, _ string) (*models.User, error) { return nil, want },
}
svc := services.NewUserService(repo)
_, err := svc.GetUser(context.Background(), "ghost")
if !errors.Is(err, want) {
t.Errorf("expected %v, got %v", want, err)
}
}
