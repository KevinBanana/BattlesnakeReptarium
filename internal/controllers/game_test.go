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
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestGame_Health(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			testGameController := NewGameController(&repo.Database{}, nil)

			testGameController.Health(b.ctx)
			assert.Equal(t, http.StatusOK, b.ctx.Writer.Status())
		})
	})
}

func TestGame_CalculateMove(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			moveRequestSetup(b.ctx)

			b.mockBot.EXPECT().CalculateMove(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&model.SnakeAction{}, nil).Times(1)
			testGameController := NewGameController(&repo.Database{}, b.mockBot)

			testGameController.CalculateMove(b.ctx)
			assert.Equal(t, http.StatusOK, b.ctx.Writer.Status())
		})
	})

	t.Run("Bot not set", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			moveRequestSetup(b.ctx)

			testGameController := NewGameController(&repo.Database{}, nil)

			testGameController.CalculateMove(b.ctx)
			assert.Equal(t, http.StatusInternalServerError, b.ctx.Writer.Status())
		})
	})

	t.Run("Bad request", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			jsonDefaultMoveRequest, _ := json.Marshal("bad request")
			b.ctx.Request, _ = http.NewRequest("POST", "/move", bytes.NewBuffer(jsonDefaultMoveRequest))

			b.mockBot.EXPECT().CalculateMove(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&model.SnakeAction{}, nil).Times(0)
			testGameController := NewGameController(&repo.Database{}, b.mockBot)

			testGameController.CalculateMove(b.ctx)
			assert.Equal(t, http.StatusBadRequest, b.ctx.Writer.Status())
		})
	})

	t.Run("Internal service error", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			moveRequestSetup(b.ctx)

			b.mockBot.EXPECT().CalculateMove(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				nil, errors.New("error")).Times(1)
			testGameController := NewGameController(&repo.Database{}, b.mockBot)

			testGameController.CalculateMove(b.ctx)
			assert.Equal(t, http.StatusInternalServerError, b.ctx.Writer.Status())
		})
	})
}

func moveRequestSetup(ctx *gin.Context) {
	jsonDefaultMoveRequest, _ := json.Marshal(defaultMoveRequest)
	ctx.Request, _ = http.NewRequest("POST", "/move", bytes.NewBuffer(jsonDefaultMoveRequest))
}

func withGameSetup(t gomock.TestReporter, testFunc func(testBundle gameTestBundle)) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	mockCtrl := gomock.NewController(t)
	mockBot := services.NewMockBot(mockCtrl)

	testFunc(gameTestBundle{
		ctx:     ctx,
		mockBot: mockBot,
	})
}

type gameTestBundle struct {
	ctx     *gin.Context
	mockBot *services.MockBot
}
