package models

import "time"

// User represents a player tracked in the leaderboard system.
// Stored persistently in Postgres.
type User struct {
	ID        string    `json:"id"` // e.g. UUID or external user identifier
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}
