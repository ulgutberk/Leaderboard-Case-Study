package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgconn"

	"leaderboard-case-study/internal/models"
	"leaderboard-case-study/internal/services"
)

type ScoreHandler struct {
	service      services.ScoreService
	boardService services.BoardService
	userService  services.UserService
	randSource   *rand.Rand
}

func NewScoreHandler(service services.ScoreService, boardService services.BoardService, userService services.UserService) *ScoreHandler {
	return &ScoreHandler{
		service:      service,
		boardService: boardService,
		userService:  userService,
		randSource:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (h *ScoreHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/boards/{boardId}/scores", h.SetScore).Methods(http.MethodPost)
	router.HandleFunc("/boards/{boardId}/scores", h.GetTopScores).Methods(http.MethodGet)
	router.HandleFunc("/boards/{boardId}/mock-scores", h.PopulateMockScores).Methods(http.MethodPost)
	router.HandleFunc("/boards/{boardId}/scores/{userId}/surroundings", h.GetSurroundings).Methods(http.MethodGet)
	router.HandleFunc("/boards/{boardId}/reset", h.ResetScores).Methods(http.MethodPost)
}

func (h *ScoreHandler) parseBoardID(r *http.Request) (id int, boardIDStr string, err error) {
	raw := mux.Vars(r)["boardId"]
	trimmed := strings.TrimPrefix(raw, "board_")
	id, err = strconv.Atoi(trimmed)
	if err != nil {
		return 0, "", fmt.Errorf("invalid board id: %s", raw)
	}
	return id, fmt.Sprintf("board_%d", id), nil
}

func (h *ScoreHandler) requireBoard(w http.ResponseWriter, r *http.Request, boardID int) bool {
	if _, err := h.boardService.GetBoard(r.Context(), boardID); err != nil {
		jsonError(w, "Board not found", http.StatusNotFound)
		return false
	}
	return true
}

// SetScore godoc
// @Summary      Set a user's score
// @Description  Creates or overwrites a user's score on the board. Higher scores rank higher; ties broken by submission time (earlier = higher rank).
// @Tags         scores
// @Accept       json
// @Produce      json
// @Param        boardId  path  string  true  "Board ID (e.g. board_123)"
// @Param        body     body  models.SetScoreRequest  true  "userId and score"
// @Success      200  {object}  models.SetScoreResponse
// @Failure      400  {object}  map[string]string  "validation error"
// @Failure      404  {object}  map[string]string  "Board or user not found"
// @Router       /boards/{boardId}/scores [post]
func (h *ScoreHandler) SetScore(w http.ResponseWriter, r *http.Request) {
	boardID, boardIDStr, err := h.parseBoardID(r)
	if err != nil {
		jsonError(w, "Board not found", http.StatusNotFound)
		return
	}
	if !h.requireBoard(w, r, boardID) {
		return
	}

	var payload models.SetScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if payload.UserID == "" {
		jsonError(w, `"userId" is required`, http.StatusBadRequest)
		return
	}

	if err := h.service.SetScore(r.Context(), boardID, payload.UserID, payload.Score); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			jsonError(w, "User not found", http.StatusNotFound)
			return
		}
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SetScoreResponse{
		BoardID: boardIDStr,
		UserID:  payload.UserID,
		Score:   payload.Score,
	})
}

// GetTopScores godoc
// @Summary      Get top scores
// @Description  Returns top n users ranked by score (descending). Ties broken by submission time.
// @Tags         scores
// @Produce      json
// @Param        boardId  path   string  true   "Board ID (e.g. board_123)"
// @Param        n        query  int     false  "Number of scores to return (default 10)"
// @Success      200  {array}   models.ScoreEntry
// @Failure      400  {object}  map[string]string  "Invalid value for n"
// @Failure      404  {object}  map[string]string  "Board not found"
// @Router       /boards/{boardId}/scores [get]
func (h *ScoreHandler) GetTopScores(w http.ResponseWriter, r *http.Request) {
	boardID, _, err := h.parseBoardID(r)
	if err != nil {
		jsonError(w, "Board not found", http.StatusNotFound)
		return
	}
	if !h.requireBoard(w, r, boardID) {
		return
	}

	nStr := r.URL.Query().Get("n")
	if nStr == "" {
		nStr = "10"
	}
	n, err := strconv.ParseInt(nStr, 10, 64)
	if err != nil || n <= 0 {
		jsonError(w, "Invalid value for n", http.StatusBadRequest)
		return
	}

	scores, err := h.service.GetTopScores(r.Context(), boardID, n)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}

// PopulateMockScores godoc
// @Summary      Populate a board with mock scores
// @Description  Creates n mock users with random scores on the board to facilitate testing.
// @Tags         scores
// @Accept       json
// @Produce      json
// @Param        boardId  path  string                            true  "Board ID (e.g. board_123)"
// @Param        body     body  models.PopulateMockScoresRequest  true  "Number of mock users to create"
// @Success      201      {object}  models.PopulateMockScoresResponse
// @Failure      400      {object}  map[string]string  "validation error"
// @Failure      404      {object}  map[string]string  "Board not found"
// @Router       /boards/{boardId}/mock-scores [post]
func (h *ScoreHandler) PopulateMockScores(w http.ResponseWriter, r *http.Request) {
	boardID, boardIDStr, err := h.parseBoardID(r)
	if err != nil {
		jsonError(w, "Board not found", http.StatusNotFound)
		return
	}
	if !h.requireBoard(w, r, boardID) {
		return
	}

	var payload models.PopulateMockScoresRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if payload.Count <= 0 {
		jsonError(w, `"count" must be a positive integer`, http.StatusBadRequest)
		return
	}

	entries := make([]models.ScoreEntry, 0, payload.Count)
	seed := time.Now().UnixNano()
	for index := 0; index < payload.Count; index++ {
		userID := fmt.Sprintf("mock_user_%d_%d", seed, index+1)
		username := fmt.Sprintf("mock_player_%d_%d", seed, index+1)
		score := float64(h.randSource.Intn(9901) + 100)

		user := &models.User{ID: userID, Username: username}
		if err := h.userService.CreateUser(r.Context(), user); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := h.service.SetScore(r.Context(), boardID, userID, score); err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		entries = append(entries, models.ScoreEntry{UserID: userID, Score: score})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.PopulateMockScoresResponse{
		BoardID: boardIDStr,
		Count:   payload.Count,
		Scores:  entries,
	})
}

// GetSurroundings godoc
// @Summary      Get score surroundings
// @Description  Returns n users above and below the specified user in the rankings.
// @Tags         scores
// @Produce      json
// @Param        boardId  path   string  true   "Board ID (e.g. board_123)"
// @Param        userId   path   string  true   "User ID"
// @Param        n        query  int     false  "Number of neighbours above and below (default 5)"
// @Success      200  {object}  models.SurroundingsResponse
// @Failure      404  {object}  map[string]string  "Board or user not found"
// @Router       /boards/{boardId}/scores/{userId}/surroundings [get]
func (h *ScoreHandler) GetSurroundings(w http.ResponseWriter, r *http.Request) {
	boardID, _, err := h.parseBoardID(r)
	if err != nil {
		jsonError(w, "Board or user not found", http.StatusNotFound)
		return
	}
	if !h.requireBoard(w, r, boardID) {
		return
	}

	userID := mux.Vars(r)["userId"]

	nStr := r.URL.Query().Get("n")
	if nStr == "" {
		nStr = "5"
	}
	n, err := strconv.ParseInt(nStr, 10, 64)
	if err != nil || n <= 0 {
		jsonError(w, `"n" must be a positive integer`, http.StatusBadRequest)
		return
	}

	result, err := h.service.GetSurroundings(r.Context(), boardID, userID, n)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result == nil {
		jsonError(w, "Board or user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ResetScores godoc
// @Summary      Reset all scores on a board
// @Description  Deletes all scores for the board from Postgres and Redis.
// @Tags         scores
// @Param        boardId  path  string  true  "Board ID (e.g. board_123)"
// @Success      204
// @Router       /boards/{boardId}/reset [post]
func (h *ScoreHandler) ResetScores(w http.ResponseWriter, r *http.Request) {
	boardID, _, err := h.parseBoardID(r)
	if err != nil {
		jsonError(w, "Board not found", http.StatusNotFound)
		return
	}
	if err := h.service.ResetScores(r.Context(), boardID); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
