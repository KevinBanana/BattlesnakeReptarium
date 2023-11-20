package services

import (
	"context"
	"sync"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/util"
)

type BananatronV1Svc struct {
	mux sync.RWMutex
}

func NewBananatronSvc() *BananatronV1Svc {
	return &BananatronV1Svc{}
}

func (svc *BananatronV1Svc) CalculateMove(ctx context.Context, game model.Game, turn int, board model.Board, selfSnake model.Snake) (*model.SnakeAction, error) {
	weightedOptions := map[model.Direction]float64{model.UP: 0, model.LEFT: 0, model.DOWN: 0, model.RIGHT: 0}
	wg := new(sync.WaitGroup)

	// TODO consider when an enemy snake only has one option

	wg.Add(1)
	go svc.adjustWeightsForOccupiedSquares(wg, &weightedOptions, selfSnake.Head, board)

	//options = excludeCoordsAnySnakeIsHeadingFor(options, selfSnake, board)

	// Return the highest weighted option
	// TODO
	wg.Wait()
	return &model.SnakeAction{Move: model.UP}, nil
}

func (svc *BananatronV1Svc) adjustWeightsForOccupiedSquares(wg *sync.WaitGroup, options *map[model.Direction]float64, selfHead model.Coord, board model.Board) {
	defer wg.Done()

	for direction, _ := range *options {
		targetSquare := selfHead.GetSquareInDirection(direction)
		if !board.IsCoordClear(*targetSquare) {
			// Coord is occupied, penalize option
			svc.mux.Lock()
			(*options)[direction] -= 100
			svc.mux.Unlock()
		}
	}

}

func excludeCoordsAnySnakeIsHeadingFor(options []model.Direction, selfSnake model.Snake, board model.Board) []model.Direction {
	var filteredOptions []model.Direction
	var futureOccupiedCoords []model.Coord // An estimate based on where enemy snakes are heading if they go straight
	for _, snake := range board.Snakes {
		// Exclude self
		if snake.ID == selfSnake.ID {
			continue
		}
		snakeTravelDirection := snake.FindSnakeTravelDirection()
		nextOccupiedCoord := snake.Head.GetSquareInDirection(snakeTravelDirection)
		if nextOccupiedCoord != nil {
			futureOccupiedCoords = append(futureOccupiedCoords, *nextOccupiedCoord)
		}
	}

	for _, move := range options {
		targetSquare := selfSnake.Head.GetSquareInDirection(move)
		if !util.Contains(futureOccupiedCoords, *targetSquare) {
			filteredOptions = append(filteredOptions, move)
		}
	}

	return filteredOptions
}
