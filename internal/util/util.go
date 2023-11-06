package util

import "BattlesnakeReptarium/internal/model"

func IsSnakeOnBoard(snake model.Snake, board model.Board) bool {
	for _, boardSnake := range board.Snakes {
		if boardSnake.ID == snake.ID {
			return true
		}
	}
	return false
}
