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

func TestDetermineFloodFillCoords(t *testing.T) {
	t.Run("Flood fill from origin", func(t *testing.T) {
		board := Board{Width: 3, Height: 3}
		snakes := []Snake{
			{
				Body: []Coord{{X: 1, Y: 1}, {X: 2, Y: 1}},
				Head: Coord{X: 1, Y: 2},
			},
		}
		board.Snakes = snakes

		got := board.DetermineFloodFillCoords(Coord{X: 0, Y: 0})
		assert.Equal(t, 5, len(got))
		want := []Coord{{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 0, Y: 2}, {X: 1, Y: 0}, {X: 2, Y: 0}}
		if !unorderedEqual(got, want) {
			t.Errorf("DetermineFloodFillCoords got = %v, want %v", got, want)
		}
	})

	t.Run("Flood fill from origin with hazard", func(t *testing.T) {
		board := Board{Width: 3, Height: 3, Hazards: []Coord{{X: 0, Y: 0}}}
		snakes := []Snake{{Body: []Coord{{X: 1, Y: 1}, {X: 1, Y: 2}, {X: 2, Y: 1}}}}
		board.Snakes = snakes

		got := board.DetermineFloodFillCoords(Coord{X: 0, Y: 0})
		assert.Equal(t, 0, len(got))
	})
}

func unorderedEqual(first, second []Coord) bool {
	if len(first) != len(second) {
		return false
	}
	exists := make(map[Coord]bool)
	for _, value := range first {
		exists[value] = true
	}
	for _, value := range second {
		if !exists[value] {
			return false
		}
	}
	return true
}
