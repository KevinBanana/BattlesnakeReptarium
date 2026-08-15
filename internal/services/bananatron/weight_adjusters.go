package bananatron

import (
	"context"
	"log/slog"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/util"
)

const (
	OccupiedSquarePenalty            = -9999
	EnemyPotentialMovePenalty        = -4
	CollisionCoursePenalty           = -10
	CorneredSnakeEscapeSquarePenalty = -15
)

type WeightAdjuster interface {
	AdjustWeight(ctx context.Context, selfSnake model.Snake, board model.Board) map[model.Direction]float64
}

func newDeltas() map[model.Direction]float64 {
	deltas := make(map[model.Direction]float64, len(model.AllDirections))
	for _, direction := range model.AllDirections {
		deltas[direction] = 0
	}
	return deltas
}

// OccupiedSquaresAdjuster If a square is occupied, severely penalize that square
type OccupiedSquaresAdjuster struct{}

func (a *OccupiedSquaresAdjuster) AdjustWeight(ctx context.Context, selfSnake model.Snake, board model.Board) map[model.Direction]float64 {
	deltas := newDeltas()

	for _, direction := range model.AllDirections {
		targetSquare := selfSnake.Head.GetCoordInDirection(direction)
		if !board.IsCoordClear(*targetSquare) {
			// Coord is occupied, penalize option
			deltas[direction] += OccupiedSquarePenalty
		}
	}

	return deltas
}

// CollisionCourseAdjuster a collision course is when an enemy snake is heading straight and its next coord
// is one that our snake may move to. In this case, that coord should be penalized, but not forbidden since
// it is not guaranteed the enemy will continue straight.
type CollisionCourseAdjuster struct{}

func (a *CollisionCourseAdjuster) AdjustWeight(ctx context.Context, selfSnake model.Snake, board model.Board) map[model.Direction]float64 {
	deltas := newDeltas()

	var nextOccupiedCoords []model.Coord
	for _, snake := range board.Snakes {
		// Exclude self
		if snake.ID == selfSnake.ID {
			continue
		}
		// A snake that has not moved yet has nothing to extrapolate from.
		snakeTravelDirection, ok := snake.TravelDirection()
		if !ok {
			continue
		}

		nextOccupiedCoord := snake.Head.GetCoordInDirection(snakeTravelDirection)
		if nextOccupiedCoord != nil {
			nextOccupiedCoords = append(nextOccupiedCoords, *nextOccupiedCoord)
		}
	}

	for _, direction := range model.AllDirections {
		targetSquare := selfSnake.Head.GetCoordInDirection(direction)
		if util.Contains(nextOccupiedCoords, *targetSquare) {
			// Coord is a collision course coord, penalize option
			deltas[direction] += CollisionCoursePenalty
		}
	}

	return deltas
}

// PotentialEnemyMoveAdjuster If an enemy snake could move onto a square, penalize that square
type PotentialEnemyMoveAdjuster struct{}

func (a *PotentialEnemyMoveAdjuster) AdjustWeight(ctx context.Context, selfSnake model.Snake, board model.Board) map[model.Direction]float64 {
	deltas := newDeltas()

	for _, direction := range model.AllDirections {
		potentialMoveSquare := selfSnake.Head.GetCoordInDirection(direction)
		for _, enemySnake := range board.Snakes {
			// Exclude self
			if enemySnake.ID == selfSnake.ID {
				continue
			}
			for _, snakeDirection := range model.AllDirections {
				potentialEnemyMoveSquare := enemySnake.Head.GetCoordInDirection(snakeDirection)
				if potentialEnemyMoveSquare != nil && *potentialEnemyMoveSquare == *potentialMoveSquare {
					// Coord is a potential enemy move coord, penalize option
					deltas[direction] += EnemyPotentialMovePenalty
				}
			}
		}
	}

	return deltas
}

// CavernSizeAdjuster uses flood fill to determine how many open squares are reachable from each direction
type CavernSizeAdjuster struct{}

func (a *CavernSizeAdjuster) AdjustWeight(ctx context.Context, selfSnake model.Snake, board model.Board) map[model.Direction]float64 {
	deltas := newDeltas()

	for _, direction := range model.AllDirections {
		targetSquare := selfSnake.Head.GetCoordInDirection(direction)
		floodFillCoords := board.FloodFill(*targetSquare)
		if len(floodFillCoords) == 0 {
			continue
		}

		// Divide the total squares by the number of players in the cavern since they will each consume a portion
		snakesInCavern := board.SnakesInCavern(floodFillCoords)
		if len(snakesInCavern) == 0 {
			slog.ErrorContext(ctx, "snakes in cavern is 0, should have found at least self snake",
				"direction", direction,
				"target", targetSquare,
			)
			continue
		}

		cavernScore := float64(len(floodFillCoords)) / float64(len(snakesInCavern))
		deltas[direction] += cavernScore
	}

	return deltas
}

// AvoidingCorneredSnakesAdjuster When an enemy snake has only one valid move, avoid moving to that coord
type AvoidingCorneredSnakesAdjuster struct{}

func (a *AvoidingCorneredSnakesAdjuster) AdjustWeight(ctx context.Context, selfSnake model.Snake, board model.Board) map[model.Direction]float64 {
	deltas := newDeltas()

	var nextOccupiedCoords []model.Coord // Tracks escape coords for cornered snakes
	for _, snake := range board.Snakes {
		// Exclude self
		if snake.ID == selfSnake.ID {
			continue
		}
		var clearOptions []model.Coord
		for _, direction := range model.AllDirections {
			targetSquare := snake.Head.GetCoordInDirection(direction)
			if board.IsCoordClear(*targetSquare) {
				clearOptions = append(clearOptions, *targetSquare)
			}
		}
		if len(clearOptions) == 1 { // Snake only has one valid move, it is cornered
			nextOccupiedCoords = append(nextOccupiedCoords, clearOptions[0])
		}
	}

	for _, direction := range model.AllDirections {
		targetSquare := selfSnake.Head.GetCoordInDirection(direction)
		if util.Contains(nextOccupiedCoords, *targetSquare) {
			// Coord is an escape coord for cornered snake, penalize option
			deltas[direction] += CorneredSnakeEscapeSquarePenalty
		}
	}

	return deltas
}
