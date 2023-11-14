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
