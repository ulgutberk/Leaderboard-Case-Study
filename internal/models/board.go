package models

import "time"

// Board represents a leaderboard.
// Metadata (name, schedule) is stored persistently in Postgres.
type Board struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	// ResetCron is an optional cron expression (e.g. "0 0 * * MON") that
	// triggers a score reset when the schedule fires.
	ResetCron *string   `json:"reset_cron,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
