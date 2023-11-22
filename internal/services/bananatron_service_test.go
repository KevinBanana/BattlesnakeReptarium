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
			map[model.Direction]float64{model.UP: 0, model.LEFT: -OccupiedSquarePenalty, model.DOWN: -OccupiedSquarePenalty, model.RIGHT: 0}},
		{"no snakes, top right coord", model.Coord{X: 9, Y: 9}, model.Board{Height: 10, Width: 10},
			map[model.Direction]float64{model.UP: -OccupiedSquarePenalty, model.LEFT: 0, model.DOWN: 0, model.RIGHT: -OccupiedSquarePenalty}},
		{"no snakes, middle coord", model.Coord{X: 5, Y: 5}, model.Board{Height: 10, Width: 10},
			map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}},
		{"snake to the left", model.Coord{X: 5, Y: 5}, model.Board{Height: 10, Width: 10, Snakes: []model.Snake{
			{Body: []model.Coord{{X: 4, Y: 5}}}}}, map[model.Direction]float64{model.UP: 0, model.LEFT: -OccupiedSquarePenalty, model.DOWN: 0, model.RIGHT: 0}},
		{"snake head to the left", model.Coord{X: 5, Y: 5}, model.Board{Height: 10, Width: 10, Snakes: []model.Snake{
			{Head: model.Coord{X: 4, Y: 5}}}}, map[model.Direction]float64{model.UP: 0, model.LEFT: -OccupiedSquarePenalty, model.DOWN: 0, model.RIGHT: 0}},
		{"No exclusions", model.Coord{X: 5, Y: 5}, model.Board{Height: 10, Width: 10},
			map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}
			wg.Add(1)
			svc.adjustWeightsForOccupiedSquares(wg, &options, tt.selfHead, tt.board)
			wg.Wait()
			if !reflect.DeepEqual(options, tt.want) {
				t.Errorf("adjustWeightsForOccupiedSquares() = %v, want %v", options, tt.want)
			}
		})
	}
}

func TestAdjustWeightsForCollisionCourse(t *testing.T) {
	svc := NewBananatronSvc()
	wg := new(sync.WaitGroup)

	t.Run("no snakes", func(t *testing.T) {
		options := map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}
		want := map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}
		board := model.Board{Height: 10, Width: 10}
		selfSnake := model.Snake{Head: model.Coord{X: 5, Y: 5}, ID: "self"}

		wg.Add(1)
		svc.adjustWeightsForCollisionCourse(wg, &options, selfSnake, board)
		wg.Wait()
		if !reflect.DeepEqual(options, want) {
			t.Errorf("adjustWeightsForCollisionCourse() = %v, want %v", options, want)
		}
	})

	t.Run("Avoid head on collision", func(t *testing.T) {
		options := map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}
		want := map[model.Direction]float64{model.UP: 0, model.LEFT: -CollisionCoursePenalty, model.DOWN: 0, model.RIGHT: 0}
		selfSnake := model.Snake{Head: model.Coord{X: 2, Y: 10}, ID: "self"}
		board := model.Board{Height: 11, Width: 11, Snakes: []model.Snake{snakeHeadingTox1y10, selfSnake}}

		wg.Add(1)
		svc.adjustWeightsForCollisionCourse(wg, &options, selfSnake, board)
		wg.Wait()
		if !reflect.DeepEqual(options, want) {
			t.Errorf("adjustWeightsForCollisionCourse() = %v, want %v", options, want)
		}
	})
}

func TestAdjustWeightsForCavernSize(t *testing.T) {
	svc := NewBananatronSvc()
	wg := new(sync.WaitGroup)

	t.Run("Avoid smaller cavern", func(t *testing.T) {
		options := map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}
		snakes := []model.Snake{
			{ID: "enemySnake", Head: model.Coord{X: 1, Y: 2}, Body: []model.Coord{{X: 1, Y: 1}, {X: 1, Y: 2}, {X: 2, Y: 1}}},
			{ID: "selfSnake", Head: model.Coord{X: 0, Y: 1}},
		}
		board := model.Board{
			Height:  3,
			Width:   3,
			Food:    nil,
			Hazards: nil,
			Snakes:  snakes,
		}

		wg.Add(1)
		svc.adjustWeightsForCavernSize(wg, &options, snakes[1].Head, board)
		wg.Wait()
		want := map[model.Direction]float64{model.UP: 1, model.LEFT: 0, model.DOWN: 3, model.RIGHT: 0}
		if !reflect.DeepEqual(options, want) {
			t.Errorf("adjustWeightsForCavernSize() = %v, want %v", options, want)
		}
	})
}

func TestAdjustWeightsForAvoidingCorneredSnakes(t *testing.T) {

}

func TestDetermineSnakeAction(t *testing.T) {
	tests := []struct {
		name            string
		weightedOptions map[model.Direction]float64
		want            model.Direction
	}{
		{"One positive weight", map[model.Direction]float64{model.UP: 1, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}, model.UP},
		{"Three negative weights", map[model.Direction]float64{model.UP: -1, model.LEFT: -1, model.DOWN: -1, model.RIGHT: 0}, model.RIGHT},
		{"Two positive weights", map[model.Direction]float64{model.UP: 1, model.LEFT: 0, model.DOWN: 2, model.RIGHT: 0}, model.DOWN},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := determineSnakeAction(tt.weightedOptions).Move; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("determineSnakeAction() = %v, want %v", got, tt.want)
			}
		})
	}
}
