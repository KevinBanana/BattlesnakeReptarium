package model

// RequestBody contains information about the current state of the game.
// The same body is used for Start, Move, and End requests.
type RequestBody struct {
	Game      Game  `json:"game"`
	Turn      int   `json:"turn,omitempty"`
	Board     Board `json:"board"`
	SelfSnake Snake `json:"you"`
}
