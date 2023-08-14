package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/repo"
	"BattlesnakeReptarium/internal/services"
	"go.uber.org/mock/gomock"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGame_CalculateMove(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
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
		mockCtrl := gomock.NewController(t)
		mockBot := services.NewMockBot(mockCtrl)
		mockBot.EXPECT().CalculateMove(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
			model.SnakeAction{}, nil).Times(1)
		testGameController := NewGameController(&repo.Database{}, mockBot)

		testGameController.CalculateMove(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// TODO clean up above test and add more
}
