package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindSnakeTravelDirection(t *testing.T) {
	tests := []struct {
		name     string
		snake    Snake
		expect   Direction
		expectOK bool
	}{
		{"Up", Snake{Head: Coord{X: 1, Y: 1}, Body: []Coord{{X: 1, Y: 1}, {X: 1, Y: 0}}}, UP, true},
		{"Down", Snake{Head: Coord{X: 1, Y: 1}, Body: []Coord{{X: 1, Y: 1}, {X: 1, Y: 2}}}, DOWN, true},
		{"Left", Snake{Head: Coord{X: 1, Y: 1}, Body: []Coord{{X: 1, Y: 1}, {X: 2, Y: 1}}}, LEFT, true},
		{"Right", Snake{Head: Coord{X: 1, Y: 1}, Body: []Coord{{X: 1, Y: 1}, {X: 0, Y: 1}}}, RIGHT, true},
		// Every snake starts a game with its body stacked on one square.
		{"Stacked body at game start", Snake{Head: Coord{X: 1, Y: 1}, Body: []Coord{{X: 1, Y: 1}, {X: 1, Y: 1}}}, "", false},
		{"No body", Snake{Head: Coord{X: 1, Y: 1}, Body: []Coord{{X: 1, Y: 1}}}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			direction, ok := tt.snake.TravelDirection()
			assert.Equal(t, tt.expect, direction)
			assert.Equal(t, tt.expectOK, ok)
		})
	}
}
