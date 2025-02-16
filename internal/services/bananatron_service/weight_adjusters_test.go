package bananatron_service

import (
	"reflect"
	"sync"
	"testing"

	"BattlesnakeReptarium/internal/mock"
	"BattlesnakeReptarium/internal/model"
)

func TestOccupiedSquaresAdjuster(t *testing.T) {
	tests := map[string]struct {
		selfSnake model.Snake
		board     model.Board
		want      map[model.Direction]float64
	}{
		"no snakes, bottom left coord": {
			selfSnake: model.Snake{Head: model.Coord{X: 0, Y: 0}},
			board:     mock.NewBoard(mock.WithBoardHeight(10), mock.WithBoardWidth(10)),
			want:      map[model.Direction]float64{model.UP: 0, model.LEFT: OccupiedSquarePenalty, model.DOWN: OccupiedSquarePenalty, model.RIGHT: 0},
		},
		"no snakes, top right coord": {
			selfSnake: model.Snake{Head: model.Coord{X: 9, Y: 9}},
			board:     mock.NewBoard(mock.WithBoardHeight(10), mock.WithBoardWidth(10)),
			want:      map[model.Direction]float64{model.UP: OccupiedSquarePenalty, model.LEFT: 0, model.DOWN: 0, model.RIGHT: OccupiedSquarePenalty},
		},
		"no snakes, middle coord": {
			selfSnake: model.Snake{Head: model.Coord{X: 5, Y: 5}},
			board:     mock.NewBoard(mock.WithBoardHeight(10), mock.WithBoardWidth(10)),
			want:      map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0},
		},
		"snake to the left": {
			selfSnake: model.Snake{Head: model.Coord{X: 5, Y: 5}},
			board: mock.NewBoard(mock.WithBoardHeight(10), mock.WithBoardWidth(10), mock.WithBoardSnakes([]model.Snake{
				{ID: "enemySnake", Head: model.Coord{X: 4, Y: 5}},
			})),
			want: map[model.Direction]float64{model.UP: 0, model.LEFT: OccupiedSquarePenalty, model.DOWN: 0, model.RIGHT: 0},
		},
		"snake head to the left": {
			selfSnake: model.Snake{Head: model.Coord{X: 5, Y: 5}},
			board: mock.NewBoard(mock.WithBoardHeight(10), mock.WithBoardWidth(10), mock.WithBoardSnakes([]model.Snake{
				{ID: "enemySnake", Head: model.Coord{X: 4, Y: 5}},
			})),
			want: map[model.Direction]float64{model.UP: 0, model.LEFT: OccupiedSquarePenalty, model.DOWN: 0, model.RIGHT: 0},
		},
		"No exclusions": {
			selfSnake: model.Snake{Head: model.Coord{X: 5, Y: 5}},
			board:     mock.NewBoard(mock.WithBoardHeight(10), mock.WithBoardWidth(10)),
			want:      map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			wg := new(sync.WaitGroup)
			wg.Add(1)

			adjuster := &OccupiedSquaresAdjuster{}
			options := map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}
			adjuster.AdjustWeight(wg, &options, tc.selfSnake, tc.board, &sync.RWMutex{})
			if !reflect.DeepEqual(options, tc.want) {
				t.Errorf("Expected %v but got %v", tc.want, options)
			}
		})
	}
}

func TestCollisionCourseAdjuster(t *testing.T) {
	tests := map[string]struct {
		selfSnake model.Snake
		board     model.Board
		want      map[model.Direction]float64
	}{
		"no snakes": {
			selfSnake: model.Snake{Head: model.Coord{X: 5, Y: 5}},
			board:     mock.NewBoard(mock.WithBoardHeight(10), mock.WithBoardWidth(10)),
			want:      map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0},
		},
		"avoid collision": {
			selfSnake: model.Snake{ID: "self", Head: model.Coord{X: 2, Y: 10}},
			board: mock.NewBoard(mock.WithBoardHeight(11), mock.WithBoardWidth(11), mock.WithBoardSnakes([]model.Snake{
				{ID: "enemy snake", Head: model.Coord{X: 2, Y: 8}, Body: []model.Coord{{X: 2, Y: 8}, {X: 2, Y: 7}}},
				{ID: "self", Head: model.Coord{X: 2, Y: 10}},
			})),
			want: map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: CollisionCoursePenalty, model.RIGHT: 0},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			wg := new(sync.WaitGroup)
			wg.Add(1)

			adjuster := &CollisionCourseAdjuster{}
			options := map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}
			adjuster.AdjustWeight(wg, &options, tc.selfSnake, tc.board, &sync.RWMutex{})
			if !reflect.DeepEqual(options, tc.want) {
				t.Errorf("Expected %v but got %v", tc.want, options)
			}
		})
	}
}

func TestCavernSizeAdjuster(t *testing.T) {
	tests := map[string]struct {
		selfSnake model.Snake
		board     model.Board
		want      map[model.Direction]float64
	}{
		"avoid smaller cavern": {
			selfSnake: model.Snake{Head: model.Coord{X: 0, Y: 1}},
			board: mock.NewBoard(mock.WithBoardHeight(3), mock.WithBoardWidth(3), mock.WithBoardSnakes([]model.Snake{
				{ID: "enemySnake", Head: model.Coord{X: 1, Y: 2}, Body: []model.Coord{{X: 1, Y: 1}, {X: 2, Y: 1}}},
				{ID: "selfSnake", Head: model.Coord{X: 0, Y: 1}},
			})),
			want: map[model.Direction]float64{model.UP: .5, model.LEFT: 0, model.DOWN: 3, model.RIGHT: 0},
		},
		"choose cavern even though it has snake": {
			selfSnake: model.Snake{Head: model.Coord{X: 0, Y: 1}},
			board: mock.NewBoard(mock.WithBoardHeight(3), mock.WithBoardWidth(4), mock.WithBoardSnakes([]model.Snake{
				{ID: "enemySnake", Head: model.Coord{X: 3, Y: 2}, Body: []model.Coord{{X: 2, Y: 2}, {X: 1, Y: 2}}},
				{ID: "selfSnake", Head: model.Coord{X: 0, Y: 1}, Body: []model.Coord{{X: 0, Y: 0}}},
			})),
			want: map[model.Direction]float64{model.UP: 1, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 3},
		},
		"avoid spacious cavern because it has more snakes": {
			selfSnake: model.Snake{Head: model.Coord{X: 0, Y: 0}},
			board: mock.NewBoard(mock.WithBoardHeight(5), mock.WithBoardWidth(4), mock.WithBoardSnakes([]model.Snake{
				{ID: "enemySnake", Head: model.Coord{X: 3, Y: 2}, Body: []model.Coord{{X: 3, Y: 3}, {X: 2, Y: 3}, {X: 1, Y: 3}, {X: 1, Y: 1}, {X: 1, Y: 2}, {X: 1, Y: 4}}},
				{ID: "selfSnake", Head: model.Coord{X: 0, Y: 0}},
			})),
			want: map[model.Direction]float64{model.UP: 4, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 3},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			wg := new(sync.WaitGroup)
			wg.Add(1)

			adjuster := &CavernSizeAdjuster{}
			options := map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}
			adjuster.AdjustWeight(wg, &options, tc.selfSnake, tc.board, &sync.RWMutex{})
			if !reflect.DeepEqual(options, tc.want) {
				t.Errorf("Expected %v but got %v", tc.want, options)
			}
		})
	}
}

func TestAvoidingCorneredSnakesAdjuster(t *testing.T) {
	tests := map[string]struct {
		selfSnake model.Snake
		board     model.Board
		want      map[model.Direction]float64
	}{
		"Avoid cornered snake": {
			selfSnake: model.Snake{Head: model.Coord{X: 0, Y: 1}},
			board: mock.NewBoard(mock.WithBoardHeight(3), mock.WithBoardWidth(3), mock.WithBoardSnakes([]model.Snake{
				{ID: "enemySnake", Head: model.Coord{X: 1, Y: 2}, Body: []model.Coord{{X: 2, Y: 2}, {X: 2, Y: 1}, {X: 1, Y: 1}}},
				{ID: "selfSnake", Head: model.Coord{X: 0, Y: 1}},
			})),
			want: map[model.Direction]float64{model.UP: CorneredSnakeEscapeSquarePenalty, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			wg := new(sync.WaitGroup)
			wg.Add(1)

			adjuster := &AvoidingCorneredSnakesAdjuster{}
			options := map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}
			adjuster.AdjustWeight(wg, &options, tc.selfSnake, tc.board, &sync.RWMutex{})
			if !reflect.DeepEqual(options, tc.want) {
				t.Errorf("Expected %v but got %v", tc.want, options)
			}
		})
	}
}

func TestPotentialEnemyMoveAdjuster(t *testing.T) {
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
			want: map[model.Direction]float64{model.UP: EnemyPotentialMovePenalty, model.LEFT: EnemyPotentialMovePenalty, model.DOWN: 0, model.RIGHT: 0},
		},
		"Enemy snake can move to one potential move": {
			selfSnake: model.Snake{ID: "selfSnake", Head: model.Coord{X: 1, Y: 0}},
			board: mock.NewBoard(mock.WithBoardHeight(3), mock.WithBoardWidth(3), mock.WithBoardSnakes([]model.Snake{
				{ID: "enemySnake", Head: model.Coord{X: 1, Y: 2}},
				{ID: "selfSnake", Head: model.Coord{X: 1, Y: 0}},
			})),
			want: map[model.Direction]float64{model.UP: EnemyPotentialMovePenalty, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0},
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

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			wg := new(sync.WaitGroup)
			wg.Add(1)

			adjuster := &PotentialEnemyMoveAdjuster{}
			options := map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}
			adjuster.AdjustWeight(wg, &options, tc.selfSnake, tc.board, &sync.RWMutex{})
			if !reflect.DeepEqual(options, tc.want) {
				t.Errorf("Expected %v but got %v", tc.want, options)
			}
		})
	}
}
