package services

import (
	"context"

	"BattlesnakeReptarium/internal/model"
)

type BananaBotV1Service interface {
	GetMove(ctx context.Context, game model.Game, turn int, board model.Board, self model.Snake) (model.SnakeAction, error)
}

type BananaBotV1Svc struct{}

func NewBananaBotV1Svc() *BananaBotV1Svc {
	return &BananaBotV1Svc{}
}

func (svc *BananaBotV1Svc) GetMove(ctx context.Context, game model.Game, turn int, board model.Board, self model.Snake) (model.SnakeAction, error) {
	panic("implement me")
	return model.SnakeAction{}, nil
}
