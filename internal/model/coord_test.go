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
	t.Run("Test horizontal distance", func(t *testing.T) {
		dist := x0y0.GetDistance(x1y0)
		assert.Equal(t, 1.0, dist)
	})

	t.Run("Test vertical distance", func(t *testing.T) {
		dist := x0y0.GetDistance(x0y1)
		assert.Equal(t, 1.0, dist)
	})

	t.Run("Test diagonal distance", func(t *testing.T) {
		dist := x0y0.GetDistance(x1y1)
		assert.Equal(t, math.Sqrt(2), dist)
	})
}
