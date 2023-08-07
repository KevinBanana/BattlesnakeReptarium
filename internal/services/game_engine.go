package services

import (
	"context"

	"BattlesnakeReptarium/internal/model"
)

type GameEngineService interface {
	StartGame(ctx context.Context, game model.Game, board model.Board, self model.Snake) error
	EndGame(ctx context.Context, game model.Game, turn int, board model.Board, self model.Snake) error
}
