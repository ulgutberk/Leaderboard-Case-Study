package models

// Score represents a user's score entry on a specific board.
// The live value is stored in Redis as a ZSET member for fast ranked access.
// Rank is populated on read (not stored).
type Score struct {
	UserID  string  `json:"user_id"`
	BoardID int     `json:"board_id"`
	Value   float64 `json:"score"`
	Rank    int64   `json:"rank,omitempty"` // 1-based rank, populated on read
}
