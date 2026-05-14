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


func newUserHandler(svc *mockUserService) *handlers.UserHandler {
return handlers.NewUserHandler(svc)
}

func executeUserRequest(t *testing.T, h *handlers.UserHandler, req *http.Request) *httptest.ResponseRecorder {
t.Helper()
router := mux.NewRouter()
h.RegisterRoutes(router)
rr := httptest.NewRecorder()
router.ServeHTTP(rr, req)
return rr
}


func TestCreateUser_Success(t *testing.T) {
svc := &mockUserService{
createFn: func(_ context.Context, user *models.User) error {
if user.ID != "user_123" {
t.Errorf("expected id user_123, got %s", user.ID)
}
if user.Username != "alice" {
t.Errorf("expected username alice, got %s", user.Username)
}
return nil
},
getFn: func(_ context.Context, _ string) (*models.User, error) { return nil, nil },
}
body, _ := json.Marshal(map[string]any{"id": "user_123", "username": "alice"})
req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")

rr := executeUserRequest(t, newUserHandler(svc), req)

if rr.Code != http.StatusCreated {
t.Fatalf("expected 201, got %d — body: %s", rr.Code, rr.Body.String())
}
var resp models.User
if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
t.Fatalf("decode response: %v", err)
}
if resp.ID != "user_123" {
t.Errorf("expected id user_123, got %s", resp.ID)
}
if resp.Username != "alice" {
t.Errorf("expected username alice, got %s", resp.Username)
}
}

func TestCreateUser_InvalidJSON(t *testing.T) {
svc := &mockUserService{
createFn: func(_ context.Context, _ *models.User) error {
t.Error("CreateUser should not be called on invalid JSON")
return nil
},
getFn: func(_ context.Context, _ string) (*models.User, error) { return nil, nil },
}
req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{invalid`))
req.Header.Set("Content-Type", "application/json")

rr := executeUserRequest(t, newUserHandler(svc), req)

if rr.Code != http.StatusBadRequest {
t.Fatalf("expected 400, got %d", rr.Code)
}
}

func TestCreateUser_ServiceError(t *testing.T) {
svc := &mockUserService{
createFn: func(_ context.Context, _ *models.User) error {
return errors.New("db unavailable")
},
getFn: func(_ context.Context, _ string) (*models.User, error) { return nil, nil },
}
body, _ := json.Marshal(map[string]any{"id": "user_123", "username": "alice"})
req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
req.Header.Set("Content-Type", "application/json")

rr := executeUserRequest(t, newUserHandler(svc), req)

if rr.Code != http.StatusInternalServerError {
t.Fatalf("expected 500, got %d — body: %s", rr.Code, rr.Body.String())
}
}


func TestGetUser_Success(t *testing.T) {
svc := &mockUserService{
createFn: func(_ context.Context, _ *models.User) error { return nil },
getFn: func(_ context.Context, id string) (*models.User, error) {
if id != "user_123" {
t.Errorf("expected id user_123, got %s", id)
}
return &models.User{ID: "user_123", Username: "alice"}, nil
},
}
req, _ := http.NewRequest(http.MethodGet, "/users/user_123", nil)

rr := executeUserRequest(t, newUserHandler(svc), req)

if rr.Code != http.StatusOK {
t.Fatalf("expected 200, got %d — body: %s", rr.Code, rr.Body.String())
}
var resp models.User
if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
t.Fatalf("decode response: %v", err)
}
if resp.ID != "user_123" {
t.Errorf("expected id user_123, got %s", resp.ID)
}
if resp.Username != "alice" {
t.Errorf("expected username alice, got %s", resp.Username)
}
}

func TestGetUser_NotFound(t *testing.T) {
svc := &mockUserService{
createFn: func(_ context.Context, _ *models.User) error { return nil },
getFn: func(_ context.Context, _ string) (*models.User, error) {
return nil, errors.New("user not found")
},
}
req, _ := http.NewRequest(http.MethodGet, "/users/unknown_user", nil)

rr := executeUserRequest(t, newUserHandler(svc), req)

if rr.Code != http.StatusNotFound {
t.Fatalf("expected 404, got %d — body: %s", rr.Code, rr.Body.String())
}
}
