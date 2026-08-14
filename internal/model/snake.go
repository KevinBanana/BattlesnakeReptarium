package model

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

// TravelDirection returns the direction a snake last moved in, determined by
// where the head sits relative to the first body segment.
//
// ok is false when the snake has no travel direction yet
func (s Snake) TravelDirection() (direction Direction, ok bool) {
	if len(s.Body) <= 1 { // Snake is only a head
		return "", false
	}
	body := s.Body[1]
	if s.Head == *body.GetCoordInDirection(UP) {
		return UP, true
	}
	if s.Head == *body.GetCoordInDirection(DOWN) {
		return DOWN, true
	}
	if s.Head == *body.GetCoordInDirection(LEFT) {
		return LEFT, true
	}
	if s.Head == *body.GetCoordInDirection(RIGHT) {
		return RIGHT, true
	}
	return "", false
}
