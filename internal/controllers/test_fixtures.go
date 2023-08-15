package controllers

import "BattlesnakeReptarium/internal/model"

var (
	game      = model.Game{}
	turn      = 0
	board     = model.Board{}
	selfSnake = model.Snake{}

	defaultMoveRequest = model.MoveRequestBody{
		Game:      model.Game{},
		Turn:      1,
		Board:     model.Board{},
		SelfSnake: model.Snake{},
	}
)
