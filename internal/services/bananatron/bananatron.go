package bananatron

import (
	"context"
	"fmt"
	"math"
	"sync"

	"BattlesnakeReptarium/internal/model"
)

type Service struct{}

func New() *Service {
	return &Service{}
}

func (svc *Service) Customizations() model.SnakeCustomizations {
	return model.SnakeCustomizations{
		Head:  "moto-helmet",
		Tail:  "nr-booster",
		Color: "#00e5ff",
	}
}

func (svc *Service) Gamemodes() []string {
	// collision and cornered-snake weighting assumes no
	// food and no starvation, so it only plays constrictor
	return []string{model.GamemodeConstrictor}
}

func (svc *Service) CalculateMove(ctx context.Context, game model.Game, turn int, board model.Board, selfSnake model.Snake) (*model.SnakeAction, error) {
	weightedOptions := map[model.Direction]float64{}
	for i, direction := range model.AllDirections {
		weightedOptions[direction] = float64(i) // Give default weight of i so that snake will prefer clockwise movement
	}

	adjusters := []WeightAdjuster{
		&OccupiedSquaresAdjuster{},
		&CollisionCourseAdjuster{},
		&CavernSizeAdjuster{},
		&AvoidingCorneredSnakesAdjuster{},
		&PotentialEnemyMoveAdjuster{},
	}

	results := make([]map[model.Direction]float64, len(adjusters))
	var wg sync.WaitGroup
	for i, adjuster := range adjusters {
		wg.Go(func() {
			results[i] = adjuster.AdjustWeight(ctx, selfSnake, board)
		})
	}
	wg.Wait()

	for _, deltas := range results {
		for direction, weight := range deltas {
			weightedOptions[direction] += weight
		}
	}

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
