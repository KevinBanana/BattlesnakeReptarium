package model

type Direction string

const (
	UP    Direction = "up"
	DOWN  Direction = "down"
	LEFT  Direction = "left"
	RIGHT Direction = "right"
)

type Snake struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Health         int                 `json:"health"`
	Body           []Coord             `json:"body"`
	Head           Coord               `json:"head"`
	Length         int                 `json:"length"`
	Latency        string              `json:"latency"`
	Shout          string              `json:"shout"`
	Squad          string              `json:"squad"`
	Customizations SnakeCustomizations `json:"customizations"`
}

type SnakeCustomizations struct {
	Head  string `json:"head"`
	Tail  string `json:"tail"`
	Color string `json:"color"`
}

type SnakeAction struct {
	Move  Direction `json:"move"`
	Shout string    `json:"shout"`
}

// FindSnakeTravelDirection returns the direction a snake last moved in
// It does this by determining where the head of the snake is in relation to the first body segment
func (s Snake) FindSnakeTravelDirection() Direction {
	// TODO handle case where snake is only one segment long, so it only has a head
	// TODO the first element in body array is head, look at index [1] instead
	// Find which direction the head is in from the first body segment
	body := s.Body[0]
	if s.Head == *body.GetSquareInDirection(UP) {
		return UP
	}
	if s.Head == *body.GetSquareInDirection(DOWN) {
		return DOWN
	}
	if s.Head == *body.GetSquareInDirection(LEFT) {
		return LEFT
	}
	if s.Head == *body.GetSquareInDirection(RIGHT) {
		return RIGHT
	}
	return ""
}
