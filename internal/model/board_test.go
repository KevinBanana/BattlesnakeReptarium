package model_test

import (
	"testing"

	"BattlesnakeReptarium/internal/mock"
	"BattlesnakeReptarium/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestIsSnakeOnBoard(t *testing.T) {
	tests := map[string]struct {
		snake  model.Snake
		board  model.Board
		expect bool
	}{
		"Snake is on board":     {model.Snake{ID: "snake1"}, model.Board{Snakes: []model.Snake{{ID: "someSnake"}, {ID: "snake1"}}}, true},
		"Snake is not on board": {model.Snake{ID: "snake2"}, model.Board{Snakes: []model.Snake{{ID: "someSnake"}, {ID: "snake1"}}}, false},
		"Snake has no ID":       {model.Snake{}, model.Board{Snakes: []model.Snake{{ID: "snake1"}}}, false},
		"Board has no snakes":   {model.Snake{ID: "snake1"}, model.Board{}, false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.board.IsSnakeOnBoard(tt.snake))
		})
	}
}

func TestIsCoordClear(t *testing.T) {
	tests := map[string]struct {
		coord  model.Coord
		board  model.Board
		expect bool
	}{
		"Coord is clear":         {model.Coord{X: 1, Y: 1}, model.Board{Width: 2, Height: 2}, true},
		"Coord is a snake body":  {model.Coord{X: 1, Y: 1}, model.Board{Width: 2, Height: 2, Snakes: []model.Snake{{Body: []model.Coord{{X: 1, Y: 1}}}}}, false},
		"Coord is a snake head":  {model.Coord{X: 1, Y: 1}, model.Board{Width: 2, Height: 2, Snakes: []model.Snake{{Head: model.Coord{X: 1, Y: 1}}}}, false},
		"Coord is a hazard":      {model.Coord{X: 1, Y: 1}, model.Board{Width: 2, Height: 2, Hazards: []model.Coord{{X: 1, Y: 1}}}, false},
		"Coord is off the board": {model.Coord{X: 2, Y: 1}, model.Board{Width: 2, Height: 2}, false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.board.IsCoordClear(tt.coord))
		})
	}
}

func TestIsCoordOnBoard(t *testing.T) {
	tests := map[string]struct {
		coord  model.Coord
		board  model.Board
		expect bool
	}{
		"Coord is on board":             {model.Coord{X: 1, Y: 1}, model.Board{Width: 2, Height: 2}, true},
		"Coord is too far to the right": {model.Coord{X: 2, Y: 1}, model.Board{Width: 2, Height: 2}, false},
		"Coord is too far to the left":  {model.Coord{X: -1, Y: 1}, model.Board{Width: 2, Height: 2}, false},
		"Coord is too far up":           {model.Coord{X: 1, Y: 2}, model.Board{Width: 2, Height: 2}, false},
		"Coord is too far down":         {model.Coord{X: 1, Y: -1}, model.Board{Width: 2, Height: 2}, false},
		"Coord is at origin":            {model.Coord{X: 0, Y: 0}, model.Board{Width: 2, Height: 2}, true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.board.IsCoordOnBoard(tt.coord))
		})
	}
}

func TestDetermineFloodFillCoords(t *testing.T) {
	t.Run("Flood fill from origin", func(t *testing.T) {
		board := model.Board{Width: 3, Height: 3}
		snakes := []model.Snake{
			{
				Body: []model.Coord{{X: 1, Y: 1}, {X: 2, Y: 1}},
				Head: model.Coord{X: 1, Y: 2},
			},
		}
		board.Snakes = snakes

		got := board.DetermineFloodFillCoords(model.Coord{X: 0, Y: 0})
		assert.Equal(t, 5, len(got))
		want := []model.Coord{{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 0, Y: 2}, {X: 1, Y: 0}, {X: 2, Y: 0}}
		if !unorderedEqual(got, want) {
			t.Errorf("DetermineFloodFillCoords got = %v, want %v", got, want)
		}
	})

	t.Run("Flood fill from origin with hazard", func(t *testing.T) {
		board := model.Board{Width: 3, Height: 3, Hazards: []model.Coord{{X: 0, Y: 0}}}
		snakes := []model.Snake{{Body: []model.Coord{{X: 1, Y: 1}, {X: 1, Y: 2}, {X: 2, Y: 1}}}}
		board.Snakes = snakes

		got := board.DetermineFloodFillCoords(model.Coord{X: 0, Y: 0})
		assert.Equal(t, 0, len(got))
	})
}

func TestFindAllSnakesInCavern(t *testing.T) {
	tests := map[string]struct {
		board            model.Board
		cavernCoords     []model.Coord
		expectedSnakeIDs []string
	}{
		"no snakes in cavern": {
			board:        mock.NewBoard(),
			cavernCoords: []model.Coord{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}},
		},
		"one snake in cavern": {
			board: mock.NewBoard(mock.WithBoardSnakes([]model.Snake{
				{
					ID:   "snake1",
					Head: model.Coord{X: 1, Y: 1},
					Body: []model.Coord{{X: 1, Y: 2}},
				},
				{
					ID:   "snake2",
					Head: model.Coord{X: 2, Y: 2},
					Body: []model.Coord{{X: 2, Y: 3}},
				},
			})),
			cavernCoords:     []model.Coord{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}},
			expectedSnakeIDs: []string{"snake1"},
		},
		"multiple snakes in cavern": {
			board: mock.NewBoard(mock.WithBoardSnakes([]model.Snake{
				{
					ID:   "snake1",
					Head: model.Coord{X: 1, Y: 1},
					Body: []model.Coord{{X: 1, Y: 2}},
				},
				{
					ID:   "snake2",
					Head: model.Coord{X: 4, Y: 3},
					Body: []model.Coord{{X: 4, Y: 4}},
				},
				{
					ID:   "snake3",
					Head: model.Coord{X: 0, Y: 3},
					Body: []model.Coord{{X: 0, Y: 4}},
				},
			})),
			cavernCoords:     []model.Coord{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 0, Y: 2}},
			expectedSnakeIDs: []string{"snake1", "snake3"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.board.FindAllSnakesInCavern(tt.cavernCoords)
			for _, snake := range got {
				assert.Contains(t, tt.expectedSnakeIDs, snake.ID)
			}
		})
	}
}

func unorderedEqual(first, second []model.Coord) bool {
	if len(first) != len(second) {
		return false
	}
	exists := make(map[model.Coord]bool)
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
