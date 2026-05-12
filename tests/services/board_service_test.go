package services_test

import (
"context"
"errors"
"testing"

"leaderboard-case-study/internal/models"
"leaderboard-case-study/internal/services"
)

func TestBoardService_CreateBoard_Success(t *testing.T) {
called := false
repo := &mockBoardRepo{
createFn: func(_ context.Context, b *models.Board) error {
called = true
b.DbID = 10
b.BoardID = "board_10"
return nil
},
}
svc := services.NewBoardService(repo)
board := &models.Board{Name: "test"}
if err := svc.CreateBoard(context.Background(), board); err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !called {
t.Error("expected repo.CreateBoard to be called")
}
if board.DbID != 10 {
t.Errorf("expected DbID=10, got %d", board.DbID)
}
}

func TestBoardService_CreateBoard_Error(t *testing.T) {
want := errors.New("duplicate name")
repo := &mockBoardRepo{
createFn: func(_ context.Context, _ *models.Board) error { return want },
}
svc := services.NewBoardService(repo)
err := svc.CreateBoard(context.Background(), &models.Board{Name: "dup"})
if !errors.Is(err, want) {
t.Errorf("expected %v, got %v", want, err)
}
}

func TestBoardService_GetBoard(t *testing.T) {
repo := &mockBoardRepo{
getFn: func(_ context.Context, id int) (*models.Board, error) {
return &models.Board{DbID: id, Name: "found"}, nil
},
}
svc := services.NewBoardService(repo)
got, err := svc.GetBoard(context.Background(), 5)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if got.DbID != 5 || got.Name != "found" {
t.Errorf("unexpected board: %+v", got)
}
}

func TestBoardService_GetBoard_NotFound(t *testing.T) {
want := errors.New("not found")
repo := &mockBoardRepo{
getFn: func(_ context.Context, _ int) (*models.Board, error) { return nil, want },
}
svc := services.NewBoardService(repo)
_, err := svc.GetBoard(context.Background(), 99)
if !errors.Is(err, want) {
t.Errorf("expected %v, got %v", want, err)
}
}

func TestBoardService_GetBoardByName(t *testing.T) {
repo := &mockBoardRepo{
getByNameFn: func(_ context.Context, name string) (*models.Board, error) {
return &models.Board{Name: name}, nil
},
}
svc := services.NewBoardService(repo)
got, err := svc.GetBoardByName(context.Background(), "weekly")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if got.Name != "weekly" {
t.Errorf("expected name 'weekly', got %q", got.Name)
}
}

func TestBoardService_ListBoards(t *testing.T) {
want := []models.BoardSummary{{BoardID: "board_1", Name: "a"}, {BoardID: "board_2", Name: "b"}}
repo := &mockBoardRepo{
listFn: func(_ context.Context) ([]models.BoardSummary, error) { return want, nil },
}
svc := services.NewBoardService(repo)
got, err := svc.ListBoards(context.Background())
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(got) != 2 {
t.Fatalf("expected 2 boards, got %d", len(got))
}
for i := range want {
if got[i].Name != want[i].Name {
t.Errorf("board %d: expected name %q, got %q", i, want[i].Name, got[i].Name)
}
}
}
