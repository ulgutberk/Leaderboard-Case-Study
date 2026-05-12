package handlers_test

import (
"bytes"
"context"
"encoding/json"
"errors"
"net/http"
"net/http/httptest"
"testing"

"github.com/gorilla/mux"

"leaderboard-case-study/internal/handlers"
"leaderboard-case-study/internal/models"
)

// ── Mock: ScoreService ────────────────────────────────────────────────────────

type mockScoreService struct {
setScoreFn       func(ctx context.Context, boardID int, userID string, score float64) error
getTopScoresFn   func(ctx context.Context, boardID int, limit int64) ([]models.ScoreEntry, error)
getAroundFn      func(ctx context.Context, boardID int, userID string, n int64) (*models.SurroundingsResponse, error)
resetScoresFn    func(ctx context.Context, boardID int) error
resetDueBoardsFn func(ctx context.Context) error
}

func (m *mockScoreService) SetScore(ctx context.Context, boardID int, userID string, score float64) error {
return m.setScoreFn(ctx, boardID, userID, score)
}
func (m *mockScoreService) GetTopScores(ctx context.Context, boardID int, limit int64) ([]models.ScoreEntry, error) {
return m.getTopScoresFn(ctx, boardID, limit)
}
func (m *mockScoreService) GetSurroundings(ctx context.Context, boardID int, userID string, n int64) (*models.SurroundingsResponse, error) {
return m.getAroundFn(ctx, boardID, userID, n)
}
func (m *mockScoreService) ResetScores(ctx context.Context, boardID int) error {
return m.resetScoresFn(ctx, boardID)
}
func (m *mockScoreService) ResetDueBoards(ctx context.Context) error {
return m.resetDueBoardsFn(ctx)
}

// ── Mock: UserService ─────────────────────────────────────────────────────────

type mockUserService struct {
createFn func(ctx context.Context, user *models.User) error
getFn    func(ctx context.Context, id string) (*models.User, error)
}

func (m *mockUserService) CreateUser(ctx context.Context, user *models.User) error {
return m.createFn(ctx, user)
}
func (m *mockUserService) GetUser(ctx context.Context, id string) (*models.User, error) {
return m.getFn(ctx, id)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func newScoreHandler(boardSvc *mockBoardService, scoreSvc *mockScoreService, userSvc *mockUserService) *handlers.ScoreHandler {
return handlers.NewScoreHandler(scoreSvc, boardSvc, userSvc)
}

func executeScoreRequest(t *testing.T, handler *handlers.ScoreHandler, req *http.Request) *httptest.ResponseRecorder {
t.Helper()
router := mux.NewRouter()
handler.RegisterRoutes(router)
rr := httptest.NewRecorder()
router.ServeHTTP(rr, req)
return rr
}

func nopScoreSvc() *mockScoreService {
return &mockScoreService{
setScoreFn:       func(_ context.Context, _ int, _ string, _ float64) error { return nil },
getTopScoresFn:   func(_ context.Context, _ int, _ int64) ([]models.ScoreEntry, error) { return nil, nil },
getAroundFn:      func(_ context.Context, _ int, _ string, _ int64) (*models.SurroundingsResponse, error) { return nil, nil },
resetScoresFn:    func(_ context.Context, _ int) error { return nil },
resetDueBoardsFn: func(_ context.Context) error { return nil },
}
}

func nopUserSvc() *mockUserService {
return &mockUserService{
createFn: func(_ context.Context, _ *models.User) error { return nil },
getFn:    func(_ context.Context, _ string) (*models.User, error) { return nil, nil },
}
}

// ── PopulateMockScores ────────────────────────────────────────────────────────

func TestPopulateMockScores_Success(t *testing.T) {
boardSvc := nopSvc()
boardSvc.getFn = func(_ context.Context, id int) (*models.Board, error) {
return &models.Board{DbID: id, BoardID: "board_12", Name: "Weekly"}, nil
}
createdUsers := make([]string, 0, 3)
userSvc := &mockUserService{
createFn: func(_ context.Context, user *models.User) error {
createdUsers = append(createdUsers, user.ID)
return nil
},
getFn: func(_ context.Context, _ string) (*models.User, error) { return nil, nil },
}
setCalls := 0
scoreSvc := &mockScoreService{
setScoreFn: func(_ context.Context, boardID int, userID string, score float64) error {
setCalls++
if boardID != 12 {
t.Fatalf("expected boardID 12, got %d", boardID)
}
if userID == "" {
t.Fatal("expected generated userID")
}
if score < 100 || score > 10000 {
t.Fatalf("expected score range 100-10000, got %.2f", score)
}
return nil
},
getTopScoresFn:   func(_ context.Context, _ int, _ int64) ([]models.ScoreEntry, error) { return nil, nil },
getAroundFn:      func(_ context.Context, _ int, _ string, _ int64) (*models.SurroundingsResponse, error) { return nil, nil },
resetScoresFn:    func(_ context.Context, _ int) error { return nil },
resetDueBoardsFn: func(_ context.Context) error { return nil },
}
handler := newScoreHandler(boardSvc, scoreSvc, userSvc)
data, _ := json.Marshal(map[string]any{"count": 3})
req, _ := http.NewRequest(http.MethodPost, "/boards/board_12/mock-scores", bytes.NewReader(data))
req.Header.Set("Content-Type", "application/json")
rr := executeScoreRequest(t, handler, req)
if rr.Code != http.StatusCreated {
t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
}
if len(createdUsers) != 3 {
t.Fatalf("expected 3 created users, got %d", len(createdUsers))
}
if setCalls != 3 {
t.Fatalf("expected 3 score writes, got %d", setCalls)
}
var resp models.PopulateMockScoresResponse
if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
t.Fatalf("decode response: %v", err)
}
if resp.BoardID != "board_12" {
t.Fatalf("expected board_12, got %s", resp.BoardID)
}
if resp.Count != 3 {
t.Fatalf("expected count 3, got %d", resp.Count)
}
if len(resp.Scores) != 3 {
t.Fatalf("expected 3 scores, got %d", len(resp.Scores))
}
}

func TestPopulateMockScores_RejectsInvalidCount(t *testing.T) {
boardSvc := nopSvc()
boardSvc.getFn = func(_ context.Context, id int) (*models.Board, error) {
return &models.Board{DbID: id, BoardID: "board_12", Name: "Weekly"}, nil
}
userSvc := &mockUserService{
createFn: func(_ context.Context, _ *models.User) error {
t.Fatal("CreateUser should not be called for invalid count")
return nil
},
getFn: func(_ context.Context, _ string) (*models.User, error) { return nil, nil },
}
handler := newScoreHandler(boardSvc, nopScoreSvc(), userSvc)
req, _ := http.NewRequest(http.MethodPost, "/boards/board_12/mock-scores", bytes.NewBufferString(`{"count":0}`))
req.Header.Set("Content-Type", "application/json")
rr := executeScoreRequest(t, handler, req)
if rr.Code != http.StatusBadRequest {
t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
}
}

// ── SetScore ──────────────────────────────────────────────────────────────────

func TestSetScore_Success(t *testing.T) {
boardSvc := nopSvc()
boardSvc.getFn = func(_ context.Context, id int) (*models.Board, error) {
return &models.Board{DbID: id, BoardID: "board_1"}, nil
}
scoreSvc := nopScoreSvc()
scoreSvc.setScoreFn = func(_ context.Context, boardID int, userID string, score float64) error {
if boardID != 1 {
t.Errorf("expected boardID 1, got %d", boardID)
}
if userID != "user_789" {
t.Errorf("expected userID user_789, got %s", userID)
}
if score != 1500 {
t.Errorf("expected score 1500, got %f", score)
}
return nil
}
h := newScoreHandler(boardSvc, scoreSvc, nopUserSvc())
body, _ := json.Marshal(map[string]any{"userId": "user_789", "score": 1500})
req, _ := http.NewRequest(http.MethodPost, "/boards/board_1/scores", bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
rr := executeScoreRequest(t, h, req)
if rr.Code != http.StatusOK {
t.Fatalf("expected 200, got %d — body: %s", rr.Code, rr.Body.String())
}
var resp models.SetScoreResponse
if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
t.Fatalf("decode response: %v", err)
}
if resp.BoardID != "board_1" {
t.Errorf("expected boardId board_1, got %s", resp.BoardID)
}
if resp.UserID != "user_789" {
t.Errorf("expected userId user_789, got %s", resp.UserID)
}
if resp.Score != 1500 {
t.Errorf("expected score 1500, got %f", resp.Score)
}
}

func TestSetScore_BoardNotFound(t *testing.T) {
boardSvc := nopSvc()
boardSvc.getFn = func(_ context.Context, _ int) (*models.Board, error) {
return nil, errors.New("not found")
}
h := newScoreHandler(boardSvc, nopScoreSvc(), nopUserSvc())
body, _ := json.Marshal(map[string]any{"userId": "user_789", "score": 1500})
req, _ := http.NewRequest(http.MethodPost, "/boards/board_999/scores", bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
rr := executeScoreRequest(t, h, req)
if rr.Code != http.StatusNotFound {
t.Fatalf("expected 404, got %d — body: %s", rr.Code, rr.Body.String())
}
var resp map[string]string
json.NewDecoder(rr.Body).Decode(&resp)
if resp["error"] != "Board not found" {
t.Errorf(`expected "Board not found", got %q`, resp["error"])
}
}

func TestSetScore_MissingUserId(t *testing.T) {
boardSvc := nopSvc()
boardSvc.getFn = func(_ context.Context, id int) (*models.Board, error) {
return &models.Board{DbID: id, BoardID: "board_1"}, nil
}
h := newScoreHandler(boardSvc, nopScoreSvc(), nopUserSvc())
body, _ := json.Marshal(map[string]any{"score": 1500})
req, _ := http.NewRequest(http.MethodPost, "/boards/board_1/scores", bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")
rr := executeScoreRequest(t, h, req)
if rr.Code != http.StatusBadRequest {
t.Fatalf("expected 400, got %d — body: %s", rr.Code, rr.Body.String())
}
}

func TestSetScore_InvalidJSON(t *testing.T) {
boardSvc := nopSvc()
boardSvc.getFn = func(_ context.Context, id int) (*models.Board, error) {
return &models.Board{DbID: id, BoardID: "board_1"}, nil
}
h := newScoreHandler(boardSvc, nopScoreSvc(), nopUserSvc())
req, _ := http.NewRequest(http.MethodPost, "/boards/board_1/scores", bytes.NewBufferString(`{invalid`))
req.Header.Set("Content-Type", "application/json")
rr := executeScoreRequest(t, h, req)
if rr.Code != http.StatusBadRequest {
t.Fatalf("expected 400, got %d", rr.Code)
}
}

// ── GetTopScores ──────────────────────────────────────────────────────────────

func TestGetTopScores_Success(t *testing.T) {
boardSvc := nopSvc()
boardSvc.getFn = func(_ context.Context, id int) (*models.Board, error) {
return &models.Board{DbID: id, BoardID: "board_1"}, nil
}
scoreSvc := nopScoreSvc()
scoreSvc.getTopScoresFn = func(_ context.Context, _ int, limit int64) ([]models.ScoreEntry, error) {
if limit != 3 {
t.Errorf("expected limit 3, got %d", limit)
}
return []models.ScoreEntry{
{UserID: "user_1", Score: 5000},
{UserID: "user_2", Score: 3000},
{UserID: "user_3", Score: 1500},
}, nil
}
h := newScoreHandler(boardSvc, scoreSvc, nopUserSvc())
req, _ := http.NewRequest(http.MethodGet, "/boards/board_1/scores?n=3", nil)
rr := executeScoreRequest(t, h, req)
if rr.Code != http.StatusOK {
t.Fatalf("expected 200, got %d — body: %s", rr.Code, rr.Body.String())
}
var scores []models.ScoreEntry
if err := json.NewDecoder(rr.Body).Decode(&scores); err != nil {
t.Fatalf("decode response: %v", err)
}
if len(scores) != 3 {
t.Fatalf("expected 3 scores, got %d", len(scores))
}
if scores[0].Score < scores[1].Score || scores[1].Score < scores[2].Score {
t.Errorf("expected descending order, got %v", scores)
}
}

func TestGetTopScores_EmptyBoard(t *testing.T) {
boardSvc := nopSvc()
boardSvc.getFn = func(_ context.Context, id int) (*models.Board, error) {
return &models.Board{DbID: id, BoardID: "board_1"}, nil
}
scoreSvc := nopScoreSvc()
scoreSvc.getTopScoresFn = func(_ context.Context, _ int, _ int64) ([]models.ScoreEntry, error) {
return nil, nil
}
h := newScoreHandler(boardSvc, scoreSvc, nopUserSvc())
req, _ := http.NewRequest(http.MethodGet, "/boards/board_1/scores", nil)
rr := executeScoreRequest(t, h, req)
if rr.Code != http.StatusOK {
t.Fatalf("expected 200, got %d", rr.Code)
}
}

func TestGetTopScores_BoardNotFound(t *testing.T) {
boardSvc := nopSvc()
boardSvc.getFn = func(_ context.Context, _ int) (*models.Board, error) {
return nil, errors.New("not found")
}
h := newScoreHandler(boardSvc, nopScoreSvc(), nopUserSvc())
req, _ := http.NewRequest(http.MethodGet, "/boards/board_999/scores?n=10", nil)
rr := executeScoreRequest(t, h, req)
if rr.Code != http.StatusNotFound {
t.Fatalf("expected 404, got %d — body: %s", rr.Code, rr.Body.String())
}
var resp map[string]string
json.NewDecoder(rr.Body).Decode(&resp)
if resp["error"] != "Board not found" {
t.Errorf(`expected "Board not found", got %q`, resp["error"])
}
}

func TestGetTopScores_InvalidN(t *testing.T) {
boardSvc := nopSvc()
boardSvc.getFn = func(_ context.Context, id int) (*models.Board, error) {
return &models.Board{DbID: id, BoardID: "board_1"}, nil
}
h := newScoreHandler(boardSvc, nopScoreSvc(), nopUserSvc())
for _, qp := range []string{"n=0", "n=-5", "n=abc"} {
req, _ := http.NewRequest(http.MethodGet, "/boards/board_1/scores?"+qp, nil)
rr := executeScoreRequest(t, h, req)
if rr.Code != http.StatusBadRequest {
t.Errorf("[%s] expected 400, got %d — body: %s", qp, rr.Code, rr.Body.String())
}
var resp map[string]string
json.NewDecoder(rr.Body).Decode(&resp)
if resp["error"] != "Invalid value for n" {
t.Errorf("[%s] expected 'Invalid value for n', got %q", qp, resp["error"])
}
}
}

func TestGetTopScores_DefaultsToTen(t *testing.T) {
boardSvc := nopSvc()
boardSvc.getFn = func(_ context.Context, id int) (*models.Board, error) {
return &models.Board{DbID: id, BoardID: "board_1"}, nil
}
scoreSvc := nopScoreSvc()
var capturedLimit int64
scoreSvc.getTopScoresFn = func(_ context.Context, _ int, limit int64) ([]models.ScoreEntry, error) {
capturedLimit = limit
return []models.ScoreEntry{}, nil
}
h := newScoreHandler(boardSvc, scoreSvc, nopUserSvc())
req, _ := http.NewRequest(http.MethodGet, "/boards/board_1/scores", nil)
executeScoreRequest(t, h, req)
if capturedLimit != 10 {
t.Errorf("expected default limit 10, got %d", capturedLimit)
}
}

// ── GetSurroundings ───────────────────────────────────────────────────────────

func TestGetSurroundings_Success(t *testing.T) {
boardSvc := nopSvc()
boardSvc.getFn = func(_ context.Context, id int) (*models.Board, error) {
return &models.Board{DbID: id, BoardID: "board_1"}, nil
}
scoreSvc := nopScoreSvc()
scoreSvc.getAroundFn = func(_ context.Context, _ int, userID string, n int64) (*models.SurroundingsResponse, error) {
if userID != "user_789" {
t.Errorf("expected userId user_789, got %s", userID)
}
if n != 5 {
t.Errorf("expected n=5, got %d", n)
}
return &models.SurroundingsResponse{
User:  models.ScoreEntry{UserID: "user_789", Score: 1500},
Above: []models.ScoreEntry{{UserID: "user_above", Score: 1600}},
Below: []models.ScoreEntry{{UserID: "user_below", Score: 1400}},
}, nil
}
h := newScoreHandler(boardSvc, scoreSvc, nopUserSvc())
req, _ := http.NewRequest(http.MethodGet, "/boards/board_1/scores/user_789/surroundings?n=5", nil)
rr := executeScoreRequest(t, h, req)
if rr.Code != http.StatusOK {
t.Fatalf("expected 200, got %d — body: %s", rr.Code, rr.Body.String())
}
var resp models.SurroundingsResponse
if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
t.Fatalf("decode response: %v", err)
}
if resp.User.UserID != "user_789" {
t.Errorf("expected user user_789, got %s", resp.User.UserID)
}
if len(resp.Above) != 1 || len(resp.Below) != 1 {
t.Errorf("expected 1 above and 1 below, got above=%d below=%d", len(resp.Above), len(resp.Below))
}
}

func TestGetSurroundings_BoardNotFound(t *testing.T) {
boardSvc := nopSvc()
boardSvc.getFn = func(_ context.Context, _ int) (*models.Board, error) {
return nil, errors.New("not found")
}
h := newScoreHandler(boardSvc, nopScoreSvc(), nopUserSvc())
req, _ := http.NewRequest(http.MethodGet, "/boards/board_999/scores/user_789/surroundings", nil)
rr := executeScoreRequest(t, h, req)
if rr.Code != http.StatusNotFound {
t.Fatalf("expected 404, got %d — body: %s", rr.Code, rr.Body.String())
}
var resp map[string]string
json.NewDecoder(rr.Body).Decode(&resp)
if resp["error"] != "Board not found" {
t.Errorf(`expected "Board not found", got %q`, resp["error"])
}
}

func TestGetSurroundings_UserNotFound(t *testing.T) {
boardSvc := nopSvc()
boardSvc.getFn = func(_ context.Context, id int) (*models.Board, error) {
return &models.Board{DbID: id, BoardID: "board_1"}, nil
}
scoreSvc := nopScoreSvc()
scoreSvc.getAroundFn = func(_ context.Context, _ int, _ string, _ int64) (*models.SurroundingsResponse, error) {
return nil, nil
}
h := newScoreHandler(boardSvc, scoreSvc, nopUserSvc())
req, _ := http.NewRequest(http.MethodGet, "/boards/board_1/scores/unknown_user/surroundings", nil)
rr := executeScoreRequest(t, h, req)
if rr.Code != http.StatusNotFound {
t.Fatalf("expected 404, got %d — body: %s", rr.Code, rr.Body.String())
}
var resp map[string]string
json.NewDecoder(rr.Body).Decode(&resp)
if resp["error"] != "Board or user not found" {
t.Errorf(`expected "Board or user not found", got %q`, resp["error"])
}
}
