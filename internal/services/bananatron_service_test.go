package services

import (
	"reflect"
	"sync"
	"testing"

	"BattlesnakeReptarium/internal/model"
)

func TestAdjustWeightsForOccupiedSquares(t *testing.T) {
	svc := NewBananatronSvc()
	wg := new(sync.WaitGroup)

	tests := []struct {
		name     string
		selfHead model.Coord
		board    model.Board
		want     map[model.Direction]float64
	}{
		{"no snakes, bottom left coord", model.Coord{X: 0, Y: 0}, model.Board{Height: 10, Width: 10},
			map[model.Direction]float64{model.UP: 0, model.LEFT: -100, model.DOWN: -100, model.RIGHT: 0}},
		{"no snakes, top right coord", model.Coord{X: 9, Y: 9}, model.Board{Height: 10, Width: 10},
			map[model.Direction]float64{model.UP: -100, model.LEFT: 0, model.DOWN: 0, model.RIGHT: -100}},
		{"no snakes, middle coord", model.Coord{X: 5, Y: 5}, model.Board{Height: 10, Width: 10},
			map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}},
		{"snake to the left", model.Coord{X: 5, Y: 5}, model.Board{Height: 10, Width: 10, Snakes: []model.Snake{
			{Body: []model.Coord{{X: 4, Y: 5}}}}}, map[model.Direction]float64{model.UP: 0, model.LEFT: -100, model.DOWN: 0, model.RIGHT: 0}},
		{"snake head to the left", model.Coord{X: 5, Y: 5}, model.Board{Height: 10, Width: 10, Snakes: []model.Snake{
			{Head: model.Coord{X: 4, Y: 5}}}}, map[model.Direction]float64{model.UP: 0, model.LEFT: -100, model.DOWN: 0, model.RIGHT: 0}},
		{"No exclusions", model.Coord{X: 5, Y: 5}, model.Board{Height: 10, Width: 10},
			map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}
			wg.Add(1)
			svc.adjustWeightsForOccupiedSquares(wg, &options, tt.selfHead, tt.board)
			if !reflect.DeepEqual(options, tt.want) {
				t.Errorf("adjustWeightsForOccupiedSquares() = %v, want %v", options, tt.want)
			}
		})
	}
}

func TestExcludeCoordsAnySnakeIsHeadingFor(t *testing.T) {
	t.Run("no snakes", func(t *testing.T) {
		options := []model.Direction{model.UP, model.LEFT, model.DOWN, model.RIGHT}
		board := model.Board{Height: 10, Width: 10}
		selfSnake := model.Snake{Head: model.Coord{X: 5, Y: 5}, ID: "self"}
		if got := excludeCoordsAnySnakeIsHeadingFor(options, selfSnake, board); !reflect.DeepEqual(got, options) {
			t.Errorf("excludeCoordsAnySnakeIsHeadingFor() = %v, want %v", got, options)
		}
	})

	t.Run("Avoid head on collision", func(t *testing.T) {
		options := []model.Direction{model.LEFT, model.DOWN}
		selfSnake := model.Snake{Head: model.Coord{X: 2, Y: 10}, ID: "self"}
		board := model.Board{Height: 11, Width: 11, Snakes: []model.Snake{snakeHeadingTox1y10, selfSnake}}
		want := []model.Direction{model.DOWN}
		if got := excludeCoordsAnySnakeIsHeadingFor(options, selfSnake, board); !reflect.DeepEqual(got, want) {
			t.Errorf("excludeCoordsAnySnakeIsHeadingFor() = %v, want %v", got, want)
		}
	})
}
