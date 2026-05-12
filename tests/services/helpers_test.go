package services_test

import (
"context"
"time"

"leaderboard-case-study/internal/models"
)

// ---- mock BoardRepository ----

type mockBoardRepo struct {
createFn            func(context.Context, *models.Board) error
getFn               func(context.Context, int) (*models.Board, error)
getByNameFn         func(context.Context, string) (*models.Board, error)
listFn              func(context.Context) ([]models.BoardSummary, error)
listScheduledFn     func(context.Context) ([]models.Board, error)
updateLastResetAtFn func(context.Context, int, time.Time) error
}

func (m *mockBoardRepo) CreateBoard(ctx context.Context, b *models.Board) error {
if m.createFn != nil {
return m.createFn(ctx, b)
}
return nil
}
func (m *mockBoardRepo) GetBoard(ctx context.Context, id int) (*models.Board, error) {
if m.getFn != nil {
return m.getFn(ctx, id)
}
return &models.Board{DbID: id}, nil
}
func (m *mockBoardRepo) GetBoardByName(ctx context.Context, name string) (*models.Board, error) {
if m.getByNameFn != nil {
return m.getByNameFn(ctx, name)
}
return &models.Board{Name: name}, nil
}
func (m *mockBoardRepo) ListBoards(ctx context.Context) ([]models.BoardSummary, error) {
if m.listFn != nil {
return m.listFn(ctx)
}
return nil, nil
}
func (m *mockBoardRepo) ListScheduledBoards(ctx context.Context) ([]models.Board, error) {
if m.listScheduledFn != nil {
return m.listScheduledFn(ctx)
}
return nil, nil
}
func (m *mockBoardRepo) UpdateLastResetAt(ctx context.Context, id int, t time.Time) error {
if m.updateLastResetAtFn != nil {
return m.updateLastResetAtFn(ctx, id, t)
}
return nil
}

// ---- mock ScoreRepository ----

type mockScoreRepo struct {
setScoreFn        func(context.Context, int, string, float64) error
getTopFn          func(context.Context, int, int64) ([]models.ScoreEntry, error)
getSurroundingsFn func(context.Context, int, string, int64) (*models.SurroundingsResponse, error)
resetFn           func(context.Context, int) error
warmCacheFn       func(context.Context) error
}

func (m *mockScoreRepo) SetScore(ctx context.Context, boardID int, userID string, score float64) error {
if m.setScoreFn != nil {
return m.setScoreFn(ctx, boardID, userID, score)
}
return nil
}
func (m *mockScoreRepo) GetTopScores(ctx context.Context, boardID int, limit int64) ([]models.ScoreEntry, error) {
if m.getTopFn != nil {
return m.getTopFn(ctx, boardID, limit)
}
return nil, nil
}
func (m *mockScoreRepo) GetSurroundings(ctx context.Context, boardID int, userID string, n int64) (*models.SurroundingsResponse, error) {
if m.getSurroundingsFn != nil {
return m.getSurroundingsFn(ctx, boardID, userID, n)
}
return nil, nil
}
func (m *mockScoreRepo) ResetScores(ctx context.Context, boardID int) error {
if m.resetFn != nil {
return m.resetFn(ctx, boardID)
}
return nil
}
func (m *mockScoreRepo) WarmCache(ctx context.Context) error {
if m.warmCacheFn != nil {
return m.warmCacheFn(ctx)
}
return nil
}

// ---- mock UserRepository ----

type mockUserRepo struct {
createFn func(context.Context, *models.User) error
getFn    func(context.Context, string) (*models.User, error)
}

func (m *mockUserRepo) CreateUser(ctx context.Context, u *models.User) error {
if m.createFn != nil {
return m.createFn(ctx, u)
}
return nil
}
func (m *mockUserRepo) GetUser(ctx context.Context, id string) (*models.User, error) {
if m.getFn != nil {
return m.getFn(ctx, id)
}
return &models.User{ID: id}, nil
}
