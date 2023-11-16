package services

import (
	"reflect"
	"testing"

	"BattlesnakeReptarium/internal/model"
)

func TestExcludeOccupiedSquaresFromOptions(t *testing.T) {
	options := []model.Direction{model.UP, model.LEFT, model.DOWN, model.RIGHT}
	tests := []struct {
		name     string
		selfHead model.Coord
		board    model.Board
		want     []model.Direction
	}{
		{"no snakes, bottom left coord", model.Coord{X: 0, Y: 0}, model.Board{Height: 10, Width: 10},
			[]model.Direction{model.UP, model.RIGHT}},
		{"no snakes, top right coord", model.Coord{X: 9, Y: 9}, model.Board{Height: 10, Width: 10},
			[]model.Direction{model.LEFT, model.DOWN}},
		{"no snakes, middle coord", model.Coord{X: 5, Y: 5}, model.Board{Height: 10, Width: 10},
			[]model.Direction{model.UP, model.LEFT, model.DOWN, model.RIGHT}},
		{"snake to the left", model.Coord{X: 5, Y: 5}, model.Board{Height: 10, Width: 10, Snakes: []model.Snake{
			{Body: []model.Coord{{X: 4, Y: 5}}}}}, []model.Direction{model.UP, model.DOWN, model.RIGHT}},
		{"snake head to the left", model.Coord{X: 5, Y: 5}, model.Board{Height: 10, Width: 10, Snakes: []model.Snake{
			{Head: model.Coord{X: 4, Y: 5}}}}, []model.Direction{model.UP, model.DOWN, model.RIGHT}},
		{"No exclusions", model.Coord{X: 5, Y: 5}, model.Board{Height: 10, Width: 10},
			[]model.Direction{model.UP, model.LEFT, model.DOWN, model.RIGHT}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := excludeOccupiedCoordsFromOptions(options, tt.selfHead, tt.board); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("excludeOccupiedCoordsFromOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExcludeCoordsAnySnakeIsHeadingFor(t *testing.T) {
	t.Run("no snakes", func(t *testing.T) {
		options := []model.Direction{model.UP, model.LEFT, model.DOWN, model.RIGHT}
		board := model.Board{Height: 10, Width: 10}
		selfSnake := model.Snake{Head: model.Coord{X: 5, Y: 5}}
		if got := excludeCoordsAnySnakeIsHeadingFor(options, selfSnake, board); !reflect.DeepEqual(got, options) {
			t.Errorf("excludeCoordsAnySnakeIsHeadingFor() = %v, want %v", got, options)
		}
	})
}
