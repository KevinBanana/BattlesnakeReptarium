package util

import (
	"testing"

	"BattlesnakeReptarium/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestIsSnakeOnBoard(t *testing.T) {
	tests := []struct {
		name   string
		snake  model.Snake
		board  model.Board
		expect bool
	}{
		{"Snake is on board", model.Snake{ID: "snake1"}, model.Board{Snakes: []model.Snake{{ID: "someSnake"}, {ID: "snake1"}}}, true},
		{"Snake is not on board", model.Snake{ID: "snake2"}, model.Board{Snakes: []model.Snake{{ID: "someSnake"}, {ID: "snake1"}}}, false},
		{"Snake has no ID", model.Snake{}, model.Board{Snakes: []model.Snake{{ID: "snake1"}}}, false},
		{"Board has no snakes", model.Snake{ID: "snake1"}, model.Board{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, IsSnakeOnBoard(tt.snake, tt.board))
		})
	}
}

func TestCalculateMovesDistance(t *testing.T) {
	test := []struct {
		name   string
		a      model.Coord
		b      model.Coord
		expect int
	}{
		{"Horizontal only", model.Coord{X: 1, Y: 1}, model.Coord{X: 5, Y: 1}, 4},
		{"Horizontal only, negative", model.Coord{X: 5, Y: 1}, model.Coord{X: 1, Y: 1}, 4},
		{"Vertical only", model.Coord{X: 1, Y: 1}, model.Coord{X: 1, Y: 5}, 4},
		{"Vertical only, negative", model.Coord{X: 1, Y: 5}, model.Coord{X: 1, Y: 1}, 4},
		{"Horizontal and vertical", model.Coord{X: 1, Y: 1}, model.Coord{X: 5, Y: 5}, 8},
		{"Horizontal and vertical, negative", model.Coord{X: 5, Y: 5}, model.Coord{X: 1, Y: 1}, 8},
		{"Same coord", model.Coord{X: 1, Y: 1}, model.Coord{X: 1, Y: 1}, 0},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, calculateMovesDistance(tt.a, tt.b))
		})
	}
}
