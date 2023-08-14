package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGame_CreateSelectedBot(t *testing.T) {
	t.Run("banana_bot", func(t *testing.T) {
		bot := createSelectedBotService(bananaBotV1)
		assert.NotNil(t, bot)
	})

	t.Run("No bot found", func(t *testing.T) {
		bot := createSelectedBotService("not_a_bot")
		assert.Nil(t, bot)
	})
}

func TestGame_CalculateMove(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		gameController := NewGameController(nil, bananaBotV1)
		gameController.CalculateMove()
	})
}
