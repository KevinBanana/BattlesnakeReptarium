package services

import "BattlesnakeReptarium/internal/model"

var (
	snakeHeadingDown  = model.Snake{Head: model.Coord{X: 5, Y: 5}, Body: []model.Coord{{X: 5, Y: 6}, {X: 5, Y: 7}}}
	snakeHeadingUp    = model.Snake{Head: model.Coord{X: 5, Y: 5}, Body: []model.Coord{{X: 5, Y: 4}, {X: 5, Y: 3}}}
	snakeHeadingLeft  = model.Snake{Head: model.Coord{X: 5, Y: 5}, Body: []model.Coord{{X: 6, Y: 5}, {X: 7, Y: 5}}}
	snakeHeadingRight = model.Snake{Head: model.Coord{X: 5, Y: 5}, Body: []model.Coord{{X: 4, Y: 5}, {X: 3, Y: 5}}}
)
