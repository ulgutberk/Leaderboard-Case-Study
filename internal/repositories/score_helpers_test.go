package repositories

import "testing"

func TestSortEntries_DescendingScore(t *testing.T) {
entries := []scoredEntry{
{UserID: "a", Score: 100, Ts: 1000},
{UserID: "b", Score: 300, Ts: 2000},
{UserID: "c", Score: 200, Ts: 3000},
}
sortEntries(entries)
want := []string{"b", "c", "a"}
for i, e := range entries {
if e.UserID != want[i] {
t.Errorf("position %d: expected %q, got %q", i, want[i], e.UserID)
}
}
}

func TestSortEntries_TieBreakByTimestamp(t *testing.T) {
entries := []scoredEntry{
{UserID: "late", Score: 500, Ts: 3000},
{UserID: "early", Score: 500, Ts: 1000},
{UserID: "mid", Score: 500, Ts: 2000},
}
sortEntries(entries)
// Earlier Ts = higher rank on equal scores.
want := []string{"early", "mid", "late"}
for i, e := range entries {
if e.UserID != want[i] {
t.Errorf("position %d: expected %q, got %q", i, want[i], e.UserID)
}
}
}

func TestSortEntries_MixedScoreAndTie(t *testing.T) {
entries := []scoredEntry{
{UserID: "x", Score: 200, Ts: 500},
{UserID: "y", Score: 300, Ts: 100},
{UserID: "z", Score: 300, Ts: 200},
}
sortEntries(entries)
// y and z both 300; y earlier (Ts=100) → rank 0; then z (Ts=200) → rank 1; then x (score 200) → rank 2.
want := []string{"y", "z", "x"}
for i, e := range entries {
if e.UserID != want[i] {
t.Errorf("position %d: expected %q, got %q", i, want[i], e.UserID)
}
}
}

func TestSortEntries_SingleEntry(t *testing.T) {
entries := []scoredEntry{{UserID: "only", Score: 42, Ts: 999}}
sortEntries(entries)
if len(entries) != 1 || entries[0].UserID != "only" {
t.Errorf("unexpected result: %+v", entries)
}
}

func TestSortEntries_Empty(t *testing.T) {
entries := []scoredEntry{}
sortEntries(entries) // must not panic
if len(entries) != 0 {
t.Errorf("expected empty slice, got %+v", entries)
}
}
