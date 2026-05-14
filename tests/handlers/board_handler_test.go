package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"leaderboard-case-study/internal/handlers"
	"leaderboard-case-study/internal/models"
)

// --- Mock ---

// mockBoardService is a test double for services.BoardService.
type mockBoardService struct {
	createFn    func(ctx context.Context, board *models.Board) error
	getFn       func(ctx context.Context, id int) (*models.Board, error)
	getByNameFn func(ctx context.Context, name string) (*models.Board, error)
	listFn      func(ctx context.Context) ([]models.BoardSummary, error)
}

func (m *mockBoardService) CreateBoard(ctx context.Context, board *models.Board) error {
	return m.createFn(ctx, board)
}

func (m *mockBoardService) GetBoard(ctx context.Context, id int) (*models.Board, error) {
	return m.getFn(ctx, id)
}

func (m *mockBoardService) GetBoardByName(ctx context.Context, name string) (*models.Board, error) {
	return m.getByNameFn(ctx, name)
}

func (m *mockBoardService) ListBoards(ctx context.Context) ([]models.BoardSummary, error) {
	return m.listFn(ctx)
}

// --- Helpers ---

// newCreateBoardRequest builds a POST /boards HTTP request from the given body.
func newCreateBoardRequest(t *testing.T, body any) *http.Request {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, "/boards", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

// executeRequest fires the handler and returns the recorded response.
func executeRequest(t *testing.T, h *handlers.BoardHandler, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	router := mux.NewRouter()
	h.RegisterRoutes(router)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// nopSvc returns a mockBoardService with no-op stubs for all methods
func nopSvc() *mockBoardService {
	return &mockBoardService{
		createFn:    func(_ context.Context, _ *models.Board) error { return nil },
		getFn:       func(_ context.Context, _ int) (*models.Board, error) { return nil, nil },
		getByNameFn: func(_ context.Context, _ string) (*models.Board, error) { return nil, nil },
		listFn:      func(_ context.Context) ([]models.BoardSummary, error) { return nil, nil },
	}
}

// --- CreateBoard Tests ---

func TestCreateBoard_Success(t *testing.T) {
	interval := 604800
	svc := nopSvc()
	svc.createFn = func(ctx context.Context, board *models.Board) error {
		board.DbID = 123
		board.BoardID = "board_123"
		return nil
	}
	h := handlers.NewBoardHandler(svc)

	body := map[string]any{
		"name":        "Weekly Tournament",
		"description": "Global leaderboard for weekly tournament",
		"schedule": map[string]any{
			"type":            "interval",
			"intervalSeconds": interval,
		},
	}

	rr := executeRequest(t, h, newCreateBoardRequest(t, body))

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d — body: %s", rr.Code, rr.Body.String())
	}

	var resp models.Board
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.BoardID != "board_123" {
		t.Errorf("expected boardId %q, got %q", "board_123", resp.BoardID)
	}
	if resp.Name != "Weekly Tournament" {
		t.Errorf("expected name %q, got %q", "Weekly Tournament", resp.Name)
	}
	if resp.Schedule == nil {
		t.Fatal("expected schedule to be present")
	}
	if resp.Schedule.Type != "interval" {
		t.Errorf("expected schedule.type %q, got %q", "interval", resp.Schedule.Type)
	}
	if resp.Schedule.IntervalSeconds == nil || *resp.Schedule.IntervalSeconds != interval {
		t.Errorf("expected schedule.intervalSeconds %d", interval)
	}
}

func TestCreateBoard_ValidationErrors(t *testing.T) {
	svc := nopSvc()
	svc.createFn = func(ctx context.Context, board *models.Board) error {
		t.Error("service.CreateBoard should not be called on invalid input")
		return nil
	}
	h := handlers.NewBoardHandler(svc)

	validInterval := 604800

	tests := []struct {
		name string
		body map[string]any
	}{
		{
			name: "missing name",
			body: map[string]any{
				"description": "desc",
				"schedule":    map[string]any{"type": "interval", "intervalSeconds": validInterval},
			},
		},
		{
			name: "missing description",
			body: map[string]any{
				"name":     "Weekly Tournament",
				"schedule": map[string]any{"type": "interval", "intervalSeconds": validInterval},
			},
		},
		{
			name: "missing schedule",
			body: map[string]any{
				"name":        "Weekly Tournament",
				"description": "desc",
			},
		},
		{
			name: "missing schedule.type",
			body: map[string]any{
				"name":        "Weekly Tournament",
				"description": "desc",
				"schedule":    map[string]any{"intervalSeconds": validInterval},
			},
		},
		{
			name: "missing schedule.intervalSeconds",
			body: map[string]any{
				"name":        "Weekly Tournament",
				"description": "desc",
				"schedule":    map[string]any{"type": "interval"},
			},
		},
		{
			name: "schedule.intervalSeconds is zero",
			body: map[string]any{
				"name":        "Weekly Tournament",
				"description": "desc",
				"schedule":    map[string]any{"type": "interval", "intervalSeconds": 0},
			},
		},
		{
			name: "schedule.intervalSeconds is negative",
			body: map[string]any{
				"name":        "Weekly Tournament",
				"description": "desc",
				"schedule":    map[string]any{"type": "interval", "intervalSeconds": -1},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rr := executeRequest(t, h, newCreateBoardRequest(t, tc.body))
			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d — body: %s", rr.Code, rr.Body.String())
			}
			var resp map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Errorf("expected JSON error body, got: %s", rr.Body.String())
			}
			if resp["error"] == "" {
				t.Errorf("expected non-empty error field in response")
			}
		})
	}
}

func TestCreateBoard_InvalidJSON(t *testing.T) {
	svc := nopSvc()
	svc.createFn = func(ctx context.Context, board *models.Board) error {
		t.Error("service.CreateBoard should not be called on invalid JSON")
		return nil
	}
	h := handlers.NewBoardHandler(svc)

	req, _ := http.NewRequest(http.MethodPost, "/boards", bytes.NewBufferString(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	rr := executeRequest(t, h, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestCreateBoard_DuplicateName(t *testing.T) {
	svc := nopSvc()
	svc.createFn = func(ctx context.Context, board *models.Board) error {
		return &duplicateNameError{}
	}
	h := handlers.NewBoardHandler(svc)

	body := map[string]any{
		"name":        "Weekly Tournament",
		"description": "desc",
		"schedule":    map[string]any{"type": "interval", "intervalSeconds": 604800},
	}
	rr := executeRequest(t, h, newCreateBoardRequest(t, body))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for unrecognized service error, got %d — body: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("expected JSON error body: %v", err)
	}
	if resp["error"] == "" {
		t.Error("expected non-empty error field")
	}
}

// --- ListBoards Tests ---

func TestListBoards_Success(t *testing.T) {
	svc := nopSvc()
	svc.listFn = func(_ context.Context) ([]models.BoardSummary, error) {
		return []models.BoardSummary{
			{BoardID: "board_123", Name: "Weekly Tournament"},
			{BoardID: "board_456", Name: "All-time Top Scores"},
		}, nil
	}
	h := handlers.NewBoardHandler(svc)

	req, _ := http.NewRequest(http.MethodGet, "/boards", nil)
	rr := executeRequest(t, h, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", rr.Code, rr.Body.String())
	}

	var resp []models.BoardSummary
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("expected 2 boards, got %d", len(resp))
	}
	if resp[0].BoardID != "board_123" {
		t.Errorf("expected boardId %q, got %q", "board_123", resp[0].BoardID)
	}
	if resp[1].Name != "All-time Top Scores" {
		t.Errorf("expected name %q, got %q", "All-time Top Scores", resp[1].Name)
	}
}

// --- GetBoard Tests ---

func TestGetBoard_Success(t *testing.T) {
	interval := 604800
	createdAt := time.Now().Add(-8 * 24 * time.Hour)
	svc := nopSvc()
	svc.getFn = func(_ context.Context, id int) (*models.Board, error) {
		if id != 123 {
			t.Errorf("expected id 123, got %d", id)
		}
		return &models.Board{
			BoardID:     "board_123",
			Name:        "Weekly Tournament",
			Description: "Global leaderboard for weekly tournament",
			CreatedAt:   createdAt,
			Schedule:    &models.Schedule{Type: "interval", IntervalSeconds: &interval},
		}, nil
	}
	h := handlers.NewBoardHandler(svc)

	req, _ := http.NewRequest(http.MethodGet, "/boards/board_123", nil)
	rr := executeRequest(t, h, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", rr.Code, rr.Body.String())
	}

	var resp models.Board
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.BoardID != "board_123" {
		t.Errorf("expected boardId %q, got %q", "board_123", resp.BoardID)
	}
	if resp.NextResetAt == nil {
		t.Error("expected nextResetAt to be set for a board with an interval schedule")
	}
	if resp.NextResetAt != nil && resp.NextResetAt.Before(time.Now()) {
		t.Error("expected nextResetAt to be in the future")
	}
}

func TestGetBoard_NotFound(t *testing.T) {
	svc := nopSvc()
	svc.getFn = func(_ context.Context, _ int) (*models.Board, error) {
		return nil, &boardNotFoundError{}
	}
	h := handlers.NewBoardHandler(svc)

	req, _ := http.NewRequest(http.MethodGet, "/boards/board_999", nil)
	rr := executeRequest(t, h, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d — body: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if resp["error"] != "Board not found" {
		t.Errorf("expected error %q, got %q", "Board not found", resp["error"])
	}
}

func TestGetBoard_InvalidID(t *testing.T) {

	svc := nopSvc()
	h := handlers.NewBoardHandler(svc)

	req, _ := http.NewRequest(http.MethodGet, "/boards/not-a-number", nil)
	rr := executeRequest(t, h, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

type boardNotFoundError struct{}

func (e *boardNotFoundError) Error() string { return "board not found" }

type duplicateNameError struct{}

func (e *duplicateNameError) Error() string { return "duplicate board name" }
func (e *duplicateNameError) Is(target error) bool {
	return errors.Is(target, e)
}
