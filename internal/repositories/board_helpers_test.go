package repositories

import (
"testing"

"leaderboard-case-study/internal/models"
)

func TestBuildSchedule_WithTypeAndInterval(t *testing.T) {
sType := "interval"
interval := 3600
s := buildSchedule(&sType, &interval)
if s == nil {
t.Fatal("expected non-nil Schedule")
}
if s.Type != "interval" {
t.Errorf("expected type 'interval', got %q", s.Type)
}
if s.IntervalSeconds == nil || *s.IntervalSeconds != 3600 {
t.Errorf("expected interval 3600, got %v", s.IntervalSeconds)
}
}

func TestBuildSchedule_NilType(t *testing.T) {
s := buildSchedule(nil, nil)
if s != nil {
t.Errorf("expected nil Schedule, got %+v", s)
}
}

func TestBuildSchedule_TypeWithNilInterval(t *testing.T) {
sType := "interval"
s := buildSchedule(&sType, nil)
if s == nil {
t.Fatal("expected non-nil Schedule")
}
if s.IntervalSeconds != nil {
t.Errorf("expected nil IntervalSeconds, got %v", s.IntervalSeconds)
}
}

func TestScheduleType_NonNil(t *testing.T) {
interval := 60
sched := &models.Schedule{Type: "interval", IntervalSeconds: &interval}
result := scheduleType(sched)
if result == nil || *result != "interval" {
t.Errorf("expected 'interval', got %v", result)
}
}

func TestScheduleType_Nil(t *testing.T) {
result := scheduleType(nil)
if result != nil {
t.Errorf("expected nil, got %v", result)
}
}

func TestScheduleInterval_NonNil(t *testing.T) {
interval := 7200
sched := &models.Schedule{Type: "interval", IntervalSeconds: &interval}
result := scheduleInterval(sched)
if result == nil || *result != 7200 {
t.Errorf("expected 7200, got %v", result)
}
}

func TestScheduleInterval_Nil(t *testing.T) {
result := scheduleInterval(nil)
if result != nil {
t.Errorf("expected nil, got %v", result)
}
}
