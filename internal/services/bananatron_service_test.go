package services

import (
	"reflect"
	"sync"
	"testing"

	"BattlesnakeReptarium/internal/mock"
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
			map[model.Direction]float64{model.UP: 0, model.LEFT: OccupiedSquarePenalty, model.DOWN: OccupiedSquarePenalty, model.RIGHT: 0}},
		{"no snakes, top right coord", model.Coord{X: 9, Y: 9}, model.Board{Height: 10, Width: 10},
			map[model.Direction]float64{model.UP: OccupiedSquarePenalty, model.LEFT: 0, model.DOWN: 0, model.RIGHT: OccupiedSquarePenalty}},
		{"no snakes, middle coord", model.Coord{X: 5, Y: 5}, model.Board{Height: 10, Width: 10},
			map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}},
		{"snake to the left", model.Coord{X: 5, Y: 5}, model.Board{Height: 10, Width: 10, Snakes: []model.Snake{
			{Body: []model.Coord{{X: 4, Y: 5}}}}}, map[model.Direction]float64{model.UP: 0, model.LEFT: OccupiedSquarePenalty, model.DOWN: 0, model.RIGHT: 0}},
		{"snake head to the left", model.Coord{X: 5, Y: 5}, model.Board{Height: 10, Width: 10, Snakes: []model.Snake{
			{Head: model.Coord{X: 4, Y: 5}}}}, map[model.Direction]float64{model.UP: 0, model.LEFT: OccupiedSquarePenalty, model.DOWN: 0, model.RIGHT: 0}},
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
		want := map[model.Direction]float64{model.UP: 0, model.LEFT: CollisionCoursePenalty, model.DOWN: 0, model.RIGHT: 0}
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

	tests := map[string]struct {
		selfHeadCoord model.Coord
		board         model.Board
		want          map[model.Direction]float64
	}{
		"Avoid smaller cavern": {
			selfHeadCoord: model.Coord{X: 0, Y: 1},
			board: mock.NewBoard(mock.WithBoardHeight(3), mock.WithBoardWidth(3), mock.WithBoardSnakes([]model.Snake{
				{ID: "enemySnake", Head: model.Coord{X: 1, Y: 2}, Body: []model.Coord{{X: 1, Y: 1}, {X: 2, Y: 1}}},
				{ID: "selfSnake", Head: model.Coord{X: 0, Y: 1}},
			})),
			want: map[model.Direction]float64{model.UP: .5, model.LEFT: 0, model.DOWN: 3, model.RIGHT: 0},
		},
		"Choose cavern even though it has snake": {
			selfHeadCoord: model.Coord{X: 0, Y: 1},
			board: mock.NewBoard(mock.WithBoardHeight(3), mock.WithBoardWidth(4), mock.WithBoardSnakes([]model.Snake{
				{ID: "enemySnake", Head: model.Coord{X: 3, Y: 2}, Body: []model.Coord{{X: 2, Y: 2}, {X: 1, Y: 2}}},
				{ID: "selfSnake", Head: model.Coord{X: 0, Y: 1}, Body: []model.Coord{{X: 0, Y: 0}}},
			})),
			want: map[model.Direction]float64{model.UP: 1, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 3},
		},
		"Avoid spacious cavern because it has more snakes": {
			selfHeadCoord: model.Coord{X: 0, Y: 0},
			board: mock.NewBoard(mock.WithBoardHeight(5), mock.WithBoardWidth(4), mock.WithBoardSnakes([]model.Snake{
				{ID: "enemySnake", Head: model.Coord{X: 3, Y: 2}, Body: []model.Coord{{X: 3, Y: 3}, {X: 2, Y: 3}, {X: 1, Y: 3}, {X: 1, Y: 1}, {X: 1, Y: 2}, {X: 1, Y: 4}}},
				{ID: "selfSnake", Head: model.Coord{X: 0, Y: 0}},
			})),
			want: map[model.Direction]float64{model.UP: 4, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 3},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			options := map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}
			wg.Add(1)
			svc.adjustWeightsForCavernSize(wg, &options, tt.selfHeadCoord, tt.board)
			wg.Wait()
			if !reflect.DeepEqual(options, tt.want) {
				t.Errorf("adjustWeightsForCavernSize() = %v, want %v", options, tt.want)
			}
		})
	}
}

func TestAdjustWeightsForAvoidingCorneredSnakes(t *testing.T) {
	svc := NewBananatronSvc()
	wg := new(sync.WaitGroup)

	t.Run("Avoid cornered snake", func(t *testing.T) {
		options := map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}
		snakes := []model.Snake{
			// Enemy snake must move left
			{ID: "enemySnake", Head: model.Coord{X: 1, Y: 2}, Body: []model.Coord{{X: 1, Y: 2}, {X: 2, Y: 2}, {X: 2, Y: 1}, {X: 1, Y: 1}}},
			// Self snake can move up or down
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
		svc.adjustWeightsForAvoidingCorneredSnakes(wg, &options, snakes[1], board)
		wg.Wait()
		want := map[model.Direction]float64{model.UP: CorneredSnakeEscapeSquarePenalty, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}
		if !reflect.DeepEqual(options, want) {
			t.Errorf("adjustWeightsForAvoidingCorneredSnakes() = %v, want %v", options, want)
		}
	})
}

func TestAdjustWeightsForPotentialEnemyMove(t *testing.T) {
	svc := NewBananatronSvc()
	wg := new(sync.WaitGroup)

	tests := map[string]struct {
		selfSnake model.Snake
		board     model.Board
		want      map[model.Direction]float64
	}{
		"Enemy snake can move to two potential move": {
			selfSnake: model.Snake{ID: "selfSnake", Head: model.Coord{X: 1, Y: 0}},
			board: mock.NewBoard(mock.WithBoardHeight(3), mock.WithBoardWidth(3), mock.WithBoardSnakes([]model.Snake{
				{ID: "enemySnake", Head: model.Coord{X: 0, Y: 1}},
				{ID: "selfSnake", Head: model.Coord{X: 1, Y: 0}},
			})),
			want: map[model.Direction]float64{model.UP: -4, model.LEFT: -4, model.DOWN: 0, model.RIGHT: 0},
		},
		"Enemy snake can move to one potential move": {
			selfSnake: model.Snake{ID: "selfSnake", Head: model.Coord{X: 1, Y: 0}},
			board: mock.NewBoard(mock.WithBoardHeight(3), mock.WithBoardWidth(3), mock.WithBoardSnakes([]model.Snake{
				{ID: "enemySnake", Head: model.Coord{X: 1, Y: 2}},
				{ID: "selfSnake", Head: model.Coord{X: 1, Y: 0}},
			})),
			want: map[model.Direction]float64{model.UP: -4, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0},
		},
		"Enemy snake can move to no potential move": {
			selfSnake: model.Snake{ID: "selfSnake", Head: model.Coord{X: 0, Y: 0}},
			board: mock.NewBoard(mock.WithBoardHeight(3), mock.WithBoardWidth(3), mock.WithBoardSnakes([]model.Snake{
				{ID: "enemySnake", Head: model.Coord{X: 2, Y: 2}},
				{ID: "selfSnake", Head: model.Coord{X: 0, Y: 0}},
			})),
			want: map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			options := map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}
			wg.Add(1)
			svc.adjustWeightsForPotentialEnemyMove(wg, &options, tt.selfSnake, tt.board)
			wg.Wait()
			if !reflect.DeepEqual(options, tt.want) {
				t.Errorf("adjustWeightsForPotentialEnemyMove() = %v, want %v", options, tt.want)
			}
		})
	}
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
