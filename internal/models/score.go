package models

type ScoreEntry struct {
	UserID string  `json:"userId"`
	Score  float64 `json:"score"`
}

type SetScoreRequest struct {
	UserID string  `json:"userId"`
	Score  float64 `json:"score"`
}

type PopulateMockScoresRequest struct {
	Count int `json:"count"`
}

type SetScoreResponse struct {
	BoardID string  `json:"boardId"`
	UserID  string  `json:"userId"`
	Score   float64 `json:"score"`
}

type PopulateMockScoresResponse struct {
	BoardID string       `json:"boardId"`
	Count   int          `json:"count"`
	Scores  []ScoreEntry `json:"scores"`
}

type SurroundingsResponse struct {
	User  ScoreEntry   `json:"user"`
	Above []ScoreEntry `json:"above"`
	Below []ScoreEntry `json:"below"`
}
