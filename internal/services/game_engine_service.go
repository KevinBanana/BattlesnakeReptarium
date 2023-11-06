package services

import (
	"context"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/repo"
	"BattlesnakeReptarium/internal/util"

	"github.com/pkg/errors"
)

type GameEngineService interface {
	StartGame(ctx context.Context, game model.Game, board model.Board, self model.Snake) error
	EndGame(ctx context.Context, game model.Game, board model.Board, self model.Snake) error
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
	if err := svc.db.CreateGame(ctx, game); err != nil {
		return errors.Wrap(err, "StartGame::failed to create game in DB")
	}
	return nil
}

func (svc *GameEngineSvc) EndGame(ctx context.Context, game model.Game, board model.Board, self model.Snake) error {
	game.IsFinished = true
	game.IsWin = util.IsSnakeOnBoard(self, board)

	if err := svc.db.UpdateGame(ctx, game); err != nil {
		return errors.Wrap(err, "EndGame::failed to update game in DB")
	}
	return nil
}
