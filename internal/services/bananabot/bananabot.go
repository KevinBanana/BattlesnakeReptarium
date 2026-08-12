package bananabot

import (
	"context"

	"BattlesnakeReptarium/internal/model"
)

type Service struct{}

func New() *Service {
	return &Service{}
}

func (svc *Service) CalculateMove(ctx context.Context, game model.Game, turn int, board model.Board, selfSnake model.Snake) (*model.SnakeAction, error) {
	// TODO: Implement, for now return a random move
	return &model.SnakeAction{
		Move:  model.DOWN,
		Shout: "Banana!",
	}, nil
}
