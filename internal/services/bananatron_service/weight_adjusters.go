package bananatron_service

import (
	"sync"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/util"

	log "github.com/sirupsen/logrus"
)

const (
	OccupiedSquarePenalty            = -9999
	EnemyPotentialMovePenalty        = -4
	CollisionCoursePenalty           = -10
	CorneredSnakeEscapeSquarePenalty = -15
)

type WeightAdjuster interface {
	AdjustWeight(wg *sync.WaitGroup, options *map[model.Direction]float64, selfSnake model.Snake, board model.Board, mux *sync.RWMutex)
}

// OccupiedSquaresAdjuster If a square is occupied, severely penalize that square
type OccupiedSquaresAdjuster struct{}

func (a *OccupiedSquaresAdjuster) AdjustWeight(wg *sync.WaitGroup, weightedOptions *map[model.Direction]float64, selfSnake model.Snake, board model.Board, mux *sync.RWMutex) {
	defer wg.Done()

	for _, direction := range model.AllDirections {
		targetSquare := selfSnake.Head.GetSquareInDirection(direction)
		if !board.IsCoordClear(*targetSquare) {
			// Coord is occupied, penalize option
			mux.Lock()
			(*weightedOptions)[direction] += OccupiedSquarePenalty
			mux.Unlock()
		}
	}
}

// CollisionCourseAdjuster a collision course is when an enemy snake is heading straight and its next coord
// is one that our snake may move to. In this case, that coord should be penalized, but not forbidden since
// it is not guaranteed the enemy will continue straight.
type CollisionCourseAdjuster struct{}

func (a *CollisionCourseAdjuster) AdjustWeight(wg *sync.WaitGroup, weightedOptions *map[model.Direction]float64, selfSnake model.Snake, board model.Board, mux *sync.RWMutex) {
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
			mux.Lock()
			(*weightedOptions)[direction] += CollisionCoursePenalty
			mux.Unlock()
		}
	}
}

// PotentialEnemyMoveAdjuster If an enemy snake could move onto a square, penalize that square
type PotentialEnemyMoveAdjuster struct{}

func (a *PotentialEnemyMoveAdjuster) AdjustWeight(wg *sync.WaitGroup, weightedOptions *map[model.Direction]float64, selfSnake model.Snake, board model.Board, mux *sync.RWMutex) {
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
					mux.Lock()
					(*weightedOptions)[direction] += EnemyPotentialMovePenalty
					mux.Unlock()
				}
			}
		}
	}
}

// CavernSizeAdjuster uses flood fill to determine how many open squares are reachable from each direction
type CavernSizeAdjuster struct{}

func (a *CavernSizeAdjuster) AdjustWeight(wg *sync.WaitGroup, weightedOptions *map[model.Direction]float64, selfSnake model.Snake, board model.Board, mux *sync.RWMutex) {
	defer wg.Done()

	for _, direction := range model.AllDirections {
		targetSquare := selfSnake.Head.GetSquareInDirection(direction)
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
		mux.Lock()
		(*weightedOptions)[direction] += cavernScore
		mux.Unlock()
	}
}

// AvoidingCorneredSnakesAdjuster When an enemy snake has only one valid move, avoid moving to that coord
type AvoidingCorneredSnakesAdjuster struct{}

func (a *AvoidingCorneredSnakesAdjuster) AdjustWeight(wg *sync.WaitGroup, weightedOptions *map[model.Direction]float64, selfSnake model.Snake, board model.Board, mux *sync.RWMutex) {
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
			mux.Lock()
			(*weightedOptions)[direction] += CorneredSnakeEscapeSquarePenalty
			mux.Unlock()
		}
	}
}
