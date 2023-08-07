package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGame_CreateSelectedBot(t *testing.T) {
	t.Run("banana_bot_v1", func(t *testing.T) {
		bot := createSelectedBot("banana_bot_v1")
		assert.NotNil(t, bot)
	})

	t.Run("No bot found", func(t *testing.T) {
		bot := createSelectedBot("not_a_bot")
		assert.Nil(t, bot)
	})
}
