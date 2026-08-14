package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/repo"
)

type GameEngineService interface {
	StartGame(ctx context.Context, game model.Game, board model.Board, self model.Snake) error
	EndGame(ctx context.Context, game model.Game, board model.Board, self model.Snake) (model.Game, error)
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

func (svc *GameEngineSvc) EndGame(ctx context.Context, game model.Game, board model.Board, self model.Snake) (model.Game, error) {
	game.IsFinished = true
	game.IsWin = board.IsSnakeOnBoard(self)

	// A missing game is expected when the process restarted mid-game: the
	// in-memory store is gone, but the request body still tells us the outcome,
	if err := svc.db.UpdateGame(ctx, game); err != nil {
		if errors.Is(err, repo.ErrGameNotFound) {
			slog.WarnContext(ctx, "ending a game we have no record of")
			return game, nil
		}
		return game, fmt.Errorf("EndGame::failed to update game in DB: %w", err)
	}
	return game, nil
}
