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
	// TODO: Implement, for now return a random move
	return &model.SnakeAction{
		Move:  model.Down,
		Shout: "Banana!",
	}, nil
}
