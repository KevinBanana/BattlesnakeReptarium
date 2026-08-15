package services

import (
	"context"
	"slices"

	"BattlesnakeReptarium/internal/model"
)

type Bot interface {
	Customizations() model.SnakeCustomizations

	// Gamemodes lists the model.Gamemode values this bot is built to play
	Gamemodes() []string

	CalculateMove(ctx context.Context, game model.Game, turn int, board model.Board, self model.Snake) (*model.SnakeAction, error)
}

// PlaysGamemode reports whether bot is built for the gamemode of the game it
// has been entered into.
func PlaysGamemode(bot Bot, gamemode string) bool {
	modes := bot.Gamemodes()
	return slices.Contains(modes, gamemode)
}
