package services

import (
	"context"

	"BattlesnakeReptarium/internal/model"
)

type BananaBotV1Svc struct{}

func NewBananaBotV1Svc() *BananaBotV1Svc {
	return &BananaBotV1Svc{}
}

func (svc *BananaBotV1Svc) CalculateMove(ctx context.Context, game model.Game, turn int, board model.Board, selfSnake model.Snake) (*model.SnakeAction, error) {
	// TODO: Implement, for now return a random move
	return &model.SnakeAction{
		Move:  model.Down,
		Shout: "Banana!",
	}, nil
}
