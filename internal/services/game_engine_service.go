package services

import (
	"context"
	"fmt"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/repo"
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
		return fmt.Errorf("StartGame::failed to create game in DB: %w", err)
	}
	return nil
}

func (svc *GameEngineSvc) EndGame(ctx context.Context, game model.Game, board model.Board, self model.Snake) error {
	game.IsFinished = true
	game.IsWin = board.IsSnakeOnBoard(self)

	if err := svc.db.UpdateGame(ctx, game); err != nil {
		return fmt.Errorf("EndGame::failed to update game in DB: %w", err)
	}
	return nil
}
