package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSnakeOnBoard(t *testing.T) {
	tests := []struct {
		name   string
		snake  Snake
		board  Board
		expect bool
	}{
		{"Snake is on board", Snake{ID: "snake1"}, Board{Snakes: []Snake{{ID: "someSnake"}, {ID: "snake1"}}}, true},
		{"Snake is not on board", Snake{ID: "snake2"}, Board{Snakes: []Snake{{ID: "someSnake"}, {ID: "snake1"}}}, false},
		{"Snake has no ID", Snake{}, Board{Snakes: []Snake{{ID: "snake1"}}}, false},
		{"Board has no snakes", Snake{ID: "snake1"}, Board{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.board.IsSnakeOnBoard(tt.snake))
		})
	}
}

func TestIsCoordClear(t *testing.T) {
	tests := []struct {
		name   string
		coord  Coord
		board  Board
		expect bool
	}{
		{"Coord is clear", Coord{X: 1, Y: 1}, Board{Width: 2, Height: 2}, true},
		{"Coord is a snake body", Coord{X: 1, Y: 1}, Board{Width: 2, Height: 2, Snakes: []Snake{{Body: []Coord{{X: 1, Y: 1}}}}}, false},
		{"Coord is a snake head", Coord{X: 1, Y: 1}, Board{Width: 2, Height: 2, Snakes: []Snake{{Head: Coord{X: 1, Y: 1}}}}, false},
		{"Coord is a hazard", Coord{X: 1, Y: 1}, Board{Width: 2, Height: 2, Hazards: []Coord{{X: 1, Y: 1}}}, false},
		{"Coord is off the board", Coord{X: 2, Y: 1}, Board{Width: 2, Height: 2}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.board.IsCoordClear(tt.coord))
		})
	}
}

func TestIsCoordOnBoard(t *testing.T) {
	tests := []struct {
		name   string
		coord  Coord
		board  Board
		expect bool
	}{
		{"Coord is on board", Coord{X: 1, Y: 1}, Board{Width: 2, Height: 2}, true},
		{"Coord is too far to the right", Coord{X: 2, Y: 1}, Board{Width: 2, Height: 2}, false},
		{"Coord is too far to the left", Coord{X: -1, Y: 1}, Board{Width: 2, Height: 2}, false},
		{"Coord is too far up", Coord{X: 1, Y: 2}, Board{Width: 2, Height: 2}, false},
		{"Coord is too far down", Coord{X: 1, Y: -1}, Board{Width: 2, Height: 2}, false},
		{"Coord is at origin", Coord{X: 0, Y: 0}, Board{Width: 2, Height: 2}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.board.isCoordOnBoard(tt.coord))
		})
	}
}
