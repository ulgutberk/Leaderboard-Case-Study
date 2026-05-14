package models

import "time"

type Schedule struct {
	Type            string `json:"type"`
	IntervalSeconds *int   `json:"intervalSeconds,omitempty"`
}

type BoardSummary struct {
	BoardID string `json:"boardId"`
	Name    string `json:"name"`
}

type Board struct {
	BoardID     string     `json:"boardId"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Schedule    *Schedule  `json:"schedule,omitempty"`
	NextResetAt *time.Time `json:"nextResetAt,omitempty"`

	CreatedAt time.Time `json:"-"`
	LastResetAt time.Time `json:"-"`
	DbID int `json:"-"`
}

type CreateBoardRequest struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Schedule    *Schedule `json:"schedule"`
}
