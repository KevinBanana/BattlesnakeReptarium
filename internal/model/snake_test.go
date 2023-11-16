package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindSnakeTravelDirection(t *testing.T) {
	tests := []struct {
		name   string
		snake  Snake
		expect Direction
	}{
		{"Up", Snake{Head: Coord{X: 1, Y: 1}, Body: []Coord{{X: 1, Y: 0}}}, UP},
		{"Down", Snake{Head: Coord{X: 1, Y: 1}, Body: []Coord{{X: 1, Y: 2}}}, DOWN},
		{"Left", Snake{Head: Coord{X: 1, Y: 1}, Body: []Coord{{X: 2, Y: 1}}}, LEFT},
		{"Right", Snake{Head: Coord{X: 1, Y: 1}, Body: []Coord{{X: 0, Y: 1}}}, RIGHT},
		{"No direction", Snake{Head: Coord{X: 1, Y: 1}, Body: []Coord{{X: 1, Y: 1}}}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.snake.FindSnakeTravelDirection())
		})
	}
}
