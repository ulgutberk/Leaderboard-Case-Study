package repositories_test

import (
"context"
"fmt"
"testing"
"time"

"leaderboard-case-study/internal/models"
"leaderboard-case-study/internal/repositories"
)

func uniqueName(prefix string) string {
return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func TestBoardRepository_CreateAndGet(t *testing.T) {
db := newTestDB(t)
repo := repositories.NewBoardRepository(db)
ctx := context.Background()

interval := 3600
board := &models.Board{
Name:        uniqueName("create-get"),
Description: "integration test board",
Schedule:    &models.Schedule{Type: "interval", IntervalSeconds: &interval},
}

if err := repo.CreateBoard(ctx, board); err != nil {
t.Fatalf("CreateBoard: %v", err)
}
t.Cleanup(func() { db.Exec(ctx, "DELETE FROM boards WHERE id = $1", board.DbID) })

if board.DbID == 0 {
t.Error("DbID must be populated after CreateBoard")
}
if board.BoardID == "" {
t.Error("BoardID must be set after CreateBoard")
}

got, err := repo.GetBoard(ctx, board.DbID)
if err != nil {
t.Fatalf("GetBoard: %v", err)
}
if got.Name != board.Name {
t.Errorf("name: want %q, got %q", board.Name, got.Name)
}
if got.Schedule == nil || got.Schedule.Type != "interval" {
t.Errorf("schedule type: want 'interval', got %v", got.Schedule)
}
if got.Schedule.IntervalSeconds == nil || *got.Schedule.IntervalSeconds != 3600 {
t.Errorf("schedule interval: want 3600, got %v", got.Schedule.IntervalSeconds)
}
}

func TestBoardRepository_GetBoardByName(t *testing.T) {
db := newTestDB(t)
repo := repositories.NewBoardRepository(db)
ctx := context.Background()

name := uniqueName("by-name")
board := &models.Board{Name: name, Description: "test"}
if err := repo.CreateBoard(ctx, board); err != nil {
t.Fatalf("CreateBoard: %v", err)
}
t.Cleanup(func() { db.Exec(ctx, "DELETE FROM boards WHERE id = $1", board.DbID) })

got, err := repo.GetBoardByName(ctx, name)
if err != nil {
t.Fatalf("GetBoardByName: %v", err)
}
if got.BoardID != board.BoardID {
t.Errorf("boardID: want %q, got %q", board.BoardID, got.BoardID)
}
}

func TestBoardRepository_GetBoard_NotFound(t *testing.T) {
db := newTestDB(t)
repo := repositories.NewBoardRepository(db)
ctx := context.Background()

_, err := repo.GetBoard(ctx, -9999)
if err == nil {
t.Error("expected error for non-existent board ID, got nil")
}
}

func TestBoardRepository_ListBoards(t *testing.T) {
db := newTestDB(t)
repo := repositories.NewBoardRepository(db)
ctx := context.Background()

name := uniqueName("list")
board := &models.Board{Name: name, Description: "list test"}
if err := repo.CreateBoard(ctx, board); err != nil {
t.Fatalf("CreateBoard: %v", err)
}
t.Cleanup(func() { db.Exec(ctx, "DELETE FROM boards WHERE id = $1", board.DbID) })

summaries, err := repo.ListBoards(ctx)
if err != nil {
t.Fatalf("ListBoards: %v", err)
}
found := false
for _, s := range summaries {
if s.Name == name {
found = true
break
}
}
if !found {
t.Errorf("created board %q not found in ListBoards result", name)
}
}

func TestBoardRepository_ListScheduledBoards(t *testing.T) {
db := newTestDB(t)
repo := repositories.NewBoardRepository(db)
ctx := context.Background()

interval := 7200
name := uniqueName("scheduled")
board := &models.Board{
Name:     name,
Schedule: &models.Schedule{Type: "interval", IntervalSeconds: &interval},
}
if err := repo.CreateBoard(ctx, board); err != nil {
t.Fatalf("CreateBoard: %v", err)
}
t.Cleanup(func() { db.Exec(ctx, "DELETE FROM boards WHERE id = $1", board.DbID) })

boards, err := repo.ListScheduledBoards(ctx)
if err != nil {
t.Fatalf("ListScheduledBoards: %v", err)
}
found := false
for _, b := range boards {
if b.Name == name {
found = true
if b.Schedule == nil || b.Schedule.Type != "interval" {
t.Errorf("expected schedule on board %q, got %v", name, b.Schedule)
}
break
}
}
if !found {
t.Errorf("scheduled board %q not found in ListScheduledBoards result", name)
}
}

func TestBoardRepository_UpdateLastResetAt(t *testing.T) {
db := newTestDB(t)
repo := repositories.NewBoardRepository(db)
ctx := context.Background()

interval := 3600
board := &models.Board{
Name:     uniqueName("reset-at"),
Schedule: &models.Schedule{Type: "interval", IntervalSeconds: &interval},
}
if err := repo.CreateBoard(ctx, board); err != nil {
t.Fatalf("CreateBoard: %v", err)
}
t.Cleanup(func() { db.Exec(ctx, "DELETE FROM boards WHERE id = $1", board.DbID) })

resetTime := time.Now().UTC().Truncate(time.Second)
if err := repo.UpdateLastResetAt(ctx, board.DbID, resetTime); err != nil {
t.Fatalf("UpdateLastResetAt: %v", err)
}

got, err := repo.GetBoard(ctx, board.DbID)
if err != nil {
t.Fatalf("GetBoard after UpdateLastResetAt: %v", err)
}
if !got.LastResetAt.UTC().Truncate(time.Second).Equal(resetTime) {
t.Errorf("LastResetAt: want %v, got %v", resetTime, got.LastResetAt.UTC())
}
}
