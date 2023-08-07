package bots

import (
	"context"

	"BattlesnakeReptarium/internal/model"
)

type BananaBotV1Service interface {
	GetMove(ctx context.Context, game model.Game, turn int, board model.Board, self model.Snake) (model.SnakeAction, error)
}
