package services

import (
	"context"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/repo"
)

type GameEngineService interface {
	StartGame(ctx context.Context, game model.Game, board model.Board, self model.Snake) error
	EndGame(ctx context.Context, game model.Game, turn int, board model.Board, self model.Snake) error
}

type GameEngineSvc struct {
	db repo.DB
}

func NewGameEngineSvc(db repo.DB) *GameEngineSvc {
	return &GameEngineSvc{
		db: db,
	}
}

func (svc *GameEngineSvc) StartGame(ctx context.Context, game model.Game, board model.Board, self model.Snake) error {
	panic("implement me")
	return nil
}

func (svc *GameEngineSvc) EndGame(ctx context.Context, game model.Game, turn int, board model.Board, self model.Snake) error {
	panic("implement me")
	return nil
}
