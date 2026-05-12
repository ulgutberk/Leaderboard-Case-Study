package repositories_test

import (
"context"
"fmt"
"testing"
"time"

"leaderboard-case-study/internal/models"
"leaderboard-case-study/internal/repositories"
)

func TestUserRepository_CreateAndGet(t *testing.T) {
db := newTestDB(t)
repo := repositories.NewUserRepository(db)
ctx := context.Background()

user := &models.User{
ID:       fmt.Sprintf("test-user-%d", time.Now().UnixNano()),
Username: uniqueName("username"),
}

if err := repo.CreateUser(ctx, user); err != nil {
t.Fatalf("CreateUser: %v", err)
}
t.Cleanup(func() { db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID) })

if user.CreatedAt.IsZero() {
t.Error("CreatedAt must be populated after CreateUser")
}

got, err := repo.GetUser(ctx, user.ID)
if err != nil {
t.Fatalf("GetUser: %v", err)
}
if got.ID != user.ID {
t.Errorf("ID: want %q, got %q", user.ID, got.ID)
}
if got.Username != user.Username {
t.Errorf("Username: want %q, got %q", user.Username, got.Username)
}
}

func TestUserRepository_GetNotFound(t *testing.T) {
db := newTestDB(t)
repo := repositories.NewUserRepository(db)
ctx := context.Background()

_, err := repo.GetUser(ctx, "non-existent-user-id-xyz")
if err == nil {
t.Error("expected error for non-existent user, got nil")
}
}

func TestUserRepository_DuplicateID(t *testing.T) {
db := newTestDB(t)
repo := repositories.NewUserRepository(db)
ctx := context.Background()

user := &models.User{
ID:       fmt.Sprintf("test-dup-%d", time.Now().UnixNano()),
Username: uniqueName("dupuser"),
}

if err := repo.CreateUser(ctx, user); err != nil {
t.Fatalf("CreateUser first: %v", err)
}
t.Cleanup(func() { db.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID) })

// Insert same ID again — must fail with a constraint error.
dup := &models.User{ID: user.ID, Username: uniqueName("other")}
if err := repo.CreateUser(ctx, dup); err == nil {
t.Error("expected error on duplicate user ID, got nil")
}
}
