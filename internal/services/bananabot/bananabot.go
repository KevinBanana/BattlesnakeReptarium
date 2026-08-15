package bananabot

import (
	"context"

	"BattlesnakeReptarium/internal/model"
)

type Service struct{}

func New() *Service {
	return &Service{}
}

func (svc *Service) Customizations() model.SnakeCustomizations {
	return model.SnakeCustomizations{
		Head:  "bendr",
		Tail:  "ion",
		Color: "#e8f008",
	}
}

func (svc *Service) Gamemodes() []string {
	return nil
}

func (svc *Service) CalculateMove(ctx context.Context, game model.Game, turn int, board model.Board, selfSnake model.Snake) (*model.SnakeAction, error) {
	// TODO: Implement, for now return a random move
	return &model.SnakeAction{
		Move:  model.DOWN,
		Shout: "Banana!",
	}, nil
}
