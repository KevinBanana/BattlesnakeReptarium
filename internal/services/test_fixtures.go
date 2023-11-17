package services

import "BattlesnakeReptarium/internal/model"

var (
	snakeHeadingTox1y10 = model.Snake{
		ID:             "test-snake-1",
		Name:           "test-snake-1",
		Health:         100,
		Body:           []model.Coord{{X: 1, Y: 9}, {X: 1, Y: 8}, {X: 1, Y: 7}},
		Head:           model.Coord{X: 1, Y: 9},
		Length:         0,
		Latency:        "",
		Shout:          "",
		Squad:          "",
		Customizations: model.SnakeCustomizations{},
	}
)
