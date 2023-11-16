package services

import (
	"context"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/util"
)

type BananatronV1Svc struct{}

func NewBananatronSvc() *BananatronV1Svc {
	return &BananatronV1Svc{}
}

func (svc *BananatronV1Svc) CalculateMove(ctx context.Context, game model.Game, turn int, board model.Board, selfSnake model.Snake) (*model.SnakeAction, error) {
	// Begin with all move options and filter out invalid moves
	options := []model.Direction{model.UP, model.LEFT, model.DOWN, model.RIGHT}

	options = excludeOccupiedCoordsFromOptions(options, selfSnake.Head, board)
	if len(options) == 1 {
		return &model.SnakeAction{
			Move:  options[0],
			Shout: "Only one option",
		}, nil
	}

	options = excludeCoordsAnySnakeIsHeadingFor(options, selfSnake, board)
	if len(options) == 1 {
		return &model.SnakeAction{
			Move:  options[0],
			Shout: "Only one option",
		}, nil
	}

	// Return the first remaining option
	if len(options) > 0 {
		return &model.SnakeAction{
			Move:  options[0],
			Shout: "Multiple options, but done picking",
		}, nil
	}

	// No valid options remain
	return &model.SnakeAction{
		Move:  model.DOWN,
		Shout: "Goodbye!",
	}, nil
}

func excludeOccupiedCoordsFromOptions(options []model.Direction, selfHead model.Coord, board model.Board) []model.Direction {
	var filteredOptions []model.Direction
	for _, move := range options {
		targetSquare := selfHead.GetSquareInDirection(move)
		if board.IsCoordClear(*targetSquare) {
			filteredOptions = append(filteredOptions, move)
		}
	}
	return filteredOptions
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
