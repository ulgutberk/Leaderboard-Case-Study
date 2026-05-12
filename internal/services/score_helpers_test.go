package services

import (
"testing"
"time"

"leaderboard-case-study/internal/models"
)

func intPtr(v int) *int { return &v }

func newScheduledBoard(intervalSecs int, createdAt time.Time, lastResetAt time.Time) *models.Board {
return &models.Board{
DbID:        1,
Schedule:    &models.Schedule{Type: "interval", IntervalSeconds: intPtr(intervalSecs)},
CreatedAt:   createdAt,
LastResetAt: lastResetAt,
}
}

func TestActivePeriodStart_NoSchedule(t *testing.T) {
board := &models.Board{}
_, ok := activePeriodStart(board, time.Now())
if ok {
t.Error("expected ok=false for nil schedule")
}
}

func TestActivePeriodStart_WrongType(t *testing.T) {
board := &models.Board{Schedule: &models.Schedule{Type: "daily"}}
_, ok := activePeriodStart(board, time.Now())
if ok {
t.Error("expected ok=false for non-interval schedule type")
}
}

func TestActivePeriodStart_NilInterval(t *testing.T) {
board := &models.Board{Schedule: &models.Schedule{Type: "interval", IntervalSeconds: nil}}
_, ok := activePeriodStart(board, time.Now())
if ok {
t.Error("expected ok=false for nil IntervalSeconds")
}
}

func TestActivePeriodStart_ZeroInterval(t *testing.T) {
board := &models.Board{Schedule: &models.Schedule{Type: "interval", IntervalSeconds: intPtr(0)}}
_, ok := activePeriodStart(board, time.Now())
if ok {
t.Error("expected ok=false for zero IntervalSeconds")
}
}

func TestActivePeriodStart_NowBeforeCreatedAt(t *testing.T) {
createdAt := time.Now().Add(time.Hour) // future
board := newScheduledBoard(3600, createdAt, createdAt)
result, ok := activePeriodStart(board, time.Now())
if !ok {
t.Fatal("expected ok=true")
}
if !result.Equal(createdAt) {
t.Errorf("expected period start = createdAt, got %v", result)
}
}

func TestActivePeriodStart_FirstPeriod(t *testing.T) {
createdAt := time.Now().Add(-30 * time.Minute)
board := newScheduledBoard(3600, createdAt, createdAt) // interval=1h, elapsed=30min → 0 periods
result, ok := activePeriodStart(board, time.Now())
if !ok {
t.Fatal("expected ok=true")
}
if !result.Equal(createdAt) {
t.Errorf("expected period start = createdAt (%v), got %v", createdAt, result)
}
}

func TestActivePeriodStart_SecondPeriod(t *testing.T) {
interval := time.Hour
createdAt := time.Now().Add(-90 * time.Minute) // 1.5 intervals elapsed → period 1 started
board := newScheduledBoard(3600, createdAt, createdAt)
result, ok := activePeriodStart(board, time.Now())
if !ok {
t.Fatal("expected ok=true")
}
want := createdAt.Add(1 * interval)
if !result.Equal(want) {
t.Errorf("expected %v, got %v", want, result)
}
}

func TestActivePeriodStart_ExactBoundary(t *testing.T) {
interval := time.Hour
createdAt := time.Now().Add(-2 * interval) // exactly 2 intervals elapsed
board := newScheduledBoard(3600, createdAt, createdAt)
result, ok := activePeriodStart(board, time.Now())
if !ok {
t.Fatal("expected ok=true")
}
want := createdAt.Add(2 * interval)
// Allow 1ms tolerance for elapsed time during test execution.
diff := result.Sub(want)
if diff < 0 {
diff = -diff
}
if diff > time.Millisecond {
t.Errorf("expected ~%v, got %v (diff=%v)", want, result, diff)
}
}
