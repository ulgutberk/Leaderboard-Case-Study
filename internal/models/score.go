package models

// ScoreEntry is the base score representation used in list and surrounding responses.
type ScoreEntry struct {
	UserID string  `json:"userId"`
	Score  float64 `json:"score"`
}

// SetScoreRequest is the request body for POST /boards/{boardId}/scores.
type SetScoreRequest struct {
	UserID string  `json:"userId"`
	Score  float64 `json:"score"`
}

// PopulateMockScoresRequest is the request body for POST /boards/{boardId}/mock-scores.
type PopulateMockScoresRequest struct {
	Count int `json:"count"`
}

// SetScoreResponse is returned by POST /boards/{boardId}/scores.
type SetScoreResponse struct {
	BoardID string  `json:"boardId"`
	UserID  string  `json:"userId"`
	Score   float64 `json:"score"`
}

// PopulateMockScoresResponse describes the generated mock scores for a board.
type PopulateMockScoresResponse struct {
	BoardID string       `json:"boardId"`
	Count   int          `json:"count"`
	Scores  []ScoreEntry `json:"scores"`
}

// SurroundingsResponse is returned by GET /boards/{boardId}/scores/{userId}/surroundings.
type SurroundingsResponse struct {
	User  ScoreEntry   `json:"user"`
	Above []ScoreEntry `json:"above"`
	Below []ScoreEntry `json:"below"`
}
