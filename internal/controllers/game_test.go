package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"BattlesnakeReptarium/internal/model"
	"BattlesnakeReptarium/internal/services"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
)

func TestGame_Health(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			testGameController := NewGameController(nil, nil)

			testGameController.Health(b.w, httptest.NewRequest(http.MethodGet, "/", nil))
			assert.Equal(t, http.StatusOK, b.w.Code)
		})
	})
}

func TestGame_StartGame(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			b.mockGameEngine.EXPECT().StartGame(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

			testGameController := NewGameController(b.mockBot, b.mockGameEngine)
			testGameController.StartGame(b.w, jsonRequest(http.MethodPost, "/start", defaultRequest))
			assert.Equal(t, http.StatusOK, b.w.Code)
		})
	})

	t.Run("Bad request", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			testGameController := NewGameController(b.mockBot, b.mockGameEngine)
			testGameController.StartGame(b.w, jsonRequest(http.MethodPost, "/start", "bad request"))
			assert.Equal(t, http.StatusBadRequest, b.w.Code)
		})
	})

	t.Run("Internal server error", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			b.mockGameEngine.EXPECT().StartGame(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("error")).Times(1)

			testGameController := NewGameController(b.mockBot, b.mockGameEngine)
			testGameController.StartGame(b.w, jsonRequest(http.MethodPost, "/start", defaultRequest))
			assert.Equal(t, http.StatusInternalServerError, b.w.Code)
		})
	})
}

func TestGame_CalculateMove(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			b.mockBot.EXPECT().CalculateMove(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&model.SnakeAction{}, nil).Times(1)
			testGameController := NewGameController(b.mockBot, b.mockGameEngine)

			testGameController.CalculateMove(b.w, jsonRequest(http.MethodPost, "/move", defaultRequest))
			assert.Equal(t, http.StatusOK, b.w.Code)
		})
	})

	t.Run("Bot not set", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			testGameController := NewGameController(nil, nil)

			testGameController.CalculateMove(b.w, jsonRequest(http.MethodPost, "/move", defaultRequest))
			assert.Equal(t, http.StatusInternalServerError, b.w.Code)
		})
	})

	t.Run("Bad request", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			b.mockBot.EXPECT().CalculateMove(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&model.SnakeAction{}, nil).Times(0)
			testGameController := NewGameController(b.mockBot, b.mockGameEngine)

			testGameController.CalculateMove(b.w, jsonRequest(http.MethodPost, "/move", "bad request"))
			assert.Equal(t, http.StatusBadRequest, b.w.Code)
		})
	})

	t.Run("Internal server error", func(t *testing.T) {
		withGameSetup(t, func(b gameTestBundle) {
			b.mockBot.EXPECT().CalculateMove(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				nil, errors.New("error")).Times(1)
			testGameController := NewGameController(b.mockBot, b.mockGameEngine)

			testGameController.CalculateMove(b.w, jsonRequest(http.MethodPost, "/move", defaultRequest))
			assert.Equal(t, http.StatusInternalServerError, b.w.Code)
		})
	})
}

func jsonRequest(method, path string, body any) *http.Request {
	buf, _ := json.Marshal(body)
	return httptest.NewRequest(method, path, bytes.NewReader(buf))
}

func withGameSetup(t gomock.TestReporter, testFunc func(testBundle gameTestBundle)) {
	mockCtrl := gomock.NewController(t)

	testFunc(gameTestBundle{
		w:              httptest.NewRecorder(),
		mockBot:        services.NewMockBot(mockCtrl),
		mockGameEngine: services.NewMockGameEngineService(mockCtrl),
	})
}

type gameTestBundle struct {
	w              *httptest.ResponseRecorder
	mockBot        *services.MockBot
	mockGameEngine *services.MockGameEngineService
}
