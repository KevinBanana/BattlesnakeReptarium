package model

type Board struct {
	Height  int     `json:"height"`
	Width   int     `json:"width"`
	Food    []Coord `json:"food"`
	Hazards []Coord `json:"hazards"`
	Snakes  []Snake `json:"snakes"`
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

func (b Board) isCoordOnBoard(coord Coord) bool {
	if coord.X < 0 || coord.X >= b.Width {
		return false
	}

	if coord.Y < 0 || coord.Y >= b.Height {
		return false
	}

	return true
}
