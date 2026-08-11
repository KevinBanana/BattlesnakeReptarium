package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGame_CreateSelectedBot(t *testing.T) {
	t.Run("banana_bot", func(t *testing.T) {
		bot := newBotService(bananaBotV1)
		assert.NotNil(t, bot)
	})

	t.Run("No bot found", func(t *testing.T) {
		bot := newBotService("not_a_bot")
		assert.Nil(t, bot)
	})
}
