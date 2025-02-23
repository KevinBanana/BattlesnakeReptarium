package bananatron_service

import (
	"context"
	"fmt"
	"math"
	"sync"

	"BattlesnakeReptarium/internal/model"
)

type BananatronV1Svc struct {
	mux sync.RWMutex
}

func NewBananatronSvc() *BananatronV1Svc {
	return &BananatronV1Svc{}
}

func (svc *BananatronV1Svc) CalculateMove(ctx context.Context, game model.Game, turn int, board model.Board, selfSnake model.Snake) (*model.SnakeAction, error) {
	weightedOptions := map[model.Direction]float64{}
	for i, direction := range model.AllDirections {
		weightedOptions[direction] = float64(i) // Give default weight of i so that snake will prefer clockwise movement
	}

	wg := new(sync.WaitGroup)

	adjusters := []WeightAdjuster{
		&OccupiedSquaresAdjuster{},
		&CollisionCourseAdjuster{},
		&CavernSizeAdjuster{},
		&AvoidingCorneredSnakesAdjuster{},
		&PotentialEnemyMoveAdjuster{},
		&VoronoiControlAdjuster{},
	}

	wg.Add(len(adjusters))
	for _, adjuster := range adjusters {
		go adjuster.AdjustWeight(wg, &weightedOptions, selfSnake, board, &svc.mux)
	}
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
