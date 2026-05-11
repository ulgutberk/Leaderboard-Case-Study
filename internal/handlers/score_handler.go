package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"leaderboard-case-study/internal/models"
	"leaderboard-case-study/internal/services"
)

// ScoreHandler handles HTTP requests for score operations.
// Score routes are nested under /boards/{id} since scores belong to a board.
type ScoreHandler struct {
	service services.ScoreService
}

// NewScoreHandler creates a new ScoreHandler.
func NewScoreHandler(service services.ScoreService) *ScoreHandler {
	return &ScoreHandler{service: service}
}

// RegisterRoutes registers score-related routes on the given router.
func (h *ScoreHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/boards/{id}/scores", h.SetScore).Methods(http.MethodPost)
	router.HandleFunc("/boards/{id}/scores", h.GetTopScores).Methods(http.MethodGet)
	router.HandleFunc("/boards/{id}/reset", h.ResetScores).Methods(http.MethodPost)
}

// SetScore godoc
// @Summary      Set a user's score on a board
// @Description  Adds or updates a user's score on the given board (stored in Redis ZSET)
// @Tags         scores
// @Accept       json
// @Param        id     path  string       true  "Board ID"
// @Param        score  body  models.Score true  "Score payload"
// @Success      204
// @Failure      400  {string}  string  "invalid request"
// @Failure      500  {string}  string  "internal server error"
// @Router       /boards/{id}/scores [post]
func (h *ScoreHandler) SetScore(w http.ResponseWriter, r *http.Request) {
	boardID, err := parseBoardIDFromVars(r)
	if err != nil {
		http.Error(w, "invalid board id", http.StatusBadRequest)
		return
	}
	var payload models.Score
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.service.SetScore(r.Context(), boardID, payload.UserID, payload.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetTopScores godoc
// @Summary      Get top scores for a board
// @Description  Returns the top N scores from the board's Redis ZSET in descending order
// @Tags         scores
// @Produce      json
// @Param        id     path   string  true   "Board ID"
// @Param        limit  query  int     false  "Number of scores to return (default 10)"
// @Success      200    {array}   models.Score
// @Failure      400    {string}  string  "invalid board id"
// @Failure      500    {string}  string  "internal server error"
// @Router       /boards/{id}/scores [get]
func (h *ScoreHandler) GetTopScores(w http.ResponseWriter, r *http.Request) {
	boardID, err := parseBoardIDFromVars(r)
	if err != nil {
		http.Error(w, "invalid board id", http.StatusBadRequest)
		return
	}
	limit := int64(10)
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	scores, err := h.service.GetTopScores(r.Context(), boardID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}

// ResetScores godoc
// @Summary      Reset all scores on a board
// @Description  Deletes all score entries from the board's Redis ZSET
// @Tags         scores
// @Param        id  path  string  true  "Board ID"
// @Success      204
// @Failure      400  {string}  string  "invalid board id"
// @Failure      500  {string}  string  "internal server error"
// @Router       /boards/{id}/reset [post]
func (h *ScoreHandler) ResetScores(w http.ResponseWriter, r *http.Request) {
	boardID, err := parseBoardIDFromVars(r)
	if err != nil {
		http.Error(w, "invalid board id", http.StatusBadRequest)
		return
	}
	if err := h.service.ResetScores(r.Context(), boardID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// parseBoardIDFromVars parses both "board_123" and "123" formats from the URL path variable.
func parseBoardIDFromVars(r *http.Request) (int, error) {
	raw := mux.Vars(r)["id"]
	raw = strings.TrimPrefix(raw, "board_")
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid board id: %s", raw)
	}
	return id, nil
}
