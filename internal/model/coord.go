package model

import (
	"math"

	log "github.com/sirupsen/logrus"
)

type Direction string

const (
	UP    Direction = "up"
	DOWN  Direction = "down"
	LEFT  Direction = "left"
	RIGHT Direction = "right"
)

type Coord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// GetDistance returns the distance between two coordinates
// Calculated as sqrt((x2-x1)^2 + (y2-y1)^2)
func (c Coord) GetDistance(other Coord) float64 {
	xDist := math.Pow(float64(other.X-c.X), 2)
	yDist := math.Pow(float64(other.Y-c.Y), 2)
	return math.Sqrt(xDist + yDist)
}

func (c Coord) GetSquareInDirection(direction Direction) *Coord {
	switch direction {
	case UP:
		return &Coord{X: c.X, Y: c.Y + 1}
	case DOWN:
		return &Coord{X: c.X, Y: c.Y - 1}
	case LEFT:
		return &Coord{X: c.X - 1, Y: c.Y}
	case RIGHT:
		return &Coord{X: c.X + 1, Y: c.Y}
	default:
		log.Error("Invalid direction")
		return nil
	}
}
