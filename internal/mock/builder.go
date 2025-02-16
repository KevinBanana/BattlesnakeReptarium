package mock

import "BattlesnakeReptarium/internal/model"

type BoardOption func(*model.Board)

func WithBoardHeight(height int) BoardOption {
	return func(b *model.Board) {
		b.Height = height
	}
}

func WithBoardWidth(width int) BoardOption {
	return func(b *model.Board) {
		b.Width = width
	}
}

func WithBoardSnakes(snakes []model.Snake) BoardOption {
	return func(b *model.Board) {
		b.Snakes = snakes
	}
}

func NewBoard(opts ...BoardOption) model.Board {
	board := model.Board{
		Height:  11,
		Width:   11,
		Food:    nil,
		Hazards: nil,
		Snakes:  nil,
	}
	for _, opt := range opts {
		opt(&board)
	}
	return board
}
