package model

import "math"

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
