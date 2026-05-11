package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgconn"

	"leaderboard-case-study/internal/models"
	"leaderboard-case-study/internal/services"
)

// jsonError writes a JSON {"error": msg} response with the given status code.
func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// BoardHandler handles HTTP requests for board metadata operations only.
type BoardHandler struct {
	service services.BoardService
}

// NewBoardHandler creates a new BoardHandler.
func NewBoardHandler(service services.BoardService) *BoardHandler {
	return &BoardHandler{service: service}
}

// RegisterRoutes registers board-related routes on the given router.
func (h *BoardHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/boards", h.ListBoards).Methods(http.MethodGet)
	router.HandleFunc("/boards", h.CreateBoard).Methods(http.MethodPost)
	router.HandleFunc("/boards/{boardId}", h.GetBoard).Methods(http.MethodGet)
}

// ListBoards godoc
// @Summary      List all boards
// @Description  Returns all leaderboards with boardId and name
// @Tags         boards
// @Produce      json
// @Success      200  {array}   models.Board
// @Failure      500  {string}  string  "internal server error"
// @Router       /boards [get]
func (h *BoardHandler) ListBoards(w http.ResponseWriter, r *http.Request) {
	boards, err := h.service.ListBoards(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if boards == nil {
		boards = []models.BoardSummary{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boards)
}

// CreateBoard godoc
// @Summary      Create a new board
// @Description  Creates a new leaderboard with a name, description and reset schedule
// @Tags         boards
// @Accept       json
// @Produce      json
// @Param        board  body      models.Board  true  "Board payload"
// @Success      201    {object}  models.Board
// @Failure      400    {string}  string        "validation error"
// @Failure      500    {string}  string        "internal server error"
// @Router       /boards [post]
func (h *BoardHandler) CreateBoard(w http.ResponseWriter, r *http.Request) {
	var board models.Board
	if err := json.NewDecoder(r.Body).Decode(&board); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if board.Name == "" {
		jsonError(w, `"name" is required`, http.StatusBadRequest)
		return
	}
	if board.Description == "" {
		jsonError(w, `"description" is required`, http.StatusBadRequest)
		return
	}
	if board.Schedule == nil {
		jsonError(w, `"schedule" is required`, http.StatusBadRequest)
		return
	}
	if board.Schedule.Type == "" {
		jsonError(w, `"schedule.type" is required`, http.StatusBadRequest)
		return
	}
	if board.Schedule.IntervalSeconds == nil || *board.Schedule.IntervalSeconds <= 0 {
		jsonError(w, `"schedule.intervalSeconds" is required and must be positive`, http.StatusBadRequest)
		return
	}

	if err := h.service.CreateBoard(r.Context(), &board); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			jsonError(w, "a board with this name already exists", http.StatusBadRequest)
			return
		}
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(board)
}

// GetBoard godoc
// @Summary      Get a board by ID
// @Description  Returns board details including the next scheduled reset time
// @Tags         boards
// @Produce      json
// @Param        boardId  path      string  true  "Board ID (e.g. board_123)"
// @Success      200      {object}  models.Board
// @Failure      404      {object}  map[string]string  "Board not found"
// @Router       /boards/{boardId} [get]
func (h *BoardHandler) GetBoard(w http.ResponseWriter, r *http.Request) {
	raw := mux.Vars(r)["boardId"]
	raw = strings.TrimPrefix(raw, "board_")
	id, err := strconv.Atoi(raw)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Board not found"})
		return
	}
	board, err := h.service.GetBoard(r.Context(), id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Board not found"})
		return
	}
	if board.Schedule != nil && board.Schedule.IntervalSeconds != nil && *board.Schedule.IntervalSeconds > 0 {
		interval := time.Duration(*board.Schedule.IntervalSeconds) * time.Second
		elapsed := time.Since(board.CreatedAt)
		intervals := int64(elapsed / interval)
		nextReset := board.CreatedAt.Add(time.Duration(intervals+1) * interval)
		board.NextResetAt = &nextReset
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(board)
}
