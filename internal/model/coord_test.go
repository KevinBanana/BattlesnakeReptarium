package model

import (
	"math"
	"testing"
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
			if got != tc.expectedDistance {
				t.Errorf("Expected %f, got %f", tc.expectedDistance, got)
			}
		})
	}
}
