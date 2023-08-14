package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/repo"

	"github.com/gin-gonic/gin"
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
		testGameController := NewGameController(&repo.Database{}, bananaBotV1)
		moveRequest := model.MoveRequestBody{
			Game:      model.Game{},
			Turn:      5,
			Board:     model.Board{},
			SelfSnake: model.Snake{},
		}
		jsonValue, _ := json.Marshal(moveRequest)
		req, _ := http.NewRequest("POST", "/move", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		ctx.Request = req

		testGameController.CalculateMove(ctx)
	})
}
