package model

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	x0y0 = Coord{X: 0, Y: 0}
	x0y1 = Coord{X: 0, Y: 1}
	x1y0 = Coord{X: 1, Y: 0}
	x1y1 = Coord{X: 1, Y: 1}
)

func TestCoord_GetDistance(t *testing.T) {
	tests := []struct {
		name             string
		location1        Coord
		location2        Coord
		expectedDistance float64
	}{
		{"Test horizontal distance", x0y0, x1y0, 1.0},
		{"Test vertical distance", x0y0, x0y1, 1.0},
		{"Test diagonal distance", x0y0, x1y1, math.Sqrt(2)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.location1.GetDistance(tc.location2)
			assert.Equal(t, tc.expectedDistance, got)
		})
	}
}

func TestCoord_GetSquareInDirection(t *testing.T) {
	tests := []struct {
		name      string
		start     Coord
		direction Direction
		expected  *Coord
	}{
		{"UP", x0y0, UP, &x0y1},
		{"DOWN", x0y1, DOWN, &x0y0},
		{"LEFT", x1y0, LEFT, &x0y0},
		{"RIGHT", x0y0, RIGHT, &x1y0},
		{"Invalid", x0y0, Direction("invalid"), nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.start.GetSquareInDirection(tc.direction)
			assert.Equal(t, tc.expected, got)
		})
	}
}
