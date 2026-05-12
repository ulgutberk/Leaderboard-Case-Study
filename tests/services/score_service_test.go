package services_test

import (
"context"
"errors"
"testing"
"time"

"leaderboard-case-study/internal/models"
"leaderboard-case-study/internal/services"
)

// newScoreSvc is a convenience constructor with fixed now() clock.
func newScoreSvc(score *mockScoreRepo, board *mockBoardRepo, now time.Time) services.ScoreService {
svc := services.NewScoreService(score, board)
// Use the exported constructor; inject clock via a thin adapter that wraps now.
// Since the service accepts time.Now internally, we instead rely on test board times
// anchored relative to the real clock. Tests that need deterministic "now" use
// relative offsets baked into the board's CreatedAt / LastResetAt.
_ = now
return svc
}

// ---- ResetDueBoards ----

func TestResetDueBoards_NoBoards(t *testing.T) {
boardRepo := &mockBoardRepo{}
scoreRepo := &mockScoreRepo{}
svc := services.NewScoreService(scoreRepo, boardRepo)

if err := svc.ResetDueBoards(context.Background()); err != nil {
t.Errorf("unexpected error: %v", err)
}
}

func TestResetDueBoards_NoSchedule(t *testing.T) {
called := false
boardRepo := &mockBoardRepo{
listScheduledFn: func(_ context.Context) ([]models.Board, error) {
// ListScheduledBoards normally filters at DB level, but we test
// that a board with nil Schedule is safely skipped anyway.
return []models.Board{{DbID: 1, Name: "no-sched"}}, nil
},
updateLastResetAtFn: func(_ context.Context, _ int, _ time.Time) error {
called = true
return nil
},
}
scoreRepo := &mockScoreRepo{
resetFn: func(_ context.Context, _ int) error {
called = true
return nil
},
}
svc := services.NewScoreService(scoreRepo, boardRepo)
if err := svc.ResetDueBoards(context.Background()); err != nil {
t.Fatalf("unexpected error: %v", err)
}
if called {
t.Error("expected no reset for board without schedule")
}
}

func TestResetDueBoards_SkipsWhenNotDue(t *testing.T) {
// Board created 3 days ago, interval=7 days → not yet due for reset.
interval := 7 * 24 * 3600
createdAt := time.Now().Add(-72 * time.Hour)
board := models.Board{
DbID:        5,
Name:        "not-due",
Schedule:    &models.Schedule{Type: "interval", IntervalSeconds: &interval},
CreatedAt:   createdAt,
LastResetAt: createdAt,
}

resetCalled := false
boardRepo := &mockBoardRepo{
listScheduledFn: func(_ context.Context) ([]models.Board, error) {
return []models.Board{board}, nil
},
}
scoreRepo := &mockScoreRepo{
resetFn: func(_ context.Context, _ int) error {
resetCalled = true
return nil
},
}
svc := services.NewScoreService(scoreRepo, boardRepo)
if err := svc.ResetDueBoards(context.Background()); err != nil {
t.Fatalf("unexpected error: %v", err)
}
if resetCalled {
t.Error("ResetScores must not be called when board is not due")
}
}

func TestResetDueBoards_ResetsWhenDue(t *testing.T) {
// Board created 15 days ago, interval=7 days → 2 full periods elapsed.
// lastResetAt = createdAt → reset is overdue.
interval := 7 * 24 * 3600
createdAt := time.Now().Add(-15 * 24 * time.Hour)
board := models.Board{
DbID:        7,
Name:        "overdue",
Schedule:    &models.Schedule{Type: "interval", IntervalSeconds: &interval},
CreatedAt:   createdAt,
LastResetAt: createdAt,
}

resetCalled := false
updateCalled := false
boardRepo := &mockBoardRepo{
listScheduledFn: func(_ context.Context) ([]models.Board, error) {
return []models.Board{board}, nil
},
updateLastResetAtFn: func(_ context.Context, id int, _ time.Time) error {
updateCalled = true
if id != 7 {
t.Errorf("UpdateLastResetAt: expected id=7, got %d", id)
}
return nil
},
}
scoreRepo := &mockScoreRepo{
resetFn: func(_ context.Context, id int) error {
resetCalled = true
if id != 7 {
t.Errorf("ResetScores: expected boardID=7, got %d", id)
}
return nil
},
}
svc := services.NewScoreService(scoreRepo, boardRepo)
if err := svc.ResetDueBoards(context.Background()); err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !resetCalled {
t.Error("expected ResetScores to be called for overdue board")
}
if !updateCalled {
t.Error("expected UpdateLastResetAt to be called for overdue board")
}
}

// ---- SetScore ----

func TestSetScore_DelegatesToRepo(t *testing.T) {
called := false
scoreRepo := &mockScoreRepo{
setScoreFn: func(_ context.Context, boardID int, userID string, score float64) error {
called = true
if boardID != 3 || userID != "u1" || score != 99.5 {
t.Errorf("unexpected args: boardID=%d userID=%s score=%v", boardID, userID, score)
}
return nil
},
}
// GetBoard returns board with no schedule → no reset triggered.
boardRepo := &mockBoardRepo{
getFn: func(_ context.Context, id int) (*models.Board, error) {
return &models.Board{DbID: id}, nil
},
}
svc := services.NewScoreService(scoreRepo, boardRepo)
if err := svc.SetScore(context.Background(), 3, "u1", 99.5); err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !called {
t.Error("expected repo.SetScore to be called")
}
}

func TestSetScore_PropagatesRepoError(t *testing.T) {
want := errors.New("db down")
scoreRepo := &mockScoreRepo{
setScoreFn: func(_ context.Context, _ int, _ string, _ float64) error { return want },
}
boardRepo := &mockBoardRepo{
getFn: func(_ context.Context, id int) (*models.Board, error) {
return &models.Board{DbID: id}, nil
},
}
svc := services.NewScoreService(scoreRepo, boardRepo)
err := svc.SetScore(context.Background(), 1, "u1", 10)
if !errors.Is(err, want) {
t.Errorf("expected wrapped error %v, got %v", want, err)
}
}

func TestSetScore_PropagatesBoardError(t *testing.T) {
want := errors.New("board not found")
boardRepo := &mockBoardRepo{
getFn: func(_ context.Context, _ int) (*models.Board, error) { return nil, want },
}
svc := services.NewScoreService(&mockScoreRepo{}, boardRepo)
err := svc.SetScore(context.Background(), 1, "u1", 10)
if !errors.Is(err, want) {
t.Errorf("expected %v, got %v", want, err)
}
}

// ---- GetTopScores ----

func TestGetTopScores_ReturnsEntries(t *testing.T) {
want := []models.ScoreEntry{{UserID: "u1", Score: 100}, {UserID: "u2", Score: 50}}
scoreRepo := &mockScoreRepo{
getTopFn: func(_ context.Context, _ int, _ int64) ([]models.ScoreEntry, error) {
return want, nil
},
}
boardRepo := &mockBoardRepo{
getFn: func(_ context.Context, id int) (*models.Board, error) {
return &models.Board{DbID: id}, nil
},
}
svc := services.NewScoreService(scoreRepo, boardRepo)
got, err := svc.GetTopScores(context.Background(), 1, 10)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(got) != len(want) {
t.Fatalf("expected %d entries, got %d", len(want), len(got))
}
for i := range want {
if got[i].UserID != want[i].UserID || got[i].Score != want[i].Score {
t.Errorf("entry %d: want %+v, got %+v", i, want[i], got[i])
}
}
}

func TestGetTopScores_PropagatesError(t *testing.T) {
want := errors.New("redis error")
scoreRepo := &mockScoreRepo{
getTopFn: func(_ context.Context, _ int, _ int64) ([]models.ScoreEntry, error) {
return nil, want
},
}
boardRepo := &mockBoardRepo{
getFn: func(_ context.Context, id int) (*models.Board, error) {
return &models.Board{DbID: id}, nil
},
}
svc := services.NewScoreService(scoreRepo, boardRepo)
_, err := svc.GetTopScores(context.Background(), 1, 10)
if !errors.Is(err, want) {
t.Errorf("expected %v, got %v", want, err)
}
}

// ---- GetSurroundings ----

func TestGetSurroundings_ReturnsResult(t *testing.T) {
want := &models.SurroundingsResponse{
User:  models.ScoreEntry{UserID: "u1", Score: 50},
Above: []models.ScoreEntry{{UserID: "u2", Score: 100}},
Below: []models.ScoreEntry{{UserID: "u3", Score: 10}},
}
scoreRepo := &mockScoreRepo{
getSurroundingsFn: func(_ context.Context, _ int, _ string, _ int64) (*models.SurroundingsResponse, error) {
return want, nil
},
}
boardRepo := &mockBoardRepo{
getFn: func(_ context.Context, id int) (*models.Board, error) {
return &models.Board{DbID: id}, nil
},
}
svc := services.NewScoreService(scoreRepo, boardRepo)
got, err := svc.GetSurroundings(context.Background(), 1, "u1", 3)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if got == nil || got.User.UserID != "u1" {
t.Errorf("unexpected result: %+v", got)
}
}

// ---- ResetScores ----

func TestResetScores_DelegatesToRepo(t *testing.T) {
called := false
scoreRepo := &mockScoreRepo{
resetFn: func(_ context.Context, boardID int) error {
called = true
if boardID != 42 {
t.Errorf("expected boardID=42, got %d", boardID)
}
return nil
},
}
svc := services.NewScoreService(scoreRepo, &mockBoardRepo{})
if err := svc.ResetScores(context.Background(), 42); err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !called {
t.Error("expected repo.ResetScores to be called")
}
}
