package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"leaderboard-case-study/internal/models"
	"leaderboard-case-study/internal/services"
)

// UserHandler handles HTTP requests for user operations.
type UserHandler struct {
	service services.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(service services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// RegisterRoutes registers all user-related routes on the given router.
func (h *UserHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/users", h.CreateUser).Methods(http.MethodPost)
	router.HandleFunc("/users/{id}", h.GetUser).Methods(http.MethodGet)
}

// CreateUser godoc
// @Summary      Create a new user
// @Description  Registers a new user in the system (stored in Postgres)
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      models.User  true  "User payload"
// @Success      201   {object}  models.User
// @Failure      400   {string}  string  "invalid request"
// @Failure      500   {string}  string  "internal server error"
// @Router       /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.service.CreateUser(r.Context(), &user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// GetUser godoc
// @Summary      Get a user by ID
// @Description  Retrieves a user by their ID
// @Tags         users
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  models.User
// @Failure      404  {string}  string  "not found"
// @Router       /users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	user, err := h.service.GetUser(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
