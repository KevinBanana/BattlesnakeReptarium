package model

type MoveRequestBody struct {
	Game      Game  `json:"game"`
	Turn      int   `json:"turn,omitempty"`
	Board     Board `json:"board"`
	SelfSnake Snake `json:"you"`
}
