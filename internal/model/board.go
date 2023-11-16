package model

type Board struct {
	Height  int     `json:"height"`
	Width   int     `json:"width"`
	Food    []Coord `json:"food"`
	Hazards []Coord `json:"hazards"`
	Snakes  []Snake `json:"snakes"`
}

func (b Board) IsSnakeOnBoard(snake Snake) bool {
	for _, boardSnake := range b.Snakes {
		if boardSnake.ID == snake.ID {
			return true
		}
	}
	return false
}

func (b Board) IsCoordClear(coord Coord) bool {
	if !b.isCoordOnBoard(coord) {
		return false
	}

	// Check if a snake is at the coord
	for _, snake := range b.Snakes {
		for _, snakeBodyCoord := range snake.Body {
			if coord == snakeBodyCoord {
				return false
			}
		}
		if coord == snake.Head {
			return false
		}
	}

	// Check if a hazard is at the coord
	for _, hazard := range b.Hazards {
		if coord == hazard {
			return false
		}
	}

	// The coord is on the board and clear of snakes and hazards
	return true
}

// WillCoordBeClear returns true if the coord will be clear in the given number of turns
// This is useful for checking if a snake currently covering a coord will have moved on by the time we get there
func (b Board) WillCoordBeClear(coord Coord, turns int) bool {
	// TODO implement
	return false
}

func (b Board) isCoordOnBoard(coord Coord) bool {
	if coord.X < 0 || coord.X >= b.Width {
		return false
	}
	if coord.Y < 0 || coord.Y >= b.Height {
		return false
	}
	return true
}

// CalculateCoordsInChamber uses flood fill to count the number of cells connected to a given coord
func (b Board) CalculateCoordsInChamber(coord Coord) int {
	// TODO implement flood fill
	return 0
}
