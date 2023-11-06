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
