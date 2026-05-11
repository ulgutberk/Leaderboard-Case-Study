package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"leaderboard-case-study/internal/models"
	"leaderboard-case-study/internal/services"
)

// BoardHandler handles HTTP requests for board and score operations.
// It delegates business logic to BoardService.
type BoardHandler struct {
	service services.BoardService
}

// NewBoardHandler creates a new BoardHandler.
func NewBoardHandler(service services.BoardService) *BoardHandler {
	return &BoardHandler{service: service}
}

// RegisterRoutes registers all board-related routes on the given router.
func (h *BoardHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/boards", h.CreateBoard).Methods(http.MethodPost)
	router.HandleFunc("/boards/{id}", h.GetBoard).Methods(http.MethodGet)
	router.HandleFunc("/boards/{id}/scores", h.SetScore).Methods(http.MethodPost)
	router.HandleFunc("/boards/{id}/scores", h.GetTopScores).Methods(http.MethodGet)
	router.HandleFunc("/boards/{id}/reset", h.ResetScores).Methods(http.MethodPost)
}

// CreateBoard handles POST /boards
func (h *BoardHandler) CreateBoard(w http.ResponseWriter, r *http.Request) {
	var board models.Board
	if err := json.NewDecoder(r.Body).Decode(&board); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.service.CreateBoard(r.Context(), &board); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(board)
}

// GetBoard handles GET /boards/{id}
func (h *BoardHandler) GetBoard(w http.ResponseWriter, r *http.Request) {
	id, err := parseBoardID(r)
	if err != nil {
		http.Error(w, "invalid board id", http.StatusBadRequest)
		return
	}
	board, err := h.service.GetBoard(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(board)
}

// SetScore handles POST /boards/{id}/scores
func (h *BoardHandler) SetScore(w http.ResponseWriter, r *http.Request) {
	id, err := parseBoardID(r)
	if err != nil {
		http.Error(w, "invalid board id", http.StatusBadRequest)
		return
	}
	var payload models.Score
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.service.SetScore(r.Context(), id, payload.UserID, payload.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetTopScores handles GET /boards/{id}/scores?limit=10
func (h *BoardHandler) GetTopScores(w http.ResponseWriter, r *http.Request) {
	id, err := parseBoardID(r)
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
	scores, err := h.service.GetTopScores(r.Context(), id, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}

// ResetScores handles POST /boards/{id}/reset
func (h *BoardHandler) ResetScores(w http.ResponseWriter, r *http.Request) {
	id, err := parseBoardID(r)
	if err != nil {
		http.Error(w, "invalid board id", http.StatusBadRequest)
		return
	}
	if err := h.service.ResetScores(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseBoardID(r *http.Request) (int, error) {
	return strconv.Atoi(mux.Vars(r)["id"])
}
