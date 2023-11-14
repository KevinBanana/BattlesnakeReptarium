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

func calculateMovesDistance(a, b model.Coord) int {
	xMoves := 0
	yMoves := 0

	if a.X > b.X {
		xMoves = a.X - b.X
	} else {
		xMoves = b.X - a.X
	}

	if a.Y > b.Y {
		yMoves = a.Y - b.Y
	} else {
		yMoves = b.Y - a.Y
	}

	return xMoves + yMoves
}
