package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/services"
	"go.uber.org/mock/gomock"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestGame_Health(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			testGameController := NewGameController(nil, nil)

			testGameController.Health(b.ctx)
			assert.Equal(t, http.StatusOK, b.ctx.Writer.Status())
		})
	})
}

func TestGame_StartGame(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			startRequestSetup(b.ctx)
			b.mockGameEngine.EXPECT().StartGame(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

			testGameController := NewGameController(b.mockBot, b.mockGameEngine)
			testGameController.StartGame(b.ctx)
			assert.Equal(t, http.StatusOK, b.ctx.Writer.Status())
		})
	})

	t.Run("Bad request", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			jsonDefaultMoveRequest, _ := json.Marshal("bad request")
			b.ctx.Request, _ = http.NewRequest("POST", "/start", bytes.NewBuffer(jsonDefaultMoveRequest))

			testGameController := NewGameController(b.mockBot, b.mockGameEngine)
			testGameController.StartGame(b.ctx)
			assert.Equal(t, http.StatusBadRequest, b.ctx.Writer.Status())
		})
	})

	t.Run("Internal server error", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			startRequestSetup(b.ctx)
			b.mockGameEngine.EXPECT().StartGame(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("error")).Times(1)

			testGameController := NewGameController(b.mockBot, b.mockGameEngine)
			testGameController.StartGame(b.ctx)
			assert.Equal(t, http.StatusInternalServerError, b.ctx.Writer.Status())
		})
	})
}

func TestGame_CalculateMove(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			moveRequestSetup(b.ctx)

			b.mockBot.EXPECT().CalculateMove(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&model.SnakeAction{}, nil).Times(1)
			testGameController := NewGameController(b.mockBot, b.mockGameEngine)

			testGameController.CalculateMove(b.ctx)
			assert.Equal(t, http.StatusOK, b.ctx.Writer.Status())
		})
	})

	t.Run("Bot not set", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			moveRequestSetup(b.ctx)

			testGameController := NewGameController(nil, nil)

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
			testGameController := NewGameController(b.mockBot, b.mockGameEngine)

			testGameController.CalculateMove(b.ctx)
			assert.Equal(t, http.StatusBadRequest, b.ctx.Writer.Status())
		})
	})

	t.Run("Internal server error", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			moveRequestSetup(b.ctx)

			b.mockBot.EXPECT().CalculateMove(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				nil, errors.New("error")).Times(1)
			testGameController := NewGameController(b.mockBot, b.mockGameEngine)

			testGameController.CalculateMove(b.ctx)
			assert.Equal(t, http.StatusInternalServerError, b.ctx.Writer.Status())
		})
	})
}

func moveRequestSetup(ctx *gin.Context) {
	jsonDefaultMoveRequest, _ := json.Marshal(defaultRequest)
	ctx.Request, _ = http.NewRequest("POST", "/move", bytes.NewBuffer(jsonDefaultMoveRequest))
}

func startRequestSetup(ctx *gin.Context) {
	jsonDefaultMoveRequest, _ := json.Marshal(defaultRequest)
	ctx.Request, _ = http.NewRequest("POST", "/start", bytes.NewBuffer(jsonDefaultMoveRequest))
}

func withGameSetup(t gomock.TestReporter, testFunc func(testBundle gameTestBundle)) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	mockCtrl := gomock.NewController(t)
	mockBot := services.NewMockBot(mockCtrl)
	mockGameEngine := services.NewMockGameEngineService(mockCtrl)

	testFunc(gameTestBundle{
		ctx:            ctx,
		mockBot:        mockBot,
		mockGameEngine: mockGameEngine,
	})
}

type gameTestBundle struct {
	ctx            *gin.Context
	mockBot        *services.MockBot
	mockGameEngine *services.MockGameEngineService
}
