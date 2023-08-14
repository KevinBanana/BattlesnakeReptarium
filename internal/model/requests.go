package model

type MoveRequestBody struct {
	Game      Game
	Turn      int
	Board     Board
	SelfSnake Snake
}
