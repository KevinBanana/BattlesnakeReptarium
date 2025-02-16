package services

import (
	"context"
	"fmt"
	"math"
	"sync"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/util"

	log "github.com/sirupsen/logrus"
)

type BananatronV1Svc struct {
	mux sync.RWMutex
}

func NewBananatronSvc() *BananatronV1Svc {
	return &BananatronV1Svc{}
}

const (
	OccupiedSquarePenalty            = -9999
	EnemyPotentialMovePenalty        = -4
	CollisionCoursePenalty           = -10
	CorneredSnakeEscapeSquarePenalty = -15
)

func (svc *BananatronV1Svc) CalculateMove(ctx context.Context, game model.Game, turn int, board model.Board, selfSnake model.Snake) (*model.SnakeAction, error) {
	weightedOptions := map[model.Direction]float64{}
	for i, direction := range model.AllDirections {
		weightedOptions[direction] = float64(i) // Give default weight of i so that snake will prefer clockwise movement
	}

	wg := new(sync.WaitGroup)
	wg.Add(5)
	// TODO refactor this
	go svc.adjustWeightsForOccupiedSquares(wg, &weightedOptions, selfSnake.Head, board)
	go svc.adjustWeightsForCollisionCourse(wg, &weightedOptions, selfSnake, board)
	go svc.adjustWeightsForCavernSize(wg, &weightedOptions, selfSnake.Head, board)
	go svc.adjustWeightsForAvoidingCorneredSnakes(wg, &weightedOptions, selfSnake, board)
	go svc.adjustWeightsForPotentialEnemyMove(wg, &weightedOptions, selfSnake, board)

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

// adjustWeightsForAvoidingCorneredSnakes When an enemy snake has only one valid move, avoid moving to that coord
func (svc *BananatronV1Svc) adjustWeightsForAvoidingCorneredSnakes(wg *sync.WaitGroup, weightedOptions *map[model.Direction]float64, selfSnake model.Snake, board model.Board) {
	defer wg.Done()

	var nextOccupiedCoords []model.Coord // Tracks escape coords for cornered snakes
	for _, snake := range board.Snakes {
		// Exclude self
		if snake.ID == selfSnake.ID {
			continue
		}
		var clearOptions []model.Coord
		for _, direction := range model.AllDirections {
			targetSquare := snake.Head.GetSquareInDirection(direction)
			if board.IsCoordClear(*targetSquare) {
				clearOptions = append(clearOptions, *targetSquare)
			}
		}
		if len(clearOptions) == 1 { // Snake only has one valid move, it is cornered
			nextOccupiedCoords = append(nextOccupiedCoords, clearOptions[0])
		}
	}

	for _, direction := range model.AllDirections {
		targetSquare := selfSnake.Head.GetSquareInDirection(direction)
		if util.Contains(nextOccupiedCoords, *targetSquare) {
			// Coord is an escape coord for cornered snake, penalize option
			svc.applyScoreToSquare(weightedOptions, direction, CorneredSnakeEscapeSquarePenalty)
		}
	}
}

// adjustWeightsForCavernSize uses flood fill to determine how many open squares are reachable from each direction
func (svc *BananatronV1Svc) adjustWeightsForCavernSize(wg *sync.WaitGroup, weightedOptions *map[model.Direction]float64, selfHead model.Coord, board model.Board) {
	defer wg.Done()

	for _, direction := range model.AllDirections {
		targetSquare := selfHead.GetSquareInDirection(direction)
		floodFillCoords := board.DetermineFloodFillCoords(*targetSquare)
		if len(floodFillCoords) == 0 {
			continue
		}

		// Divide the total squares by the number of players in the cavern since they will each consume a portion
		snakesInCavern := board.FindAllSnakesInCavern(floodFillCoords)
		if len(snakesInCavern) == 0 {
			log.Error("Snakes in cavern is 0. Should have found at least self snake")
			continue
		}

		cavernScore := float64(len(floodFillCoords)) / float64(len(snakesInCavern))
		svc.applyScoreToSquare(weightedOptions, direction, cavernScore)
	}
}

func (svc *BananatronV1Svc) adjustWeightsForOccupiedSquares(wg *sync.WaitGroup, weightedOptions *map[model.Direction]float64, selfHead model.Coord, board model.Board) {
	defer wg.Done()

	for _, direction := range model.AllDirections {
		targetSquare := selfHead.GetSquareInDirection(direction)
		if !board.IsCoordClear(*targetSquare) {
			// Coord is occupied, penalize option
			svc.applyScoreToSquare(weightedOptions, direction, OccupiedSquarePenalty)
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
			svc.applyScoreToSquare(weightedOptions, direction, CollisionCoursePenalty)
		}
	}
}

// adjustWeightsForPotentialEnemyMove If an enemy snake could move onto a square that our snake could move to, penalize that square
func (svc *BananatronV1Svc) adjustWeightsForPotentialEnemyMove(wg *sync.WaitGroup, weightedOptions *map[model.Direction]float64, selfSnake model.Snake, board model.Board) {
	defer wg.Done()

	for _, direction := range model.AllDirections {
		potentialMoveSquare := selfSnake.Head.GetSquareInDirection(direction)
		for _, enemySnake := range board.Snakes {
			// Exclude self
			if enemySnake.ID == selfSnake.ID {
				continue
			}
			for _, snakeDirection := range model.AllDirections {
				potentialEnemyMoveSquare := enemySnake.Head.GetSquareInDirection(snakeDirection)
				if potentialEnemyMoveSquare != nil && *potentialEnemyMoveSquare == *potentialMoveSquare {
					// Coord is a potential enemy move coord, penalize option
					svc.applyScoreToSquare(weightedOptions, direction, EnemyPotentialMovePenalty)
				}
			}
		}
	}
}

func (svc *BananatronV1Svc) applyScoreToSquare(weightedOptions *map[model.Direction]float64, direction model.Direction, score float64) {
	svc.mux.Lock()
	(*weightedOptions)[direction] += score
	svc.mux.Unlock()
}
