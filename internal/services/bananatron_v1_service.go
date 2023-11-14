package services

import (
	"context"

	"BattlesnakeReptarium/internal/model"
)

type BananatronV1Svc struct{}

func NewBananatronV1Svc() *BananatronV1Svc {
	return &BananatronV1Svc{}
}

func (svc *BananatronV1Svc) CalculateMove(ctx context.Context, game model.Game, turn int, board model.Board, selfSnake model.Snake) (*model.SnakeAction, error) {
	// Very simple snake that just tries to avoid immediately suicidal moves
	options := []model.Direction{model.UP, model.LEFT, model.DOWN, model.RIGHT}
	for _, move := range options {
		targetSquare := selfSnake.Head.GetSquareInDirection(move)
		if board.IsCoordClear(*targetSquare) {
			return &model.SnakeAction{
				Move:  move,
				Shout: "Banana!",
			}, nil
		}
	}

	return &model.SnakeAction{
		Move:  model.DOWN,
		Shout: "Goodbye!",
	}, nil
}
