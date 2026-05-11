package models

import "time"

// Schedule defines when a board's scores should be automatically reset.
type Schedule struct {
	Type            string `json:"type"`                      // e.g. "interval"
	IntervalSeconds *int   `json:"intervalSeconds,omitempty"` // used when type is "interval"
}

// BoardSummary is the lightweight projection returned by GET /boards.
type BoardSummary struct {
	BoardID string `json:"boardId"`
	Name    string `json:"name"`
}

// Board represents a leaderboard.
// Metadata is stored persistently in Postgres.
// BoardID is exposed as "board_{id}" (e.g. "board_123").
type Board struct {
	BoardID     string     `json:"boardId"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Schedule    *Schedule  `json:"schedule,omitempty"`
	NextResetAt *time.Time `json:"nextResetAt,omitempty"`

	// CreatedAt is managed by Postgres (DEFAULT NOW()); hidden from JSON.
	CreatedAt time.Time `json:"-"`
	// DbID is the internal integer primary key used only within the repository layer.
	DbID int `json:"-"`
}
