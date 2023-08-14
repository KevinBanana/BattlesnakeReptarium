package services

import (
	"context"

	"BattlesnakeReptarium/internal/model"
)

type Bot interface {
	CalculateMove(ctx context.Context, game model.Game, turn int, board model.Board, self model.Snake) (model.SnakeAction, error)
}
