package services

import (
	"context"
	"fmt"
	"math"
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

const (
	OccupiedSquarePenalty  = 9999
	CollisionCoursePenalty = 15
)

func (svc *BananatronV1Svc) CalculateMove(ctx context.Context, game model.Game, turn int, board model.Board, selfSnake model.Snake) (*model.SnakeAction, error) {
	weightedOptions := map[model.Direction]float64{}
	for i, direction := range model.AllDirections {
		weightedOptions[direction] = float64(i) // Give default weight of i so that snake will prefer CCW movement
	}

	wg := new(sync.WaitGroup)

	// TODO consider when an enemy snake only has one option

	// TODO refactor into some weightsAdjustor interface that each adjustor implements
	wg.Add(1)
	go svc.adjustWeightsForOccupiedSquares(wg, &weightedOptions, selfSnake.Head, board)

	wg.Add(1)
	go svc.adjustWeightsForCollisionCourse(wg, &weightedOptions, selfSnake, board)

	wg.Wait()

	return determineSnakeAction(weightedOptions), nil
}

func determineSnakeAction(weightedOptions map[model.Direction]float64) *model.SnakeAction {
	var highestWeightedDirection model.Direction
	highestWeight := math.Inf(-1)

	for direction, weight := range weightedOptions {
		if weight > highestWeight {
			highestWeightedDirection = direction
			highestWeight = weight
		}
	}

	return &model.SnakeAction{
		Move:  highestWeightedDirection,
		Shout: fmt.Sprintf("Option weight: %v", highestWeight),
	}
}

// adjustWeightsForCavernSize uses flood fill to determine how many open squares are reachable from each direction
func (svc *BananatronV1Svc) adjustWeightsForCavernSize(wg *sync.WaitGroup, weightedOptions *map[model.Direction]float64, selfHead model.Coord, board model.Board) {
	defer wg.Done()

	// TODO
}

func (svc *BananatronV1Svc) adjustWeightsForOccupiedSquares(wg *sync.WaitGroup, weightedOptions *map[model.Direction]float64, selfHead model.Coord, board model.Board) {
	defer wg.Done()

	for _, direction := range model.AllDirections {
		targetSquare := selfHead.GetSquareInDirection(direction)
		if !board.IsCoordClear(*targetSquare) {
			// Coord is occupied, penalize option
			svc.mux.Lock()
			(*weightedOptions)[direction] -= OccupiedSquarePenalty
			svc.mux.Unlock()
		}
	}

}

// adjustWeightsForCollisionCourse a collision course is when an enemy snake is heading straight and its next coord
// is one that our snake may move to. In this case, that coord should be penalized, but not forbidden since
// it is not guaranteed the enemy will continue straight.
func (svc *BananatronV1Svc) adjustWeightsForCollisionCourse(wg *sync.WaitGroup, weightedOptions *map[model.Direction]float64, selfSnake model.Snake, board model.Board) {
	defer wg.Done()

	var nextOccupiedCoords []model.Coord
	for _, snake := range board.Snakes {
		// Exclude self
		if snake.ID == selfSnake.ID {
			continue
		}
		snakeTravelDirection := snake.FindSnakeTravelDirection()
		nextOccupiedCoord := snake.Head.GetSquareInDirection(snakeTravelDirection)
		if nextOccupiedCoord != nil {
			nextOccupiedCoords = append(nextOccupiedCoords, *nextOccupiedCoord)
		}
	}

	for _, direction := range model.AllDirections {
		targetSquare := selfSnake.Head.GetSquareInDirection(direction)
		if util.Contains(nextOccupiedCoords, *targetSquare) {
			// Coord is a collision course coord, penalize option
			svc.mux.Lock()
			(*weightedOptions)[direction] -= CollisionCoursePenalty
			svc.mux.Unlock()
		}
	}
}
