package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"leaderboard-case-study/internal/handlers"
	"leaderboard-case-study/internal/models"
)

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
		getFn: func(_ context.Context, id string) (*models.User, error) { return nil, nil },
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
		getTopScoresFn: func(_ context.Context, _ int, _ int64) ([]models.ScoreEntry, error) { return nil, nil },
		getAroundFn: func(_ context.Context, _ int, _ string, _ int64) (*models.SurroundingsResponse, error) {
			return nil, nil
		},
		resetScoresFn:    func(_ context.Context, _ int) error { return nil },
		resetDueBoardsFn: func(_ context.Context) error { return nil },
	}

	handler := newScoreHandler(boardSvc, scoreSvc, userSvc)
	body := map[string]any{"count": 3}
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, "/boards/board_12/mock-scores", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
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
	scoreSvc := &mockScoreService{
		setScoreFn:     func(_ context.Context, _ int, _ string, _ float64) error { return nil },
		getTopScoresFn: func(_ context.Context, _ int, _ int64) ([]models.ScoreEntry, error) { return nil, nil },
		getAroundFn: func(_ context.Context, _ int, _ string, _ int64) (*models.SurroundingsResponse, error) {
			return nil, nil
		},
		resetScoresFn:    func(_ context.Context, _ int) error { return nil },
		resetDueBoardsFn: func(_ context.Context) error { return nil },
	}
	userSvc := &mockUserService{
		createFn: func(_ context.Context, _ *models.User) error {
			t.Fatal("CreateUser should not be called for invalid count")
			return nil
		},
		getFn: func(_ context.Context, _ string) (*models.User, error) { return nil, nil },
	}

	handler := newScoreHandler(boardSvc, scoreSvc, userSvc)
	req, err := http.NewRequest(http.MethodPost, "/boards/board_12/mock-scores", bytes.NewBufferString(`{"count":0}`))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := executeScoreRequest(t, handler, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
}
