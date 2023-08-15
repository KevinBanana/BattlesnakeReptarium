package controllers

import "BattlesnakeReptarium/internal/model"

var (
	game      = model.Game{}
	turn      = 0
	board     = model.Board{}
	selfSnake = model.Snake{}

	defaultRequest = model.RequestBody{
		Game:      model.Game{},
		Turn:      1,
		Board:     model.Board{},
		SelfSnake: model.Snake{},
	}
)
