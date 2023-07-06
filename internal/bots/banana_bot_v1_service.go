package bots

import (
	"context"

	"BattlesnakeReptarium/internal/model"
)

type BananaBotV1Service interface {
	StartGame(ctx context.Context, game model.Game, board model.Board, self model.Snake) error
	GetMove(ctx context.Context, game model.Game, turn int, board model.Board, self model.Snake) (model.SnakeAction, error)
	EndGame(ctx context.Context, game model.Game, turn int, board model.Board, self model.Snake) error
}
